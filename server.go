package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	ctxReplacerKey = "replacer"
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

			ctx := context.WithValue(req.Context(), ctxReplacerKey, newVariableReplacer(req))
			*req = *req.WithContext(ctx)

			req.Host = targetURL.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, r.Path)
			if req.URL.Path == "" || !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path
			}

		}

		proxy.ModifyResponse = func(res *http.Response) error {
			replacer, _ := (res.Request.Context().Value(ctxReplacerKey).(*VariableReplacer))
			for headerName, headerValue := range s.SetHeaders {
				res.Header.Set(headerName, replacer.Replace(headerValue))
			}
			for headerName, headerValue := range r.SetHeaders {
				res.Header.Set(headerName, replacer.Replace(headerValue))
			}
			return nil
		}

		mux.Handle(r.Path, proxy)
	}
}

func handleRedirect(s *Server, mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		replacer := newVariableReplacer(r)
		for headerName, headerValue := range s.SetHeaders {
			w.Header().Set(headerName, replacer.Replace(headerValue))
		}
		http.Redirect(w, r, replacer.Replace(s.Redirect), http.StatusMovedPermanently)
	})
}
