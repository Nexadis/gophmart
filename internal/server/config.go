package server

import (
	"flag"

	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/caarlos0/env/v9"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DBURI                string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JwtSecret            string `env:"JWT_SECRET"`
	Wait                 int64  `env:"WAIT"`
}

func NewConfig() *Config {
	c := &Config{}
	setFlags(c)
	return c
}

func setFlags(c *Config) {
	flag.StringVar(&c.RunAddress, "a", ":8080", "Run Address for server")
	flag.StringVar(&c.DBURI, "d", "", "Database Uri")
	flag.StringVar(&c.AccrualSystemAddress, "r", "", "Accrual System Address")
	flag.Int64Var(&c.Wait, "t", 1, "Timeout for get accruals")
}

func (c *Config) Parse() error {
	flag.Parse()
	if err := env.Parse(c); err != nil {
		return err
	}
	logger.Logger.Info(`Config:
	RunAddress: %q
	DBUri: %q
	AccrualSystemAddress: %q
	JwtSecret: %q
	Interval get Accruals: %d`)
	return nil
}
