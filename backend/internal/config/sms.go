package config

// SMSConfig — настройки SMS и email. Все поля из env.
type SMSConfig struct {
	Smtp SmtpConfig
}

// SmtpConfig — настройки SMTP для отправки писем. Все поля из env.
type SmtpConfig struct {
	Host     string `env:"SMTP_HOST" env-required:"true"`
	Port     string `env:"SMTP_PORT" env-required:"true"`
	Username string `env:"SMTP_USERNAME" env-required:"true"`
	Password string `env:"SMTP_PASSWORD" env-required:"true"`
	From     string `env:"SMTP_FROM" env-required:"true"`
}
