package metadata

import "testing"

func TestSplitICYTitle(t *testing.T) {
	artist, title := SplitICYTitle("BONES - HDMI")
	if artist != "BONES" || title != "HDMI" {
		t.Fatalf("unexpected split: %q %q", artist, title)
	}
}

func TestDisplayFallbacks(t *testing.T) {
	cases := []struct {
		name string
		md   Metadata
		want string
	}{
		{"artist title", Metadata{Artist: "A", Title: "T"}, "A — T"},
		{"title", Metadata{Title: "T"}, "T"},
		{"raw", Metadata{Raw: "R"}, "R"},
		{"empty", Metadata{}, "—"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.md.Display(); got != tc.want {
				t.Fatalf("Display() = %q, want %q", got, tc.want)
			}
		})
	}
}
