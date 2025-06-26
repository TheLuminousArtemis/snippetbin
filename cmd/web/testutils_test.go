package main

import (
	"bytes"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/ratelimiter"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/store"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/store/cache"
)

func newConfig(t *testing.T) config {
	t.Helper()
	cfg := config{
		rlCfg: ratelimiterConfig{
			RequestsPerTimeFrame: 20,
			Timeframe:            time.Second,
			Enabled:              true,
		},
	}
	return cfg
}

func newTestApplication(t *testing.T, cfg config) *application {
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}
	mockCache := cache.NewMockStorage()
	formDecoder := form.NewDecoder()
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true
	ratelimiter := ratelimiter.NewRedisFixedWindowRateLimiter(mockCache, cfg.rlCfg.RequestsPerTimeFrame, cfg.rlCfg.Timeframe)
	storage := store.NewStorage()
	return &application{
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		cache:          mockCache,
		rateLimiter:    ratelimiter,
		store:          storage,
	}
}

type testServer struct {
	*httptest.Server
}

// func (app *application) testRoutes() http.Handler {
// 	r := chi.NewMux()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Logger)
// 	r.Use(middleware.Recoverer)
// 	r.Use(app.recoverPanic)
// 	r.Use(app.logRequest)
// 	r.Use(commonHeaders)
// 	r.Get("/ping", ping)
// 	return r
// }

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

// var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

var csrfTokenRX = regexp.MustCompile(`<input type=["']hidden["'] name=["']csrf_token["'] value=["']([^"']+)["']`)

func extractCSRFToken(t *testing.T, body string) string {
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}

	return html.UnescapeString(string(matches[1]))
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	// rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	req, err := http.NewRequest(http.MethodPost, ts.URL+urlPath, strings.NewReader(form.Encode()))

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}
