package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	ctxReplacerKey = "replacer"
)

type compiledRoute struct {
	Route
	proxy *httputil.ReverseProxy
}

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
	var compiledRoutes []*proxyRouter
	for _, route := range s.Routes {
		r := route
		targetURL, err := url.Parse(r.Target)
		if err != nil {
			log.Print(err)
			continue
		}

		compiledRoutes = append(compiledRoutes, initProxyRouter(s, &route, targetURL))
	}
	sortCompiledRoutes(compiledRoutes)

	mux.HandleFunc("/", handleProxyRoutes(compiledRoutes))
}

func handleRedirect(s *Server, mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		replacer := newVariableReplacer(r)
		setHeaders(w.Header(), replacer, s.SetHeaders)
		http.Redirect(w, r, replacer.Replace(s.Redirect), http.StatusMovedPermanently)
	})
}

func setHeaders(header http.Header, replacer *VariableReplacer, setHeaders ...map[string]string) {
	for _, setHeader := range setHeaders {
		for headerName, headerValue := range setHeader {
			if replacer != nil {
				header.Set(headerName, replacer.Replace(headerValue))
			} else {
				header.Set(headerName, headerValue)
			}
		}
	}
}
