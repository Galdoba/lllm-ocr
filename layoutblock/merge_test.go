package layoutblock

import (
	"strings"
	"testing"
)

func TestReplaceBRWithSpaces(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"self-closing", `<td>a<br/>b</td>`, `<td>a b</td>`},
		{"no-slash", `<h1>a<br>b</h1>`, `<h1>a b</h1>`},
		{"with-space", `<td>a<br />b</td>`, `<td>a b</td>`},
		{"uppercase", `<td>a<BR/>b</td>`, `<td>a b</td>`},
		{"multiple", `<td>a<br/><br>b<br/></td>`, `<td>a  b </td>`},
		{"none", `<td>ab</td>`, `<td>ab</td>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ReplaceBRWithSpaces(tc.in)
			if got != tc.want {
				t.Errorf("ReplaceBRWithSpaces(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeBlockLineBreaks(t *testing.T) {
	blocks := []LayoutBlock{
		{Label: LabelSectionHeader, Raw: `<h1>Title<br/>Sub</h1>`},
		{Label: LabelPageHeader, Raw: `<h2>Page<br>Header</h2>`},
		{Label: LabelTable, Raw: `<table><tr><td>a<br/>b</td></tr></table>`},
		{Label: LabelTOC, Raw: `<table><tr><td>1<br/>2</td></tr></table>`},
		{Label: LabelText, Raw: `<p>keep<br/>this</p>`},
		{Label: LabelListGroup, Raw: `<li>x<br/>y</li>`},
	}

	NormalizeBlockLineBreaks(blocks)

	for i, b := range blocks {
		switch b.Label {
		case LabelSectionHeader, LabelPageHeader, LabelTable, LabelTOC:
			if strings.Contains(b.Raw, "<br") {
				t.Errorf("label %s block %d still contains <br: %q", b.Label, i, b.Raw)
			}
		default:
			if !strings.Contains(b.Raw, "<br") {
				t.Errorf("label %s block %d was modified: %q", b.Label, i, b.Raw)
			}
		}
	}
}
