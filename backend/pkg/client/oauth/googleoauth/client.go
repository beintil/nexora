package googleoauth

import "context"

// Client — клиент OAuth 2.0 и userinfo Google. Реализации вызывают внешние API Google.
type Client interface {
	AuthCodeURL(state string) string
	ExchangeAndGetUserInfo(ctx context.Context, code string) (email, name, picture string, err error)
}
