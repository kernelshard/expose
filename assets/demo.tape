# Output file
Output docs/assets/images/demo.gif

# Terminal appearance
Set Shell "zsh"
Set FontSize 16
Set LineHeight 1.25
Set Width 800
Set Height 310
Set Padding 22
Set BorderRadius 10
Set WindowBar Colorful
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

# Prepare clean environment (hidden)
Hide
Type "export PATH=$PWD/assets/mock:$PATH"
Enter
Type "export PS1='%F{#89b4fa}❯%f '"
Enter
Type "clear"
Enter
Show

# Step 1: Type expose command
Sleep 500ms
Type "expose tunnel 3000"
Sleep 350ms
Enter

# Step 2: Show active tunnel
Sleep 3.5s

# Step 3: Graceful shutdown
Ctrl+C
Sleep 1.2s
