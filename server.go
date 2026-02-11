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
			setHeaders(res.Header, replacer, s.SetHeaders, r.SetHeaders)
			return nil
		}

		mux.Handle(r.Path, proxy)
	}
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
