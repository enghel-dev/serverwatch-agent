package main

import (
	"fmt"
	"github.com/enghel-dev/serverwatch-agent/config"
	"github.com/enghel-dev/serverwatch-agent/installer"
	"github.com/enghel-dev/serverwatch-agent/agent"
)

func main() {
	if config.ConfigExists() {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error leyendo config:", err)
			return
		}
		agent.Run(cfg)
	} else {
		installer.RunCLI()
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error leyendo config:", err)
			return
		}
		agent.Run(cfg)
	}
}