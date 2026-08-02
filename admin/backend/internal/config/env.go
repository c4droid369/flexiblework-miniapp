package config

import "github.com/caarlos0/env/v11"

// envParse is a thin wrapper over caarlos0/env so callers don't import the
// library directly. Swap library here without touching every Load() caller.
func envParse(c *Config) error {
	return env.Parse(c)
}
