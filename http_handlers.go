package solauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// helper to send response as a json data
func defaultResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RequestAuthHandlePayload is the payload for the request authentication.
type RequestAuthHandlePayload struct {
	// PublicKey is the public key of the sender.
	PublicKey string `json:"public_key"`
}

// RequestAuth is the handler for the request authentication.
// It gets the wallet address and returns message to sign.
// The message must be signed by the wallet and sent back to the server.
// The server will verify the signature and return the result.
func RequestAuth(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request
	var payload RequestAuthHandlePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		defaultResponse(w, http.StatusBadRequest, map[string]interface{}{
			"code":  http.StatusBadRequest,
			"error": err.Error(),
		})
		return
	}

	if payload.PublicKey == "" {
		defaultResponse(w, http.StatusBadRequest, map[string]interface{}{
			"code":  http.StatusBadRequest,
			"error": "wallet_address is required",
		})
		return
	}

	message := fmt.Sprintf(
		"Sign this message to login as %s. Request ID: %s",
		payload.PublicKey,
		middleware.GetReqID(r.Context()),
	)

	defaultResponse(w, http.StatusOK, map[string]interface{}{
		"message": message,
	})
}

// VerifySignedMessagePayload is the payload for the signed message verification.
type VerifySignedMessagePayload struct {
	// Message is the message that was signed.
	Message string `json:"message"`
	// Signature is the signature of the message.
	Signature string `json:"signature"`
	// PublicKey is the public key of the sender.
	PublicKey string `json:"public_key"`
}

// Validate validates the payload.
func (p *VerifySignedMessagePayload) Validate() error {
	if p.Message == "" {
		return fmt.Errorf("message is required")
	}
	if p.Signature == "" {
		return fmt.Errorf("signature is required")
	}
	if p.PublicKey == "" {
		return fmt.Errorf("public_key is required")
	}
	return nil
}

// VerifySignedMessage is the handler for the signed message verification.
// It verifies the signature of the message using the public key of the sender.
// It returns access token if the signature is valid, otherwise error.
func VerifySignedMessage(jwt interface {
	IssueTokens(walletAddr string) (TokenResponse, error)
},
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse JSON request
		var payload VerifySignedMessagePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			defaultResponse(w, http.StatusBadRequest, map[string]interface{}{
				"code":  http.StatusBadRequest,
				"error": err.Error(),
			})
			return
		}

		// Validate the payload
		if err := payload.Validate(); err != nil {
			defaultResponse(w, http.StatusBadRequest, map[string]interface{}{
				"code":  http.StatusBadRequest,
				"error": err.Error(),
			})
			return
		}

		// Verify the signature
		if err := VerifySignature(payload.Message, payload.Signature, payload.PublicKey); err != nil {
			defaultResponse(w, http.StatusBadRequest, map[string]interface{}{
				"code":  http.StatusBadRequest,
				"error": err.Error(),
			})
			return
		}

		// Issue tokens
		tokens, err := jwt.IssueTokens(payload.PublicKey)
		if err != nil {
			defaultResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"code":  http.StatusInternalServerError,
				"error": err.Error(),
			})
			return
		}

		defaultResponse(w, http.StatusOK, tokens)
	}
}

// RefreshTokenPayload is the payload for the refresh token.
type RefreshTokenPayload struct {
	// RefreshToken is the refresh token.
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken is the handler for the refresh token.
// It refreshes the access token.
func RefreshToken(jwt interface {
	RefreshToken(tokenString string) (TokenResponse, error)
},
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse JSON request
		var payload RefreshTokenPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			defaultResponse(w, http.StatusBadRequest, map[string]interface{}{
				"code":  http.StatusBadRequest,
				"error": err.Error(),
			})
			return
		}

		// Validate the payload
		if payload.RefreshToken == "" {
			defaultResponse(w, http.StatusBadRequest, map[string]interface{}{
				"code":  http.StatusBadRequest,
				"error": "refresh_token is required",
			})
			return
		}

		// Refresh the token
		tokens, err := jwt.RefreshToken(payload.RefreshToken)
		if err != nil {
			defaultResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"code":  http.StatusInternalServerError,
				"error": err.Error(),
			})
			return
		}

		defaultResponse(w, http.StatusOK, tokens)
	}
}
