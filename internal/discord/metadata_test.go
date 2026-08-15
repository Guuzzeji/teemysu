package discord

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestExtractURLs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"no links here", nil},
		{"https://golang.org a", []string{"https://golang.org"}},
		{"http://a.com and https://b.com/x", []string{"http://a.com", "https://b.com/x"}},
		{"https://a.com. done", []string{"https://a.com"}},
		{"same https://a.com and https://a.com", []string{"https://a.com"}},
		{"(https://a.com) tail", []string{"https://a.com"}},
	}
	for _, c := range cases {
		got := extractURLs(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("extractURLs(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseMeta(t *testing.T) {
	body := `<html><head>
		<title>Fallback Title</title>
		<meta name="description" content="fallback desc">
		<meta property="og:title" content="Open &amp; Graph Title">
		<meta property="og:description" content="og description here">
	</head></html>`
	m := parseMeta([]byte(body))
	if m.Title != "Open & Graph Title" {
		t.Errorf("title = %q, want og title", m.Title)
	}
	if m.Description != "og description here" {
		t.Errorf("description = %q, want og description", m.Description)
	}

	// Fallback to <title>/name=description when og tags are missing.
	fallback := parseMeta([]byte(`<title>Plain</title><meta name="description" content="desc">`))
	if fallback.Title != "Plain" || fallback.Description != "desc" {
		t.Errorf("fallback = %+v", fallback)
	}

	empty := parseMeta([]byte(`<html><body>no meta</body></html>`))
	if empty.Title != "" || empty.Description != "" {
		t.Errorf("empty = %+v, want zero", empty)
	}
}

func TestEnrichTextFallback(t *testing.T) {
	b := &Bot{}
	if got := b.enrichText(context.Background(), "just text"); got != "just text" {
		t.Errorf("no-url fallback = %q", got)
	}
	// example.invalid never resolves: fetch fails, text unchanged.
	if got := b.enrichText(context.Background(), "see https://example.invalid/x now"); got != "see https://example.invalid/x now" {
		t.Errorf("fetch-fail fallback = %q", got)
	}
}

func TestEnrichTextAppendsMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<meta property="og:title" content="Go Channels">
			<meta property="og:description" content="how to share memory between goroutines">`))
	}))
	defer srv.Close()

	got := (&Bot{}).enrichText(context.Background(), "note "+srv.URL+" mutex idea")
	if !strings.Contains(got, "Go Channels") || !strings.Contains(got, "share memory") {
		t.Errorf("enriched text missing page metadata, got: %q", got)
	}
}