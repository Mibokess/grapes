# Grapes TUI — Detail view screenshot
# Usage: vhs doc/vhs/detail.tape

Output doc/detail.png

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"

Type "cd $(pwd) && go run ."
Enter
Sleep 2s

# Open the first issue
Enter
Sleep 1.5s

Screenshot doc/detail.png
Escape
Sleep 500ms
Type "q"
