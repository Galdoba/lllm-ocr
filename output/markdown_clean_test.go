package output

import (
	"strings"
	"testing"

	"github.com/Galdoba/lllm-ocr/layoutblock"
)

const bullet = "\u2022"

func TestCleanMarkdown_UnorderedDoubleMarker(t *testing.T) {
	in := strings.Join([]string{
		"- " + bullet + " _A PC works for another faction and is going to sabotage the work._",
		"- " + bullet + " _One of the PCs has been planning to kill another for some time._",
		"- " + bullet + " _The ship is being pursued or shadowed by a rival._",
		"- " + bullet + " _Something unusual is on-board, endangering the PCs when they find it._",
	}, "\n")
	want := strings.Join([]string{
		"- _A PC works for another faction and is going to sabotage the work._",
		"- _One of the PCs has been planning to kill another for some time._",
		"- _The ship is being pursued or shadowed by a rival._",
		"- _Something unusual is on-board, endangering the PCs when they find it._",
	}, "\n")

	if got := cleanMarkdown(in); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestCleanMarkdown_OrderedDoubleMarker(t *testing.T) {
	in := strings.Join([]string{
		"1. 1\\. Create characters.",
		"2. 2\\. Decide if the group will have their own starship.",
		"3. 3\\. If they have a starship, apply a suitable Skills Package (see Traveller Core Book)",
		"4. 4\\. If desired, select or roll for a starting situation 'In Media Res'.",
		"5. 5\\. Begin play ...",
	}, "\n")
	want := strings.Join([]string{
		"1. Create characters.",
		"2. Decide if the group will have their own starship.",
		"3. If they have a starship, apply a suitable Skills Package (see Traveller Core Book)",
		"4. If desired, select or roll for a starting situation 'In Media Res'.",
		"5. Begin play ...",
	}, "\n")

	if got := cleanMarkdown(in); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestCleanMarkdown_OrderedMarkerNotStrippedWhenNumbersDiffer(t *testing.T) {
	in := "1. Phase 1: setup\n2. Debrief."
	if got := cleanMarkdown(in); got != in {
		t.Errorf("unrelated numbered content was altered:\ngot:\n%q\nwant:\n%q", got, in)
	}
}

func TestCleanMarkdown_UnorderedUnaffectedWhenNoBullet(t *testing.T) {
	in := "- _A PC works for another faction._"
	if got := cleanMarkdown(in); got != in {
		t.Errorf("normal list item was altered:\ngot:\n%q\nwant:\n%q", got, in)
	}
}

func TestCleanMarkdown_EmptyHeaderRowRemoved(t *testing.T) {
	in := strings.Join([]string{
		"|   |   |",
		"| --- | --- |",
		"| **2D6 Combat Resolution Consequence** |   |",
		"| 2-3 | A PC or NPC severely wounded |",
		"| 8-12 | No consequence |",
		"",
	}, "\n")
	want := strings.Join([]string{
		"| **2D6 Combat Resolution Consequence** |   |",
		"| --- | --- |",
		"| 2-3 | A PC or NPC severely wounded |",
		"| 8-12 | No consequence |",
		"",
	}, "\n")

	if got := cleanMarkdown(in); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestCleanMarkdown_MergeIdenticalHeaderTables(t *testing.T) {
	tableA := strings.Join([]string{
		"| D66 | Patron |",
		"| --- | --- |",
		"| 11 | Naval Officer |",
		"| 36 | Gangster |",
	}, "\n")
	tableB := strings.Join([]string{
		"| D66 | Patron |",
		"| --- | --- |",
		"| 41 | Corporate Official |",
		"| 66 | Ship-Owner |",
	}, "\n")
	in := tableA + "\n\n" + tableB

	want := strings.Join([]string{
		"| D66 | Patron |",
		"| --- | --- |",
		"| 11 | Naval Officer |",
		"| 36 | Gangster |",
		"| 41 | Corporate Official |",
		"| 66 | Ship-Owner |",
	}, "\n")

	if got := cleanMarkdown(in); got != want {
		t.Errorf("tables were not merged:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestCleanMarkdown_MergeThreeConsecutiveTables(t *testing.T) {
	tbl := func(rows ...string) string { return strings.Join(rows, "\n") }
	tableA := tbl("| D66 | Patron |", "| --- | --- |", "| 11 | Naval Officer |")
	tableB := tbl("| D66 | Patron |", "| --- | --- |", "| 22 | Marine Officer |")
	tableC := tbl("| D66 | Patron |", "| --- | --- |", "| 33 | Merchant |")
	in := tableA + "\n\n" + tableB + "\n\n" + tableC

	want := strings.Join([]string{
		"| D66 | Patron |",
		"| --- | --- |",
		"| 11 | Naval Officer |",
		"| 22 | Marine Officer |",
		"| 33 | Merchant |",
	}, "\n")

	if got := cleanMarkdown(in); got != want {
		t.Errorf("three same-header tables were not fully merged:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestCleanMarkdown_DifferentHeaderTablesNotMerged(t *testing.T) {
	tableA := "| D66 | Skill |\n| --- | --- |\n| 11 | Boat |\n"
	tableB := "| D20 | Ability |\n| --- | --- |\n| 1 | Strength |\n"
	in := tableA + "\n\n" + tableB

	got := cleanMarkdown(in)
	if got != in {
		t.Errorf("tables with different headers should be left untouched:\ngot:\n%q\nwant:\n%q", got, in)
	}
}

func TestMarkdownOutputConvert_CleansDoubleMarker(t *testing.T) {
	blocks := []layoutblock.LayoutBlock{
		{
			Label: layoutblock.LabelListGroup,
			Raw:   "<ul><li>" + bullet + " item one</li><li>" + bullet + " item two</li></ul>",
			Page:  1,
		},
	}

	out := &MarkdownOutput{Config: MarkdownConfig{PageSeparator: ""}}
	got := out.Convert(blocks)

	want := "- item one\n- item two"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}
