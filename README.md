# papiro-indexer
The papiro indexer cli program

## Obsidian vaults

Use the `-obsidian` flag to index an Obsidian vault:

```sh
papiro-indexer -path ./my/vault -destination ./resources/wiki -obsidian
```

It converts:

- `[[note]]`, `[[folder/note#section|alias]]` wikilinks into links (resolved like Obsidian, by name or path).
- `![[image.png]]` and `![[image.png|300]]` embeds into images; note embeds into links.
- `> [!type] title` callouts into quotes with a bold title.
- `==highlights==`, `#tags` and `%%comments%%`.
- YAML frontmatter is removed and hidden folders (`.obsidian`, `.trash`) are not published.
