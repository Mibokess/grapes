#!/bin/bash
# Usage: create-mr.sh <title> <description-file> [extra glab flags...]
#
# Reads the MR description from a file to avoid command substitution
# (which triggers sandbox permission prompts) in the calling Bash command.
#
# Examples:
#   bash .claude/skills/merge-request/create-mr.sh \
#       "ETH-00042: Fix parser bug" /tmp/claude/ETH-00042/mr-description.md
#
#   bash .claude/skills/merge-request/create-mr.sh \
#       "ETH-00042: Fix parser bug" /tmp/claude/ETH-00042/mr-description.md \
#       --target-branch experiments
set -e

TITLE="$1"
DESC_FILE="$2"
shift 2

DESCRIPTION=$(cat "$DESC_FILE")
glab mr create --title "$TITLE" --description "$DESCRIPTION" "$@"
