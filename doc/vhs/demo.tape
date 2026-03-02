# Grapes TUI Demo — flows through board, list, and detail views
# Usage: vhs doc/vhs/demo.tape

Output doc/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# Start the app
Type "cd $(pwd) && go run ."
Enter
Sleep 4s

# Board view — navigate around
Sleep 1.5s
Type "l"
Sleep 600ms
Type "l"
Sleep 600ms
Type "j"
Sleep 600ms

# Switch to list view
Type "L"
Sleep 1.5s

# Navigate the list
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Open issue detail
Enter
Sleep 2s

# Scroll down to see comments
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 1.5s

# Go back
Escape
Sleep 1s

# Back to board
Type "B"
Sleep 1.5s

# Quit
Type "q"
Sleep 500ms
