package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type proxyRouter struct {
	targetURL    *url.URL
	server       *Server
	route        *Route
	reverseProxy *httputil.ReverseProxy
	fileHandler  http.Handler
}

func initProxyRouter(s *Server, r *Route, targetURL *url.URL) *proxyRouter {
	proxyRouter := &proxyRouter{targetURL, s, r, nil, nil}

	if r.Target != "" {
		proxyRouter.reverseProxy = &httputil.ReverseProxy{}
		proxyRouter.reverseProxy.Rewrite = rewriteFce(proxyRouter)
		proxyRouter.reverseProxy.ModifyResponse = modifyResponseFce(proxyRouter)
	} else if r.Root != "" {
		// TODO: do not standard http.FileServer
		proxyRouter.fileHandler = http.StripPrefix(r.Path, http.FileServer(http.FS(os.DirFS(r.Root))))
	}

	return proxyRouter
}

func rewriteFce(r *proxyRouter) func(pr *httputil.ProxyRequest) {
	return func(pr *httputil.ProxyRequest) {
		in := pr.In
		out := pr.Out

		replacer := newVariableReplacer(in)
		ctx := context.WithValue(in.Context(), ctxReplacerKey, replacer)
		*out = *out.WithContext(ctx)

		setHeaders(out.Header, replacer, r.route.ProxySetHeaders)

		out.URL.Scheme = r.targetURL.Scheme
		out.URL.Host = r.targetURL.Host
		out.Host = r.targetURL.Host

		out.URL.Path = strings.TrimPrefix(in.URL.Path, r.route.Path)
		if out.URL.Path == "" || !strings.HasPrefix(out.URL.Path, "/") {
			out.URL.Path = "/" + out.URL.Path
		}
	}
}

func modifyResponseFce(r *proxyRouter) func(res *http.Response) error {
	return func(res *http.Response) error {
		replacer, _ := (res.Request.Context().Value(ctxReplacerKey).(*VariableReplacer))
		setHeaders(res.Header, replacer, r.server.SetHeaders, r.route.SetHeaders)
		return nil
	}
}
