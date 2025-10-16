package router

import (
	"net/http"
	"strings"
)

type PathGuard struct {
	next          http.Handler
	exacts        map[string]struct{}
	prefixes      []string
	redirectSlash bool
}

func (s *PathGuard) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if path != "/" && strings.HasSuffix(path, "/") {
		canon := strings.TrimSuffix(path, "/")
		if _, ok := s.exacts[canon]; ok {
			if s.redirectSlash {
				http.Redirect(w, req, canon, http.StatusMovedPermanently)
				return
			}
			req.URL.Path = canon
			path = canon
		}
	}

	if _, ok := s.exacts[path]; ok {
		s.next.ServeHTTP(w, req)
		return
	}
	for _, p := range s.prefixes {
		if strings.HasPrefix(path, p) {
			s.next.ServeHTTP(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

func NewPathGuard(next http.Handler, exacts map[string]struct{}, prefixes []string, redirectSlash bool) http.Handler {
	return &PathGuard{next: next, exacts: exacts, prefixes: prefixes, redirectSlash: redirectSlash}
}
