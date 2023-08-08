package server

import (
	"flag"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DbURI                string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() *Config {
	c := &Config{}
	setFlags(c)
	return c
}

func setFlags(c *Config) {
	flag.StringVar(&c.RunAddress, "a", ":8080", "Run Address for server")
	flag.StringVar(&c.DbURI, "d", "", "Database Uri")
	flag.StringVar(&c.AccrualSystemAddress, "r", "", "Accrual System Address")
}

func (c *Config) Parse() error {
	flag.Parse()
	if err := env.Parse(c); err != nil {
		return err
	}
	return nil
}
