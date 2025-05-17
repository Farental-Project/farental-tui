package config

type Config struct {
	BaseURL  string
	Language string
}

func NewConfig() *Config {
	c := &Config{}

	c.BaseURL = "http://127.0.0.1:3000"
	c.Language = "en"

	return c
}
