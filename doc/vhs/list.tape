# Grapes TUI — List view screenshot
# Usage: vhs doc/vhs/list.tape

Output doc/list.png

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"

Type "cd $(pwd) && go run ."
Enter
Sleep 4s

# Switch to list view
Type "L"
Sleep 1.5s

Screenshot doc/list.png
Type "q"
