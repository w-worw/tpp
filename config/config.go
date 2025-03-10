package config

type APIConfig struct {
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"dev_password"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"postgres"`
	StripeSecretKey  string `env:"STRIPE_SECRET_KEY" envDefault:"sk_test_51QzSz2QDL7aRcA28lH8n5gdyi43ZEfdPsuIppkz2AAB5XbE0WZIbDSLI1WGoBbd3bpa8pWQgHTIiayYh3iGhhJDS005fM8TqXR"`
	WebhookSecret    string `env:"STRIPE_WEBHOOK_SECRET" envDefault:"whsec_2c6bffa991ce95972a9b822e156148e70a8133970468a5739ddca50a0ef338ab"`
}
