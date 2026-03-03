### 2026-03-03T16:15
[FINDINGS] Investigation complete. Here's the full analysis.

**Current state:**
- Go project using `gopkg.in/yaml.v3 v3.0.1`
- 4 Go files touch YAML: `loader.go`, `writer.go`, `validate.go`, `issue.go`
- ~42 YAML references across internal code
- 7 skill files + AGENTS.md + README.md + SPEC.md reference `meta.yaml` in docs
- No advanced YAML features used (no anchors, aliases, multi-doc, custom tags)
- Data is flat key-value with simple lists — trivially representable in TOML

**YAML touchpoints that would change:**
1. `loader.go` — `yaml.Unmarshal` → `toml.Unmarshal` (2 calls)
2. `writer.go` — `yaml.Marshal` → `toml.Marshal` (1 call), manual YAML string building in `SerializeIssue()`, and `sed -i` patterns in `UpdateField()`
3. `validate.go` — `yaml.Unmarshal` → `toml.Unmarshal` (2 calls)
4. `issue.go` — struct tag `yaml:"date"` on Comment (vestigial, comments are parsed from markdown not YAML)
5. All struct tags: `yaml:"field"` → `toml:"field"` across `meta`, `editMeta`, `Comment` structs

**Sed pattern change:**
- YAML: `s/^field: .*/field: newValue/`
- TOML: `s/^field = .*/field = "newValue"/` (strings need quotes, ints/bools don't)

**Frontmatter in SerializeIssue:**
- YAML uses `---` delimiters
- TOML convention is `+++` delimiters (used by Hugo, etc.)

**TOML library options for Go:**
- `github.com/BurntSushi/toml` — mature, v1.4+, widely used
- `github.com/pelletier/go-toml/v2` — faster, good Marshal/Unmarshal API

**Migration scope:**
- 4 Go source files (mechanical changes)
- 7 skill markdown files (documentation)
- 3 project docs (AGENTS.md, README.md, SPEC.md)
- All existing `.grapes/*/meta.yaml` → `.grapes/*/meta.toml` (rename + reformat)
- Could support both formats during transition by checking which file exists

**Verdict:** The migration is straightforward and mechanical. The main benefit is eliminating YAML type coercion footguns. The main cost is the breadth of documentation updates. Since agents are the primary writers of these files, the type safety benefit is modest — agents rarely write bare `no` or `3.10`. But TOML is arguably a more honest fit for what this data actually is: a flat config, not a document.
