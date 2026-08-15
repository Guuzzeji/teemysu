package discord

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// urlRe matches http(s) URLs up to whitespace or common trailing punctuation.
var urlRe = regexp.MustCompile(`https?://[^\s<>]+`)

// metaClient is a short-timeout client so a slow page never blocks a bookmark.
var metaClient = &http.Client{Timeout: 8 * time.Second}

// pageMeta is the metadata pulled from a webpage.
type pageMeta struct {
	Title       string
	Description string
}

// extractURLs returns each unique http(s) URL found in s.
func extractURLs(s string) []string {
	seen := map[string]bool{}
	var urls []string
	for _, u := range urlRe.FindAllString(s, -1) {
		u = strings.TrimRight(u, ".,;!?)]}\"'")
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		urls = append(urls, u)
	}
	return urls
}

// enrichText pulls title/description metadata for every URL in text and appends
// it so the embedding captures what the linked pages are about. If there are no
// URLs, or every fetch fails or yields nothing, text is returned unchanged, so
// callers fall back to the current behavior automatically.
func (b *Bot) enrichText(ctx context.Context, text string) string {
	urls := extractURLs(text)
	if len(urls) == 0 {
		return text
	}
	meta := fetchAllMeta(ctx, urls)
	if len(meta) == 0 {
		return text
	}
	var sb strings.Builder
	sb.WriteString(text)
	for _, u := range urls {
		m, ok := meta[u]
		if !ok {
			continue
		}
		fmt.Fprintf(&sb, "\n[%s", u)
		if m.Title != "" {
			fmt.Fprintf(&sb, " title: %s", m.Title)
		}
		if m.Description != "" {
			fmt.Fprintf(&sb, " description: %s", m.Description)
		}
		sb.WriteString("]")
	}
	return sb.String()
}

// fetchAllMeta fetches all URLs concurrently so the total cost is bounded by the
// slowest page, not the sum of all pages. Pages that fail or return no metadata
// are omitted.
func fetchAllMeta(ctx context.Context, urls []string) map[string]pageMeta {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out := make(map[string]pageMeta)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			m, err := fetchMeta(ctx, u)
			if err != nil || (m.Title == "" && m.Description == "") {
				return
			}
			mu.Lock()
			out[u] = m
			mu.Unlock()
		}(u)
	}
	wg.Wait()
	return out
}

func fetchMeta(ctx context.Context, url string) (pageMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return pageMeta{}, err
	}
	req.Header.Set("User-Agent", "teemysu-bookmark-bot/1.0")
	resp, err := metaClient.Do(req)
	if err != nil {
		return pageMeta{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return pageMeta{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	// Bound the body so a huge page cannot balloon memory.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return pageMeta{}, err
	}
	return parseMeta(body), nil
}

// parseMeta pulls og:title/og:description, falling back to <title> and
// <meta name="description">. ponytail: regex parse, no HTML dependency; upgrade
// if pages render their metadata client-side.
func parseMeta(body []byte) pageMeta {
	s := string(body)
	var m pageMeta
	m.Title = metaContent(s, `property="og:title"`)
	if m.Title == "" {
		m.Title = tagContent(s, "title")
	}
	m.Description = metaContent(s, `property="og:description"`)
	if m.Description == "" {
		m.Description = metaContent(s, `name="description"`)
	}
	return m
}

// metaContent returns the content attribute of a <meta> element whose attributes
// contain attr, e.g. property="og:title" or name="description".
func metaContent(s, attr string) string {
	re := regexp.MustCompile(`(?is)<meta[^>]*` + regexp.QuoteMeta(attr) + `[^>]*content="([^"]*)"`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return html.UnescapeString(strings.TrimSpace(m[1]))
}

// tagContent returns the inner text of the first <tag> element.
func tagContent(s, tag string) string {
	re := regexp.MustCompile(`(?is)<` + tag + `[^>]*>(.*?)</` + tag + `>`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return html.UnescapeString(strings.TrimSpace(m[1]))
}