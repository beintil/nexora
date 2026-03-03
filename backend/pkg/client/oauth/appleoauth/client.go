package appleoauth

import "context"

// Client — клиент Sign in with Apple: проверка id_token и построение URL авторизации.
type Client interface {
	AuthCodeURL(redirectURI, state string) string
	VerifyIDToken(ctx context.Context, idToken string) (email, sub string, err error)
}
