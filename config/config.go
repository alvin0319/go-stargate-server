package config

import (
	"io"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Host     string `toml:host`
	Port     int    `toml:port`
	Password string `toml:password`
}

// Read reads the Config from config.toml of current working directory and returns error if failed to read config.
func Read() (*Config, error) {
	var zero = Config{
		Host:     "0.0.0.0",
		Port:     47007,
		Password: "123456789",
	}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		b, err := toml.Marshal(zero)
		if err != nil {
			return nil, err
		}
		_ = os.WriteFile("config.toml", b, 0644)
		return &zero, nil
	}
	f, err := os.Open("config.toml")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var c Config
	err = toml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
