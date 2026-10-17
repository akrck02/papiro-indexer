package service

import (
	"strings"
	"testing"
)

func testVault() *ObsidianVault {
	vault := &ObsidianVault{byName: map[string]string{}, byPath: map[string]string{}}
	vault.register("Notes/Sub/Other Note.md")
	vault.register("attachments/Picture One.png")
	return vault
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

	output := string(testVault().ObsidianToMarkdown([]byte(input)))
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
