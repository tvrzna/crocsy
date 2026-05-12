package main

import (
	"net"
	"net/http"
	"strings"
)

const (
	varHost         = "$host"
	varPort         = "$port"
	varScheme       = "$scheme"
	varRequestUri   = "$request_uri"
	varPath         = "$path"
	varMethod       = "$method"
	varQuery        = "$query"
	varRemoteAddr   = "$remote_addr"
	varHeaderCustom = "$header."
)

type VariableReplacer struct {
	replacer *strings.Replacer
	request  *http.Request
}

func newVariableReplacer(r *http.Request) *VariableReplacer {
	host := r.Host
	hostname := host
	port := "80"
	if h, p, err := net.SplitHostPort(host); err == nil {
		hostname = h
		port = p
	}

	var remoteAddr string
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remoteAddr = h
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
		varPath, r.URL.Path,
		varMethod, r.Method,
		varQuery, r.URL.RawQuery,
		varRemoteAddr, remoteAddr,
	), r}

	return replacer
}

func (v *VariableReplacer) Resolve(value string) string {
	if strings.HasPrefix(value, varHeaderCustom) {
		headerName := value[len(varHeaderCustom):]
		return v.request.Header.Get(headerName)
	}
	return v.replacer.Replace(value)
}
