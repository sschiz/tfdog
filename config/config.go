package config

import (
	"github.com/vaughan0/go-ini"
)

// LoadConfig loads the application configuration.
func LoadConfig(path string) (ini.File, error) {
	file, err := ini.LoadFile(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}
