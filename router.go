package main

import (
	"net"
	"net/http"
	"sort"
	"strings"
)

func handleProxyRoutes(compiledRoutes []*proxyRouter) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}

		for _, r := range compiledRoutes {
			if !hostMatches(r.route.Host, host) {
				continue
			}

			if !strings.HasPrefix(req.URL.Path, r.route.Path) || req.URL.Path+"/" == r.route.Path {
				continue
			}

			if r.reverseProxy != nil {
				r.reverseProxy.ServeHTTP(w, req)
			} else if r.fileServer != nil {
				r.fileServer.Serve(w, req)
			}

			return
		}

		http.NotFound(w, req)
	}
}

func hostMatches(pattern, host string) bool {
	if pattern == "" {
		return true
	}

	patterns := strings.Split(pattern, " ")

	for _, p := range patterns {
		if p == host {
			return true
		}
		if strings.HasPrefix(p, "*.") {
			return strings.HasSuffix(host, p[1:])
		}
	}
	return false
}

func sortCompiledRoutes(routes []*proxyRouter) {
	sort.Slice(routes, func(i, j int) bool {
		ri := routes[i]
		rj := routes[j]

		if ri.route.Host == "" && rj.route.Host != "" {
			return false
		}
		if ri.route.Host != "" && rj.route.Host == "" {
			return true
		}

		if ri.route.Host != rj.route.Host {
			return ri.route.Host < rj.route.Host
		}

		if len(ri.route.Path) != len(rj.route.Path) {
			return len(ri.route.Path) > len(rj.route.Path)
		}

		return ri.route.Path < rj.route.Path
	})
}
