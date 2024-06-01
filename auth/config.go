package auth

type Config struct {
	Secret                 string `env:"AUTH_SECRET"`
	AccessLifetimeSeconds  int    `env:"AUTH_ACCESS_LIEFTIME"`
	RefreshLifetimeSeconds int    `env:"AUTH_REFRESH_LIFETIME"`
}
