# Grapes TUI — Board view screenshot
# Usage: vhs doc/vhs/board.tape

Output doc/board.png

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"

Type "cd $(pwd) && go run ."
Enter
Sleep 2s

Screenshot doc/board.png
Type "q"
