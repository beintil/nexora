package config

// AuthConfig — настройки авторизации (cookie, JWT TTL, Redis ключи, ссылка подтверждения).
// Все поля читаются из env, при отсутствии приложение не стартует.
// Поля *TTLSec задаются в секундах (целое число), например 9000.
type AuthConfig struct {
	RefreshTokenCookieName   string `env:"REFRESH_TOKEN_COOKIE_NAME" env-required:"true"`
	RefreshTokenCookieMaxAge int    `env:"REFRESH_TOKEN_COOKIE_MAX_AGE" env-required:"true"` // секунды
	RefreshTokenCookieSecure bool   `env:"REFRESH_TOKEN_COOKIE_SECURE"`                      // true в production по HTTPS
	AccessTokenTTLSec        int    `env:"ACCESS_TOKEN_TTL_SEC" env-required:"true"`         // время жизни access токена (сек)
	RefreshTokenTTLSec       int    `env:"REFRESH_TOKEN_TTL_SEC" env-required:"true"`        // время жизни refresh токена в Redis (сек)

	// Ссылка подтверждения (UUID в письме/SMS): базовый URL, TTL и лимит попыток.
	AuthLinkBaseURL string `env:"AUTH_LINK_BASE_URL" env-required:"true"` // например https://app.example.com
	AuthLinkTTLSec  int    `env:"AUTH_LINK_TTL_SEC" env-required:"true"`  // время жизни ссылки в Redis (сек)
	JWTSecret       string `env:"AUTH_JWT_SECRET" env-required:"true"`    // секрет для подписи JWT

	//OAuth OAuthConfig
}

// OAuthConfig — настройки OAuth (Google / Apple). Все поля из env.
type OAuthConfig struct {
	OAuthBackendBaseURL      string `env:"OAUTH_BACKEND_BASE_URL" env-required:"true"`       // базовый URL бэкенда для redirect_uri
	OAuthStateRedisKeyPrefix string `env:"OAUTH_STATE_REDIS_KEY_PREFIX" env-required:"true"` // префикс ключа state в Redis (CSRF)
	OAuthStateTTLSec         int    `env:"OAUTH_STATE_TTL_SEC" env-required:"true"`          // TTL state в секундах (рекомендуется 600)
	OAuthFrontendSuccessURL  string `env:"OAUTH_FRONTEND_SUCCESS_URL" env-required:"true"`   // куда редиректить после успешного OAuth
	OAuthFrontendErrorURL    string `env:"OAUTH_FRONTEND_ERROR_URL" env-required:"true"`     // куда редиректить при ошибке OAuth

	OAuthApple  OAuthAppleConfig  `env:"OAUTH_APPLE" env-required:"true"`
	OAuthGoogle OAuthGoogleConfig `env:"OAUTH_GOOGLE" env-required:"true"`
}

// OAuthAppleConfig — Apple Sign in. Все поля из env.
type OAuthAppleConfig struct {
	TeamID         string `env:"OAUTH_APPLE_TEAM_ID" env-required:"true"`       // Apple Developer Team ID
	KeyID          string `env:"OAUTH_APPLE_KEY_ID" env-required:"true"`        // Key ID из Apple (Sign in with Apple key)
	PrivateKeyPEM  string `env:"OAUTH_APPLE_PRIVATE_KEY_PEM"`                   // содержимое .p8 файла (можно в одну строку с \n)
	PrivateKeyPath string `env:"OAUTH_APPLE_PRIVATE_KEY_PATH"`                  // или путь к .p8 файлу (приоритет над PEM если задан)
	ClientID       string `env:"OAUTH_APPLE_CLIENT_ID" env-required:"true"`     // Service ID (например com.example.app)
	RedirectPath   string `env:"OAUTH_APPLE_REDIRECT_PATH" env-required:"true"` // например /v1/auth/apple/callback
	AuthURL        string `env:"OAUTH_APPLE_AUTH_URL" env-required:"true"`      // URL страницы авторизации Apple
	JWKSURL        string `env:"OAUTH_APPLE_JWKS_URL" env-required:"true"`      // URL ключей Apple
	Issuer         string `env:"OAUTH_APPLE_ISSUER" env-required:"true"`        // issuer для проверки id_token
}

// OAuthGoogleConfig — Google OAuth. Все поля из env.
type OAuthGoogleConfig struct {
	ClientSecret string `env:"OAUTH_GOOGLE_CLIENT_SECRET" env-required:"true"` // Client Secret из Google Cloud Console
	ClientID     string `env:"OAUTH_GOOGLE_CLIENT_ID" env-required:"true"`
	RedirectPath string `env:"OAUTH_GOOGLE_REDIRECT_PATH" env-required:"true"` // например /v1/auth/google/callback
	UserinfoURL  string `env:"OAUTH_GOOGLE_USERINFO_URL" env-required:"true"`  // URL userinfo
}
