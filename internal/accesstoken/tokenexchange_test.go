package accesstoken

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type staticTokenSource struct {
	token *oauth2.Token
}

func (s *staticTokenSource) Token() (*oauth2.Token, error) {
	return s.token, nil
}

// TestTokenExchangeNoAuthHeaderWithoutCredentials verifies that no Authorization header
// is sent during an RFC 8693 token exchange when no client credentials are configured.
// Without AuthStyleInParams, the oauth2 package defaults to AuthStyleAutoDetect which
// probes with Basic auth first, injecting "Basic Og==" (base64 of ":") even when both
// ClientID and ClientSecret are empty.
func TestTokenExchangeNoAuthHeaderWithoutCredentials(t *testing.T) {
	t.Parallel()

	var (
		capturedAuthHeader string
		capturedForm       map[string][]string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuthHeader = r.Header.Get("Authorization")

		require.NoError(t, r.ParseForm())
		capturedForm = r.Form

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"exchanged-token","token_type":"Bearer","expires_in":3600}`))
	}))
	t.Cleanup(server.Close)

	src := &exchangeTokenSource{
		cfg: ExchangeConfig{
			GrantType: defaultGrantType,
			TokenType: defaultTokenType,
		},
		ctx: context.Background(),
		upstream: &staticTokenSource{
			token: &oauth2.Token{AccessToken: "upstream-token"},
		},
		exchangeConfig: oauth2.Config{
			Endpoint: oauth2.Endpoint{
				TokenURL:  server.URL + "/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
		},
	}

	_, err := src.Token()
	require.NoError(t, err)

	assert.Empty(t, capturedAuthHeader, "expected no Authorization header when no client credentials are configured")
	assert.NotContains(t, capturedForm, "client_id", "expected no client_id in request body")
	assert.NotContains(t, capturedForm, "client_secret", "expected no client_secret in request body")
}
