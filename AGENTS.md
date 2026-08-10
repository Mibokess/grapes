# Agent Guide

Start with [`docs/README.md`](docs/README.md). It contains the project mental model,
repository map, and a task-to-file routing table. Do not read the whole repository
to get oriented.

Read deeper documentation only when the task needs it:

- [`docs/architecture.md`](docs/architecture.md) for runtime flow, data loading,
  worktree merging, persistence, and TUI message routing.
- [`docs/development.md`](docs/development.md) for change recipes, test locations,
  golden files, configuration, and releases.

Treat the documentation as a map, then verify the relevant symbols in code before
changing behavior. If a change moves a responsibility, changes a runtime flow, or
invalidates a documented invariant, update the corresponding document in the same
change.

# Zens
Follow these zens as closely as possible, divering only if absolutely necessary.

## Coding Zen
Beautiful is better than ugly.
Explicit is better than implicit.
Simple is better than complex.
Complex is better than complicated.
Flat is better than nested.
Sparse is better than dense.
Readability counts.
Special cases aren't special enough to break the rules.
Although practicality beats purity.
Errors should never pass silently.
Unless explicitly silenced.
In the face of ambiguity, refuse the temptation to guess.
There should be one-- and preferably only one --obvious way to do it.
Although that way may not be obvious at first unless you're Dutch.
Now is better than never.
Although never is often better than *right* now.
If the implementation is hard to explain, it's a bad idea.
If the implementation is easy to explain, it may be a good idea.

## Writing Zen
Clear is better than impressive.
Precise is better than elegant.
Simple is better than complex.
Complex is better than vague.
Although leaving out the hard part is not simplicity.
Short is better than long.
Unless brevity costs meaning.
Every sentence should earn its place.
Old information comes before new.
Actors belong in subjects.
Actions belong in verbs.
Call the same thing by the same name.
One paragraph does one job.
Structure should show the reasoning.
Evidence beats assertion.
Reasoning beats evidence alone.
Separate what you know from what you infer.
A claim should be no stronger than its support.
Hedge the claim and not the sentence.
A citation is not an argument.
A technical term should save more than it costs.
Write for the reader and not for yourself.
Break any of these before writing something no one can follow.
