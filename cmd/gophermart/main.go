package main

import (
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/server"
)

func main() {
	config := server.NewConfig()
	config.Parse()
	s, err := server.New(config)
	if err != nil {
		logger.Logger.Errorln(err)
		return
	}
	logger.Logger.Errorln(s.Run())
}
