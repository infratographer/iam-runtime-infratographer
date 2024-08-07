package filetokensource

import (
	"errors"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.infratographer.com/iam-runtime-infratographer/internal/testauth"
)

func TestToken(t *testing.T) {
	authsrv := testauth.NewServer(t)
	t.Cleanup(authsrv.Stop)

	type fileToken struct {
		subject       string
		expectSubject string
		expectError   error
	}

	testCases := []struct {
		name        string
		tokenSource *TokenSource
		tokens      []fileToken
	}{
		{
			"no token",
			&TokenSource{
				path: "/tmp/token-not-found",
			},
			[]fileToken{
				{
					expectError: os.ErrNotExist,
				},
			},
		},
		{
			"valid token",
			&TokenSource{},
			[]fileToken{{
				subject:       "token1",
				expectSubject: "token1",
			}},
		},
		{
			"no reuse token",
			&TokenSource{},
			[]fileToken{
				{subject: "token1", expectSubject: "token1"},
				{subject: "token2", expectSubject: "token2"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for i, token := range tc.tokens {
				if tc.tokenSource.path == "" {
					file, err := os.CreateTemp(t.TempDir(), "token-*")

					require.NoError(t, err, "unexpected error creating temporary token file")
					require.NoError(t, file.Close(), "unexpected error closing temporary token file")

					tc.tokenSource.path = file.Name()

					defer func() {
						require.NoError(t, os.Remove(tc.tokenSource.path), "unexpected error removing temporary token file")
					}()
				}

				if !errors.Is(token.expectError, os.ErrNotExist) {
					var jwt string

					if token.subject != "" {
						jwt = authsrv.TSignSubject(t, token.subject)
					}

					err := os.WriteFile(tc.tokenSource.path, []byte(jwt), 0600)
					require.NoError(t, err, "unexpected error writing token")
				}

				retToken, err := tc.tokenSource.Token()

				if token.expectError != nil {
					assert.Nilf(t, retToken, "token %d: expected returned token to be nil", i)
					require.Errorf(t, err, "token %d: expected error set in handler", i)
					assert.ErrorIsf(t, err, token.expectError, "token %d: unexpected error set in handler", i)

					continue
				}

				assert.NoErrorf(t, err, "token %d: no error expected in handler", i)
				require.NotNilf(t, retToken, "token %d: expected token to be returned", i)

				subject := getSubjectf(t, retToken.AccessToken, "token %d:", i)

				assert.Equalf(t, token.expectSubject, subject, "token %d: unexpected access token returned", i)
			}
		})
	}
}

func getSubjectf(t *testing.T, token string, msgPrefix string, msgArgs ...any) string {
	t.Helper()

	jwtToken, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	require.NoErrorf(t, err, msgPrefix+" unexpected error parsing jwt token", msgArgs...)

	subject, err := jwtToken.Claims.GetSubject()
	require.NoErrorf(t, err, msgPrefix+" unexpected error getting subject", msgArgs...)

	return subject
}
