<p align="center"><img alt="" src="docs/images/logo.svg" width="120"></p>

# papiro-indexer
The papiro indexer cli program

## Obsidian vaults

Use the `-obsidian` flag to index an Obsidian vault:

```sh
papiro-indexer -path ./my/vault -destination ./resources/wiki -obsidian
```

It converts:

- `[[note]]`, `[[folder/note#section|alias]]` wikilinks into links. As in Obsidian, a link leads to a path
  from the note's folder, then from the vault root, then to the nearest note with that name or alias.
- `![[image.png]]` and `![[image.png|300]]` embeds into images; note embeds into links.
- `> [!type] title` callouts into quotes with a bold title.
- `==highlights==` into `<mark>`, `#tags` into code, and `%%comments%%` are removed.
- YAML frontmatter is removed and hidden folders (`.obsidian`, `.trash`) are not published.

## AI disclosure

This project was originally handcrafted. Claude (Anthropic's AI assistant) has since helped with some changes, such as the Obsidian vault support.
