package solauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT is the interactor for JWT.
type JWT struct {
	signingKey []byte
}

// NewJWT creates a new JWT interactor.
func NewJWT(signingKey []byte) *JWT {
	return &JWT{
		signingKey: signingKey,
	}
}

// Claims is the claims for the token.
type Claims struct {
	Wallet string `json:"wallet"`
	jwt.RegisteredClaims
}

// TokenResponse is the response for the token request.
type TokenResponse struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresIn int64  `json:"expires_in"`
}

// IssueToken issues a token for the user.
// This function generates a token for the user and returns it.
func (j *JWT) IssueTokens(walletAddr string) (TokenResponse, error) {
	accessID := uuid.New().String()
	accessExpAt := time.Now().Add(time.Hour * 1).Unix() // 1 hour

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Wallet: walletAddr,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessID,
			Audience:  jwt.ClaimStrings{"access"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Unix(accessExpAt, 0)),
		},
	})

	// Sign and get the complete encoded token as a string using the secret
	accessTokenString, err := accessToken.SignedString(j.signingKey)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("failed to sign token: %w", err)
	}

	// Refresh token
	refreshID := uuid.New().String()
	refreshExpAt := time.Now().Add(time.Hour * 24 * 7).Unix() // 7 days

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Wallet: walletAddr,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshID,
			Audience:  jwt.ClaimStrings{"refresh"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Unix(refreshExpAt, 0)),
		},
	})

	// Sign and get the complete encoded token as a string using the secret
	refreshTokenString, err := refreshToken.SignedString(j.signingKey)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return TokenResponse{
		Access:    accessTokenString,
		Refresh:   refreshTokenString,
		ExpiresIn: 3600, // 1 hour
	}, nil
}

// VerifyToken verifies the token.
// This function verifies the token and returns the claims.
func (j *JWT) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.signingKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken refreshes the token.
// This function refreshes the token and returns the new token.
func (j *JWT) RefreshToken(tokenString string) (TokenResponse, error) {
	if tokenString == "" {
		return TokenResponse{}, fmt.Errorf("token is empty")
	}

	claims, err := j.VerifyToken(tokenString)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("failed to verify token: %w", err)
	}

	if claims.Audience[0] != "refresh" {
		return TokenResponse{}, fmt.Errorf("the token is not a refresh token")
	}

	return j.IssueTokens(claims.Wallet)
}
