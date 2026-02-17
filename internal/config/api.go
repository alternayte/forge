package config

// APIConfig holds configuration for the REST API layer.
// It maps to the [api] section in forge.toml.
type APIConfig struct {
	RateLimit RateLimitConfig `toml:"rate_limit"`
	CORS      CORSConfig      `toml:"cors"`
}

// RateLimitConfig holds rate limiting settings for the API.
type RateLimitConfig struct {
	// Enabled toggles rate limiting globally. Default: true.
	Enabled bool `toml:"enabled"`

	// Default applies to unauthenticated requests (default: 100 req/min).
	Default TierConfig `toml:"default"`

	// Authenticated applies to requests with a valid bearer token (default: 1000 req/min).
	Authenticated TierConfig `toml:"authenticated"`

	// APIKey applies to requests with a valid API key (default: 5000 req/min).
	APIKey TierConfig `toml:"api_key"`
}

// TierConfig defines the token bucket parameters for a single rate limit tier.
type TierConfig struct {
	// Tokens is the number of requests allowed per Interval.
	Tokens uint64 `toml:"tokens"`

	// Interval is a duration string controlling the bucket refill window (e.g. "1m", "1h").
	Interval string `toml:"interval"`
}

// CORSConfig holds cross-origin resource sharing settings.
type CORSConfig struct {
	// Enabled toggles CORS handling. Default: true.
	Enabled bool `toml:"enabled"`

	// AllowedOrigins is the list of origins that are allowed to make requests.
	// An empty list denies all cross-origin requests. Do NOT use "*" with
	// AllowCredentials — the middleware will log a warning and disable credentials.
	AllowedOrigins []string `toml:"allowed_origins"`

	// AllowedMethods is the list of HTTP methods to allow. Defaults to the most
	// common CRUD methods.
	AllowedMethods []string `toml:"allowed_methods"`

	// AllowedHeaders is the list of non-simple headers to allow in requests.
	AllowedHeaders []string `toml:"allowed_headers"`

	// ExposedHeaders is the list of response headers accessible to JavaScript.
	// Includes rate limit headers so clients can implement back-off.
	ExposedHeaders []string `toml:"exposed_headers"`

	// AllowCredentials permits cookies and authorization headers in cross-origin
	// requests. Cannot be combined with a wildcard AllowedOrigins.
	AllowCredentials bool `toml:"allow_credentials"`

	// MaxAge is the number of seconds the browser may cache a preflight response.
	// Default: 600 (10 minutes).
	MaxAge int `toml:"max_age"`
}

// DefaultAPIConfig returns an APIConfig populated with sensible production defaults.
func DefaultAPIConfig() APIConfig {
	return APIConfig{
		RateLimit: RateLimitConfig{
			Enabled: true,
			Default: TierConfig{
				Tokens:   100,
				Interval: "1m",
			},
			Authenticated: TierConfig{
				Tokens:   1000,
				Interval: "1m",
			},
			APIKey: TierConfig{
				Tokens:   5000,
				Interval: "1m",
			},
		},
		CORS: CORSConfig{
			Enabled: true,
			AllowedOrigins: []string{},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
			ExposedHeaders: []string{
				"X-RateLimit-Limit",
				"X-RateLimit-Remaining",
				"X-RateLimit-Reset",
			},
			AllowCredentials: true,
			MaxAge:           600,
		},
	}
}
