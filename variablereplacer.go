package main

import (
	"net"
	"net/http"
	"strings"
)

const (
	varHost       = "$host"
	varPort       = "$port"
	varScheme     = "$scheme"
	varRequestUri = "$request_uri"
)

type VariableReplacer struct {
	*strings.Replacer
}

func newVariableReplacer(r *http.Request) *VariableReplacer {
	host := r.Host
	hostname := host
	port := "80"
	if h, p, err := net.SplitHostPort(host); err == nil {
		hostname = h
		port = p
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	replacer := &VariableReplacer{strings.NewReplacer(
		varHost, hostname,
		varPort, port,
		varRequestUri, r.URL.RequestURI(),
		varScheme, scheme,
	)}

	return replacer
}
