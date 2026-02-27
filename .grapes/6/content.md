The detail screen renders sub-issues with a `.sub-issue` CSS class that has a `:hover` underline style, implying they're clickable. But there's no click handler — clicking does nothing.

## Current code

```python
# detail.py:62-67 — CSS promises interactivity
DetailScreen .sub-issue {
    margin-left: 2;
}
DetailScreen .sub-issue:hover {
    text-style: underline;  # suggests it's clickable
}

# detail.py:119-122 — but no on_click handler
yield Label(
    f"  #{child.id}: {child.title} — {status}",
    classes="sub-issue",
)
```

## Fix

Either make them clickable (navigate to the child issue) or remove the hover underline to avoid misleading users.

To make them clickable, the simplest approach is a custom `Label` subclass or a `Static` with an `on_click`:

```python
sub = Label(f"  #{child.id}: {child.title} — {status}", classes="sub-issue")
sub.issue_id = child.id
yield sub

# Then handle click on the screen:
def on_label_clicked(self, event):
    if hasattr(event.label, "issue_id"):
        self.app.push_screen("detail", {"issue_id": event.label.issue_id})
```

Or use a `Button` variant styled to look like a link.
