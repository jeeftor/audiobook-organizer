Output output/docs-visuals/cli/cli-organize-run.gif

Require bin/audiobook-organizer

Set Shell "bash"
Set FontSize 24
Set Width 1400
Set Height 760
Set Theme "TokyoNight"
Set Padding 24
Set Framerate 24
Set PlaybackSpeed 1.15
Set TypingSpeed 18ms
Set WindowBar Colorful
Set BorderRadius 8
Env NO_COLOR "1"

Type "bin/audiobook-organizer --dir=output/docs-cli-sample/metadata-json/source --out=output/docs-cli-sample/metadata-json/organized --layout=author-only"
Enter
Sleep 2s
Ctrl+L
Sleep 500ms
Type "cd output/docs-cli-sample/metadata-json/organized"
Enter
Sleep 500ms
Type "find . -type f ! -name '.abook-org.log' | sort"
Enter
Sleep 4s
