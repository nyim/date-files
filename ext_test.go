package main

import (
    "testing"
)

func TestGuessExt(t *testing.T) {
    cases := []struct {
		in, want string
	}{
		{"a.mp4", ".mp4"},
		{"a.MP4", ".mp4"},
		{"asdfa", ""},
		{"a.MP4.~13~", ".mp4"},
		{"a.MP4~", ".mp4"},
		{"a.MP4~.~1~", ".mp4"},
	}

    for _, c := range cases {
        got := GuessExt(c.in)
        if got != c.want {
            t.Errorf("GuessExt(%q) == %q, want %q", c.in, got, c.want)
        }
    }
}
