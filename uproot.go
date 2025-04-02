// Copyright 2025 Ghislain Bourgeois
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghislainbourgeois/uproot/internal/gtpu"
	"github.com/ghislainbourgeois/uproot/internal/pfcp"
	"gopkg.in/yaml.v3"
)

type Config struct {
	UpfIP    string `yaml:"upfIP"`
	PfcpPort int    `yaml:"pfcpPort"`
	UpfN3IP  string `yaml:"upfN3IP"`
	GnbIP    string `yaml:"gnbIP"`
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	configFile := flag.String("config", "./uproot.yml", "Path to the configuration file to use")
	flag.Parse()
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("could not load configuration: %v", err)
	}

	conn, err := pfcp.NewPFCPConnection(config.UpfIP, config.PfcpPort)
	if err != nil {
		log.Fatalf("could not open PFCP connection: %v", err)
	}
	defer conn.Close()

	err = conn.Start()
	if err != nil {
		log.Fatalf("could not create PDU session on UPF: %v", err)
	}

	fmt.Println("PDU session active")

	tun, err := gtpu.NewTunnel(config.GnbIP, config.UpfN3IP)
	if err != nil {
		log.Fatalf("could not create GTP-U tunnel: %v", err)
	}
	defer tun.Close()

	fmt.Printf("GTP-U tunnel active on interface: %s\n", tun.Name)
	fmt.Println("Press ctrl-c to terminate")

	<-c
}

func loadConfig(configFile string) (Config, error) {
	var config Config
	configYaml, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}
	if err := yaml.Unmarshal(configYaml, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}
