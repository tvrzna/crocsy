package main

import (
	"fmt"
	"os"

	"github.com/tvrzna/go-utils/args"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Servers []Server `yaml:"server"`
}

type Server struct {
	Listen string `yaml:"listen,omitempty"`
	TLS    struct {
		CertFile string `yaml:"cert_file,omitempty"`
		KeyFile  string `yaml:"key_file,omitempty"`
	} `yaml:"tls,omitempty"`
	Routes     []Route           `yaml:"route,omitempty"`
	Redirect   string            `yaml:"redirect,omitempty"`
	SetHeaders map[string]string `yaml:"set-headers,omitempty"`
}

type Route struct {
	Host            string            `yaml:"host,omitempty"`
	Path            string            `yaml:"path,omitempty"`
	Target          string            `yaml:"target,omitempty"`
	Root            string            `yaml:"root,omitempty"`
	Autoindex       bool              `yaml:"autoindex,omitempty"`
	Index           string            `yaml:"index,omitempty"`
	SetHeaders      map[string]string `yaml:"set-headers,omitempty"`
	ProxySetHeaders map[string]string `yaml:"proxy-set-headers,omitempty"`
}

var buildVersion string

func InitConfig(arg []string) (*Config, error) {
	printConfig := false
	configPath := "crocsy.yaml"
	args.ParseArgs(arg, func(arg, value string) {
		switch arg {
		case "-h", "--help":
			printHelp()
		case "-v", "--version":
			fmt.Printf("crocsy %s\nhttps://github.com/tvrzna/crocsy\n\nReleased under the MIT License.\n", getVersion())
			os.Exit(0)
		case "-c", "--config":
			configPath = value
		case "-C", "--print-config":
			printConfig = true
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

	if printConfig {
		data, err := yaml.Marshal(c)
		if err != nil {
			return nil, err
		}
		fmt.Print(string(data))
		os.Exit(0)
	}

	return &c, nil
}

func printHelp() {
	fmt.Printf(`Usage: crocsy [options]
Options:
	-h, --help		print this help
	-v, --version		print version
	-c, --config		set path to config file
	-C, --print-config	prints currently loaded configuration
`)
	os.Exit(0)
}

func getVersion() string {
	if buildVersion == "" {
		return "develop"
	}
	return buildVersion
}
