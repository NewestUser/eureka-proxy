package strip

import (
	"fmt"
	"net/http"
	"strings"
)

type String interface {
	apply(path string) string
}

func New(pattern string) (String, error) {

	find, replace, err := resolveStrip(pattern)

	if err != nil {
		return nil, err
	}

	return &stringStrip{find: find, replace: replace}, nil
}

type stringStrip struct {
	find    string
	replace string
}

func (s *stringStrip) apply(path string) string {

	return strings.Replace(path, s.find, s.replace, 1)
}

func resolveStrip(strip string) (string, string, error) {

	if strip == "" {
		return "", "", nil
	}

	if !strings.Contains(strip, ":") {
		return "", "", fmt.Errorf("incorrect strip format: '%v' example 'foo:bar'", strip)
	}

	parts := strings.Split(strip, ":")

	if len(parts) == 1 {
		return parts[0], "", nil
	}

	return parts[0], parts[1], nil
}

func NewHandler(strip String, chain http.Handler) http.Handler {

	return &stripHandler{s: strip, chain: chain}
}

type stripHandler struct {
	s     String
	chain http.Handler
}

func (h *stripHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = h.s.apply(r.URL.Path)
	h.chain.ServeHTTP(w, r)
}
