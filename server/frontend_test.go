package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestFrontendDevelopmentCacheHeaders(t *testing.T) {
	files := fstest.MapFS{
		"index.html":     {Data: []byte("<html>preview</html>")},
		"assets/app.js":  {Data: []byte("console.log('preview')")},
		"assets/app.css": {Data: []byte("body { color: white }")},
	}
	for _, devMode := range []bool{false, true} {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			for _, path := range []string{"/", "/assets/app.js", "/assets/app.css"} {
				response := httptest.NewRecorder()
				frontendHandler(files, devMode).ServeHTTP(response, httptest.NewRequest(method, path, nil))
				if response.Code != http.StatusOK {
					t.Fatalf("%s %s status = %d", method, path, response.Code)
				}
				for header, value := range map[string]string{
					"Cache-Control": "no-store, no-cache, must-revalidate",
					"Pragma":        "no-cache",
					"Expires":       "0",
				} {
					want := ""
					if devMode {
						want = value
					}
					if got := response.Header().Get(header); got != want {
						t.Fatalf("devMode=%v %s %s %s = %q, want %q", devMode, method, path, header, got, want)
					}
				}
			}
		}
	}
}
