package discord

import (
	"errors"
	"strings"

	"github.com/Guuzzeji/teemysu/internal/db"
)

const maxTagCount = 5

var errNoText = errors.New("no text to bookmark")

func parseCommand(content string) (cmd, args string, ok bool) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "!") {
		return "", "", false
	}
	fields := strings.Fields(content)
	cmd = strings.ToLower(fields[0])
	args = strings.TrimSpace(strings.TrimPrefix(content, fields[0]))
	return cmd, args, true
}

func parseBookmarkArgs(args string) (tags []string, text string, err error) {
	args = strings.TrimSpace(args)
	if args == "" {
		return nil, "", errNoText
	}
	i := strings.IndexAny(args, " \t")
	if i < 0 {
		return nil, "", errNoText
	}
	tags = splitTags(args[:i])
	text = strings.TrimSpace(args[i+1:])
	if text == "" {
		return nil, "", errNoText
	}
	return tags, text, nil
}

func splitTags(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, p := range strings.Split(s, ",") {
		t := strings.ToLower(strings.TrimSpace(p))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

func parseTagResponse(resp string) []string {
	tags := splitTags(strings.NewReplacer("\n", ",", ";", ",", ": ", ",").Replace(resp))
	if len(tags) > maxTagCount {
		tags = tags[:maxTagCount]
	}
	return tags
}

func mergeTags(groups ...[]string) []string {
	var all []string
	for _, g := range groups {
		all = append(all, g...)
	}
	return splitTags(strings.Join(all, ","))
}

func tagTypes(tags []db.Tag) []string {
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = t.Type
	}
	return out
}

func tagsFromQuery(query string) []string {
	return splitTags(strings.ReplaceAll(query, " ", ","))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func lastN[T any](in []T, n int) []T {
	if len(in) <= n {
		return in
	}
	return in[len(in)-n:]
}
