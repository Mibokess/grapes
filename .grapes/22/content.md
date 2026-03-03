## Goal
Investigate whether switching from YAML to TOML for issue metadata files is worthwhile, and what the migration would involve.

## Context
Grapes currently uses `meta.yaml` for issue metadata. TOML may be a better fit due to:
- No implicit type coercion bugs (YAML's `no` → `false`, `3.10` → `3.1`)
- Native datetime support
- Simpler, smaller spec

## Research Questions
- Where in the codebase does Grapes read/write `meta.yaml`?
- What YAML library is currently used? What TOML libraries are available for the same language?
- Are there any YAML-specific features being relied on (anchors, multi-doc, complex nesting)?
- What would the migration path look like (rename files, swap parser, backward compat)?
- Are there downstream consumers (scripts, CI, other tools) that depend on the YAML format?

## Deliverable
A summary of findings with a recommendation (migrate, don't migrate, or migrate with caveats).
