package solauth_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrymomot/solauth"
	"github.com/portto/solana-go-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	authSigningKey = []byte("secret")
	wallet, _      = types.AccountFromBase58("4JVyzx75j9s91TgwVqSPFN4pb2D8ACPNXUKKnNBvXuGukEzuFEg3sLqhPGwYe9RRbDnVoYHjz4bwQ5yUfyRZVGVU")
)

func TestRequestAuth(t *testing.T) {
	reqData := solauth.RequestAuthHandlePayload{
		PublicKey: wallet.PublicKey.ToBase58(),
	}
	jsonData, err := json.Marshal(reqData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/request", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	solauth.RequestAuth(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	response := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(t, err)
	require.NotEmpty(t, response["message"])
}

func TestVerifySignedMessage(t *testing.T) {
	message := "test message"
	signature := wallet.Sign([]byte(message))

	reqData := solauth.VerifySignedMessagePayload{
		Message:   message,
		Signature: base64.StdEncoding.EncodeToString(signature),
		PublicKey: wallet.PublicKey.ToBase58(),
	}
	jsonData, err := json.Marshal(reqData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/verify", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	solauth.VerifySignedMessage(solauth.NewJWT(authSigningKey))(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	response := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(t, err)
	require.NotEmpty(t, response["access_token"])
	require.NotEmpty(t, response["refresh_token"])
	require.NotEmpty(t, response["expires_in"])
}

func TestRefreshToken(t *testing.T) {
	tokens, err := solauth.NewJWT(authSigningKey).IssueTokens(wallet.PublicKey.ToBase58())
	require.NoError(t, err)

	reqData := solauth.RefreshTokenPayload{
		RefreshToken: tokens.Refresh,
	}
	jsonData, err := json.Marshal(reqData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	solauth.RefreshToken(solauth.NewJWT(authSigningKey))(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	response := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(t, err)
	require.NotEmpty(t, response["access_token"])
	require.NotEmpty(t, response["refresh_token"])
	require.NotEmpty(t, response["expires_in"])
}
