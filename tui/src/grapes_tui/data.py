"""Issue data loading — reads .grapes/ directory."""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from pathlib import Path

import yaml

STATUSES = ["backlog", "todo", "in_progress", "done", "cancelled"]
STATUS_LABELS = {
    "backlog": "Backlog",
    "todo": "To Do",
    "in_progress": "In Progress",
    "done": "Done",
    "cancelled": "Cancelled",
}
PRIORITY_LABELS = {
    "urgent": "Urgent",
    "high": "High",
    "medium": "Medium",
    "low": "Low",
}
PRIORITY_ORDER = {"urgent": 0, "high": 1, "medium": 2, "low": 3}


@dataclass
class Comment:
    author: str
    date: str
    body: str


@dataclass
class Issue:
    id: int
    title: str
    status: str
    priority: str
    assignee: str
    labels: list[str]
    created: str
    updated: str
    parent: int | None = None
    children: list[int] = field(default_factory=list)
    content: str = ""
    comments_raw: str = ""


def find_issues_dir() -> Path:
    """Find .grapes/ directory by walking up from this file's location."""
    d = Path(__file__).resolve().parent
    for _ in range(10):
        candidate = d / ".grapes"
        if candidate.is_dir():
            return candidate
        d = d.parent
    raise FileNotFoundError("Could not find .grapes/ directory")


def load_all_issues(issues_dir: Path | None = None) -> list[Issue]:
    """Load all issue summaries (meta only)."""
    if issues_dir is None:
        issues_dir = find_issues_dir()

    issues: list[Issue] = []
    for entry in sorted(issues_dir.iterdir()):
        if entry.is_dir() and entry.name.isdigit():
            meta_path = entry / "meta.yaml"
            if not meta_path.exists():
                continue
            meta = yaml.safe_load(meta_path.read_text())
            issues.append(
                Issue(
                    id=int(entry.name),
                    title=meta.get("title", ""),
                    status=meta.get("status", "backlog"),
                    priority=meta.get("priority", "medium"),
                    assignee=meta.get("assignee", ""),
                    labels=meta.get("labels", []),
                    parent=meta.get("parent"),
                    created=str(meta.get("created", "")),
                    updated=str(meta.get("updated", "")),
                )
            )

    # Build children arrays
    issue_map = {i.id: i for i in issues}
    for issue in issues:
        if issue.parent is not None and issue.parent in issue_map:
            issue_map[issue.parent].children.append(issue.id)

    return issues


def load_issue(issue_id: int, issues_dir: Path | None = None) -> Issue | None:
    """Load a single issue with full content and comments."""
    if issues_dir is None:
        issues_dir = find_issues_dir()

    issue_dir = issues_dir / str(issue_id)
    if not issue_dir.is_dir():
        return None

    meta_path = issue_dir / "meta.yaml"
    if not meta_path.exists():
        return None

    meta = yaml.safe_load(meta_path.read_text())
    content = ""
    comments_raw = ""

    content_path = issue_dir / "content.md"
    if content_path.exists():
        content = content_path.read_text()

    comments_path = issue_dir / "comments.md"
    if comments_path.exists():
        comments_raw = comments_path.read_text()

    # Get children
    all_issues = load_all_issues(issues_dir)
    children = [i.id for i in all_issues if i.parent == issue_id]

    return Issue(
        id=issue_id,
        title=meta.get("title", ""),
        status=meta.get("status", "backlog"),
        priority=meta.get("priority", "medium"),
        assignee=meta.get("assignee", ""),
        labels=meta.get("labels", []),
        parent=meta.get("parent"),
        created=str(meta.get("created", "")),
        updated=str(meta.get("updated", "")),
        children=children,
        content=content,
        comments_raw=comments_raw,
    )


def parse_comments(raw: str) -> list[Comment]:
    """Parse comments.md into structured Comment objects."""
    if not raw.strip():
        return []

    blocks = re.split(r"^### ", raw, flags=re.MULTILINE)
    comments = []
    for block in blocks:
        block = block.strip()
        if not block:
            continue
        lines = block.split("\n")
        header_match = re.match(r"^(.+?)\s+[—–-]+\s+(.+)$", lines[0])
        author = header_match.group(1).strip() if header_match else "Unknown"
        date = header_match.group(2).strip() if header_match else ""
        body = "\n".join(lines[1:]).strip()
        comments.append(Comment(author=author, date=date, body=body))

    return comments
