package discord

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cases := []struct {
		in       string
		cmd, arg string
		ok       bool
	}{
		{"!b tag1,tag2 hello world", "!b", "tag1,tag2 hello world", true},
		{"!bookmark x", "!bookmark", "x", true},
		{"!B  upper", "!b", "upper", true},
		{"!s golang", "!s", "golang", true},
		{"!search golang project", "!search", "golang project", true},
		{"!chat how to build golang ftp", "!chat", "how to build golang ftp", true},
		{"!b-auto text here", "!b-auto", "text here", true},
		{"!bi this is a good website to keep in mind", "!bi", "this is a good website to keep in mind", true},
		{"!bookmark-inherit text", "!bookmark-inherit", "text", true},
		{"  !b  a  b ", "!b", "a  b", true},
		{"!b", "!b", "", true},
		{"just text", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		cmd, args, ok := parseCommand(c.in)
		if ok != c.ok || cmd != c.cmd || args != c.arg {
			t.Errorf("parseCommand(%q) = (%q, %q, %v), want (%q, %q, %v)", c.in, cmd, args, ok, c.cmd, c.arg, c.ok)
		}
	}
}

func TestParseBookmarkArgs(t *testing.T) {
	tags, text, err := parseBookmarkArgs("project,idea,golang http://example.com this is example")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(tags, []string{"project", "idea", "golang"}) {
		t.Errorf("tags = %v, want [project idea golang]", tags)
	}
	if text != "http://example.com this is example" {
		t.Errorf("text = %q", text)
	}

	if _, _, err := parseBookmarkArgs(""); err == nil {
		t.Errorf("expected error for empty args")
	}
	if _, _, err := parseBookmarkArgs("only,tags"); err == nil {
		t.Errorf("expected error when no text after tags")
	}
	if _, _, err := parseBookmarkArgs("tags   "); err == nil {
		t.Errorf("expected error for whitespace-only text")
	}
}

func TestSplitTags(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"project,idea,golang", []string{"project", "idea", "golang"}},
		{" Project , Idea ", []string{"project", "idea"}},
		{"a,a,b", []string{"a", "b"}},
		{",,a,,", []string{"a"}},
		{"", nil},
	}
	for _, c := range cases {
		got := splitTags(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitTags(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseTagResponse(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"golang, project, FTP", []string{"golang", "project", "ftp"}},
		{"golang\nproject\nftp", []string{"golang", "project", "ftp"}},
		{"a,b,c,d,e,f", []string{"a", "b", "c", "d", "e"}},
		{"Tags: golang, ftp", []string{"tags", "golang", "ftp"}},
	}
	for _, c := range cases {
		got := parseTagResponse(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseTagResponse(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestMergeTags(t *testing.T) {
	got := mergeTags([]string{"golang", "ftp"}, []string{"Golang", "idea"})
	want := []string{"golang", "ftp", "idea"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeTags = %v, want %v", got, want)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello world", 5); got != "hello" {
		t.Errorf("truncate = %q", got)
	}
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate = %q", got)
	}
}

func TestLastN(t *testing.T) {
	in := []int{1, 2, 3, 4}
	if got := lastN(in, 2); !reflect.DeepEqual(got, []int{3, 4}) {
		t.Errorf("lastN = %v", got)
	}
	if got := lastN(in, 10); !reflect.DeepEqual(got, in) {
		t.Errorf("lastN = %v", got)
	}
	if got := lastN([]int{}, 2); len(got) != 0 {
		t.Errorf("lastN empty = %v", got)
	}
}
