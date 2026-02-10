package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	varHost       = "$host"
	varPort       = "$port"
	varRequestUri = "$request_uri"
)

func startServer(s *Server) {
	// TODO: validate listen
	mux := http.NewServeMux()

	if s.Redirect != "" {
		handleRedirect(s, mux)
	} else {
		proxyRoutes(s, mux)
	}

	server := &http.Server{
		Addr:    s.Listen,
		Handler: mux,
	}

	if s.TLS.CertFile != "" && s.TLS.KeyFile != "" {
		server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		go server.ListenAndServeTLS(s.TLS.CertFile, s.TLS.KeyFile)
	} else {
		go server.ListenAndServe()
	}
}

func proxyRoutes(s *Server, mux *http.ServeMux) {
	for _, route := range s.Routes {
		r := route
		targetURL, err := url.Parse(r.Target)
		if err != nil {
			log.Print(err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		originalDirector := proxy.Director

		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = targetURL.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, r.Path)
			if req.URL.Path == "" || !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path
			}
		}

		proxy.ModifyResponse = func(res *http.Response) error {
			for headerName, headerValue := range s.SetHeaders {
				res.Header.Set(headerName, headerValue)
			}
			for headerName, headerValue := range r.SetHeaders {
				res.Header.Set(headerName, headerValue)
			}
			return nil
		}

		mux.Handle(r.Path, proxy)
	}
}

func handleRedirect(s *Server, mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		hostname := host
		port := "80"
		if h, p, err := net.SplitHostPort(host); err == nil {
			hostname = h
			port = p
		}

		replacer := strings.NewReplacer(
			varHost, hostname,
			varPort, port,
			varRequestUri, r.URL.RequestURI(),
		)

		target := replacer.Replace(s.Redirect)

		for headerName, headerValue := range s.SetHeaders {
			w.Header().Set(headerName, headerValue)
		}

		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}
