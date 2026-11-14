package affiliate

import (
	"strings"
	"testing"
)

// Shapes observed in real prod posts (2026-11-14): reasoning appended
// after the link, and the prompt echoed at the start of the post.
func TestSanitizeSocialReview(t *testing.T) {
	const link = "https://amazon.com/dp/B0D8N7X9M1?tag=your-tag-20"
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			"reasoning tail after link is cut",
			"Great earbuds, love the ANC! Affiliate: " + link + "\n\nBut does it mention it's an affiliate link? Yes, \"Affiliate:\" should suffice but maybe",
			"Great earbuds, love the ANC! Affiliate: " + link,
		},
		{
			"leading prompt echo is stripped",
			"The user wants a punchy, honest one-post product recommendation. Been using the $159 PLAUD NotePin and it's shockingly tiny. Affiliate: " + link,
			"Been using the $159 PLAUD NotePin and it's shockingly tiny. Affiliate: " + link,
		},
		{
			"code fence stripped",
			"```\nSolid pick for freelancers. Affiliate: " + link + "\n```",
			"Solid pick for freelancers. Affiliate: " + link,
		},
		{
			"clean copy passes through",
			"Solid pick. Affiliate: " + link,
			"Solid pick. Affiliate: " + link,
		},
		{
			"no link: whitespace normalized only",
			"line one\n\nline   two",
			"line one line two",
		},
	}
	for _, c := range cases {
		if got := sanitizeSocialReview(c.in, link); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
	// Result must always fit one post after formatting.
	long := strings.Repeat("word ", 120) + link + " trailing reasoning"
	if got := sanitizeSocialReview(long, link); strings.Contains(got, "trailing") {
		t.Errorf("trailing reasoning survived: %q", got)
	}
}
