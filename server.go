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

		proxy := &httputil.ReverseProxy{}
		proxy.Rewrite = func(pr *httputil.ProxyRequest) {
			in := pr.In
			out := pr.Out

			replacer := newVariableReplacer(in)
			ctx := context.WithValue(in.Context(), ctxReplacerKey, replacer)
			*out = *out.WithContext(ctx)

			setHeaders(out.Header, replacer, r.ProxySetHeaders)

			out.URL.Scheme = targetURL.Scheme
			out.URL.Host = targetURL.Host
			out.Host = targetURL.Host

			out.URL.Path = strings.TrimPrefix(in.URL.Path, r.Path)
			if out.URL.Path == "" || !strings.HasPrefix(out.URL.Path, "/") {
				out.URL.Path = "/" + out.URL.Path
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
