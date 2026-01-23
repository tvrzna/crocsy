package main

import (
	"fmt"
	"os"

	"github.com/tvrzna/go-utils/args"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Servers []Server `yaml:"server"`
}

type Server struct {
	Listen string `yaml:"listen"`
	TLS    struct {
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"tls"`
	Routes []Route `yaml:"route"`
}

type Route struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
}

var buildVersion string

func InitConfig(arg []string) (*Config, error) {
	configPath := "crocsy.yaml"
	args.ParseArgs(arg, func(arg, value string) {
		switch arg {
		case "-h", "--help":
			printHelp()
		case "-v", "--version":
			fmt.Printf("lerry %s\nhttps://github.com/tvrzna/crocsy\n\nReleased under the MIT License.\n", getVersion())
			os.Exit(0)
		case "-c", "--config":
			configPath = value
		}
	})

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func printHelp() {
	fmt.Printf(`Usage: crocsy [options]
Options:
	-h, --help		print this help
	-v, --version		print version
	-c, --config		set path to config file
`)
	os.Exit(0)
}

func getVersion() string {
	if buildVersion == "" {
		return "develop"
	}
	return buildVersion
}
