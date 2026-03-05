# Grapes TUI Demo
# Usage: vhs doc/vhs/demo.tape

Output doc/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# Launch grapes
Type "grapes"
Enter
Sleep 2s

# Board view — pause to take it in
Sleep 2s

# Navigate across columns
Type "l"
Sleep 400ms
Type "l"
Sleep 400ms
Type "l"
Sleep 1s

# Open issue detail to show metadata and relationships
Enter
Sleep 2.5s

# Scroll down in detail view
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 1.5s

# Back to board
Escape
Sleep 1s

# Switch to list view
Type "L"
Sleep 2s

# Search for "theme"
Type "/"
Sleep 400ms
Type "theme"
Sleep 2s
Escape
Sleep 1s

# Open filter menu
Type "f"
Sleep 1.5s

# Navigate into status filter
Enter
Sleep 1s

# Select a status
Type "j"
Sleep 300ms
Type " "
Sleep 500ms
Enter
Sleep 1s

# Close filter menu to see filtered results
Escape
Sleep 2s

# Back to board with filters
Type "B"
Sleep 2s

# Quit
Type "q"
Sleep 500ms
