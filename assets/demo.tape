# Output file
Output docs/assets/images/demo.gif

# Terminal appearance
Set Shell "zsh"
Set FontSize 15
Set FontFamily "JetBrains Mono"
Set Width 850
Set Height 360
Set Padding 20
Set WindowBar Colorful
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms

# Setup environment
Env PATH "/Users/sam/code/opensource/expose/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"

# Prepare environment (hidden)
Hide
Type "python3 -m http.server 3000 > /dev/null 2>&1 &"
Enter
Type "clear"
Enter
Show

# Step 1: Run expose tunnel
Sleep 600ms
Type "expose tunnel 3000"
Sleep 400ms
Enter

# Step 2: Show the active tunnel
Sleep 3.5s

# Step 3: Graceful shutdown
Type "^C"
Sleep 1.5s

# Cleanup (hidden)
Hide
Type "kill $(lsof -t -i :3000) 2>/dev/null || true"
Enter
