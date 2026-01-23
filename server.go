package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func startServer(s *Server) {
	// TODO: validate listen
	mux := http.NewServeMux()
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

		mux.Handle(r.Path+"/", proxy)
		mux.Handle(r.Path, proxy)
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
