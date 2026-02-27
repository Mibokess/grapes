### matt — 2026-02-27
The header regex on line 154 also accepts three different dash characters (`[—–-]`), but the spec in `idea.md` only uses em-dash (`—`). The fix should probably standardize on em-dash and reject the others, or at least document which is canonical.
