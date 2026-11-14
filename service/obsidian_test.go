package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testVault() *ObsidianVault {
	return newObsidianVault([]string{"Notes/Sub/Other Note.md", "attachments/Picture One.png"})
}

func TestObsidianToMarkdown(t *testing.T) {
	input := "---\ntags: [a]\n---\n" +
		"[[Other Note]] [[Notes/Sub/Other Note#Part|alias]] [[Missing]]\n" +
		"![[Picture One.png|200]]\n" +
		"> [!warning] Careful\n" +
		"==mark== #tag %%hidden%% `[[code]]`\n" +
		"```\n[[fenced]]\n```"

	expected := "[Other Note](/Notes/Sub/Other%20Note.md) [alias](/Notes/Sub/Other%20Note.md) Missing\n" +
		`<img src="./attachments/picture-one.png" alt="Picture One" width="200">` + "\n" +
		"> **Careful**\n" +
		"<mark>mark</mark> `#tag`  `[[code]]`\n" +
		"```\n[[fenced]]\n```"

	output := string(testVault().ObsidianToMarkdown([]byte(input), "Index.md"))
	if output != expected {
		t.Fatalf("unexpected output:\n%s\nexpected:\n%s", output, expected)
	}
}

func TestIsObsidianIgnored(t *testing.T) {
	for name, ignored := range map[string]bool{".obsidian": true, ".trash": true, "notes": false} {
		if IsObsidianIgnored(name) != ignored {
			t.Errorf("%s ignored should be %v", name, ignored)
		}
	}
	if strings.Contains(webPath("a b/C.md"), " ") {
		t.Error("web paths must be encoded")
	}
}

func TestAnAliasedLinkInsideATableKeepsItsLink(test *testing.T) {
	input := "| Note |\n| --- |\n| [[Other Note\\|the other]] |"
	expected := "| Note |\n| --- |\n| [the other](/Notes/Sub/Other%20Note.md) |"

	output := string(testVault().ObsidianToMarkdown([]byte(input), "Index.md"))
	if output != expected {
		test.Fatalf("unexpected output:\n%s\nexpected:\n%s", output, expected)
	}
}

func TestALinkLeadsToTheNoteNearestTheLinkingNoteAsInObsidian(test *testing.T) {
	vault := newObsidianVault([]string{"Harbour.md", "Places/Harbour.md", "Places/Index.md"})
	output := string(vault.ObsidianToMarkdown([]byte("[[Harbour]]"), "Places/Index.md"))
	if expected := "[Harbour](/Places/Harbour.md)"; output != expected {
		test.Fatalf("got %s, expected %s", output, expected)
	}
}

func TestALinkMayNameANoteByItsAlias(test *testing.T) {
	root := test.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Quay.md"), []byte("---\naliases: [The old port]\n---\nText."), 0o644); nil != err {
		test.Fatal(err)
	}
	output := string(LoadObsidianVault(root).ObsidianToMarkdown([]byte("[[The old port]]"), "Index.md"))
	if expected := "[The old port](/Quay.md)"; output != expected {
		test.Fatalf("got %s, expected %s", output, expected)
	}
}

func TestCommentsOnTheirOwnLinesAndCalloutsWithoutTitleAreConverted(test *testing.T) {
	input := "Before\n%%\nhidden\n\nstill hidden\n%%\n> [!tip]\n> Body\nAfter"
	expected := "Before\n> **Tip**\n> Body\nAfter"
	output := string(testVault().ObsidianToMarkdown([]byte(input), "Index.md"))
	if output != expected {
		test.Fatalf("unexpected output:\n%s\nexpected:\n%s", output, expected)
	}
}

func TestAnOddCalloutCannotBreakTheConversion(test *testing.T) {
	for _, input := range []string{"> [!a%%b] c%%", "> [!note] ==a [[Other Note]] b==", "==a %%b== c%%"} {
		testVault().ObsidianToMarkdown([]byte(input), "Index.md") // must not panic
	}
}

func TestACommentAcrossLinesIsHidden(test *testing.T) {
	output := string(testVault().ObsidianToMarkdown([]byte("Seen %%not\nseen%% again."), "Index.md"))
	if output != "Seen  again." {
		test.Fatalf("got %q", output)
	}
}

func TestALinkInsideBoldOrAHighlightIsConverted(test *testing.T) {
	input := "**go to [[Other Note]]** and ==see [[Other Note|it]]=="
	expected := "**go to [Other Note](/Notes/Sub/Other%20Note.md)** and <mark>see [it](/Notes/Sub/Other%20Note.md)</mark>"
	if output := string(testVault().ObsidianToMarkdown([]byte(input), "Index.md")); output != expected {
		test.Fatalf("got %q", output)
	}
}

func TestANestedCalloutKeepsItsQuoteMarksAndItsLinks(test *testing.T) {
	input := "> > [!tip] See [[Other Note]]\n> > inside"
	expected := "> > **See [Other Note](/Notes/Sub/Other%20Note.md)**\n> > inside"
	if output := string(testVault().ObsidianToMarkdown([]byte(input), "Index.md")); output != expected {
		test.Fatalf("got %q", output)
	}
}

func TestACalloutWithoutTitleShowsItsTypeCapitalised(test *testing.T) {
	if output := string(testVault().ObsidianToMarkdown([]byte("> [!éclair]\n> text"), "Index.md")); output != "> **Éclair**\n> text" {
		test.Fatalf("got %q", output)
	}
}

func TestThePictureTextIsEscaped(test *testing.T) {
	vault := newObsidianVault([]string{`a "b" <c>.png`})
	output := string(vault.ObsidianToMarkdown([]byte(`![[a "b" <c>.png|200]]`), "Index.md"))
	if !strings.Contains(output, `alt="a &#34;b&#34; &lt;c&gt;"`) {
		test.Fatalf("got %q", output)
	}
}
