// Package config holds the env-driven runtime configuration. All fields use
// struct tags consumed by caarlos0/env — no Viper, no YAML indirection.
package config

import "time"

type Config struct {
	ServerHost      string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	ServerPort      int           `env:"SERVER_PORT" envDefault:"8080"`
	Env             string        `env:"SERVER_ENV" envDefault:"development"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"20s"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat       string        `env:"LOG_FORMAT" envDefault:"text"`

	DBHost     string `env:"DB_HOST"     envDefault:"127.0.0.1"`
	DBPort     int    `env:"DB_PORT"     envDefault:"3306"`
	DBUser     string `env:"DB_USER"     envDefault:"root"`
	DBPassword string `env:"DB_PASSWORD" envDefault:""`
	DBName     string `env:"DB_NAME"     envDefault:"admin_template"`

	JWTSecret     string        `env:"JWT_SECRET"      envDefault:"change-me-in-production-please-use-32-bytes-min"`
	JWTAccessTTL  time.Duration `env:"JWT_ACCESS_TTL"  envDefault:"2h"`
	JWTRefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
	JWTIssuer     string        `env:"JWT_ISSUER"      envDefault:"admin-template"`

	StorageType     string `env:"STORAGE_TYPE"      envDefault:"local"`
	StorageLocalDir string `env:"STORAGE_LOCAL_PATH" envDefault:"./storage/uploads"`
	StorageBaseURL  string `env:"STORAGE_BASE_URL"   envDefault:"/files"`
	UploadMaxSize   int64  `env:"UPLOAD_MAX_SIZE"    envDefault:"10485760"` // 10 MiB

	CORSOrigins    string `env:"CORS_ORIGINS"     envDefault:"*"`
	SwaggerEnabled bool   `env:"SWAGGER_ENABLED"  envDefault:"false"`

	SeedAdminUsername string `env:"SEED_ADMIN_USERNAME" envDefault:"admin"`
	SeedAdminPassword string `env:"SEED_ADMIN_PASSWORD" envDefault:"admin123"`
}

// Load parses environment variables into Config. Fails fast on missing required
// fields — call once at startup.
func Load() (Config, error) {
	var c Config
	if err := envParse(&c); err != nil {
		return Config{}, err
	}
	return c, nil
}
