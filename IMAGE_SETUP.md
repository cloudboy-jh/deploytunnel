# Deploy Tunnel - Image Display Setup

The startup image feature has been successfully implemented! The system will automatically display `deploytunnel.png` using the best available method.

## How It Works

### 1. Image Display Priority

The system tries these methods in order:

1. **Kitty Terminal Protocol** - Full color, high quality (Kitty terminal)
2. **iTerm2 Protocol** - Full color (iTerm2 on macOS)
3. **Sixel Protocol** - Paletted color (mlterm, yaft, etc.)
4. **ASCII Art Fallback** - Text-based rendering (all terminals)

### 2. Automatic Detection

The system automatically detects your terminal capabilities by checking:
- `$TERM_PROGRAM` environment variable
- `$KITTY_WINDOW_ID` environment variable
- `$TERM` environment variable
- Terminal capability queries

### 3. Image Location

The image is searched in these locations (in order):
1. Current directory: `deploytunnel.png`
2. Relative to executable: `../deploytunnel.png`
3. Working directory: `$PWD/deploytunnel.png`

**Current image**: `deploytunnel.png` (500x500 PNG)

## Supported Terminals

### Full Image Support (Protocol-based)
- **Kitty** - Best quality, full protocol support
- **iTerm2** - macOS, full color inline images
- **WezTerm** - Cross-platform, Kitty protocol compatible
- **mlterm** - Sixel protocol
- **yaft** - Sixel protocol

### ASCII Art Fallback (Always Works)
- **Terminal.app** (macOS default)
- **Alacritty**
- **Windows Terminal**
- **GNOME Terminal**
- **Konsole**
- **xterm**
- Any other standard terminal

## Testing

### Test in iTerm2 (macOS)
```bash
# iTerm2 should display the full PNG image
./dt help
```

### Test in Kitty
```bash
# Kitty should display the full PNG using Kitty protocol
kitty ./dt help
```

### Test ASCII Fallback
```bash
# Most terminals will show ASCII art version
./dt help
```

## Display Settings

### Current Configuration
- **Image width**: 75% of terminal width
- **Max width**: 80 columns
- **Min width**: 40 columns
- **Auto-scaling**: Yes
- **Position**: Above "DEPLOY ▸ TUNNEL" text header
- **Cache**: ASCII art is cached after first generation

## Customization

You can adjust the display settings in `internal/tui/image.go`:

### Change Image Size
```go
// Line ~131: Adjust percentage
targetCols := uint32(float64(termWidth) * 0.75)  // Change 0.75 to 0.5 for 50%
```

### Change ASCII Art Width
```go
// Line ~175: Adjust max/min width
convertOptions.FixedWidth = int(float64(termWidth) * 0.75)
if convertOptions.FixedWidth > 80 {  // Change max width here
    convertOptions.FixedWidth = 80
}
```

### Disable Image Display
To temporarily disable image display, you can:

1. Rename/move `deploytunnel.png` (system will gracefully fall back to text only)
2. Or modify `internal/tui/styles.go`:
```go
func Header() string {
    // Comment out image display
    // imageDisplay := DisplayImage()
    
    title := TitleStyle.Render("DEPLOY ▸ TUNNEL")
    subtitle := SubtitleStyle.Render("migrate safely between hosts")
    
    return lipgloss.JoinVertical(
        lipgloss.Left,
        "",
        title,
        subtitle,
        "",
    )
}
```

## Troubleshooting

### Image Not Displaying

**Check 1: Is the image file present?**
```bash
ls -la deploytunnel.png
# Should show: deploytunnel.png: PNG image data, 500 x 500
```

**Check 2: Which terminal are you using?**
```bash
echo $TERM_PROGRAM
echo $TERM
```

**Check 3: Try forcing ASCII art**
If image protocols aren't working, the system should automatically fall back to ASCII art.

### ASCII Art Quality

The ASCII art quality depends on:
- Terminal width (wider = more detail)
- Font (monospace fonts work best)
- Image contrast (high contrast = better ASCII)

### Performance

- **First load**: May take 100-200ms to generate ASCII art
- **Subsequent loads**: Instant (cached)
- **Image protocols**: Very fast (<50ms)

## File Structure

```
deploytunnel/
├── deploytunnel.png          # Your logo (500x500 PNG)
├── internal/tui/
│   ├── image.go              # Image display logic
│   └── styles.go             # Header() calls DisplayImage()
```

## Dependencies

- `github.com/BourgeoisBear/rasterm` - Terminal image protocols
- `github.com/qeesung/image2ascii` - ASCII art generation
- `golang.org/x/term` - Terminal size detection

## Notes

- The image will be displayed on **all TUI screens** (dashboard, init wizard, auth)
- If the image file is missing, the system gracefully falls back to text-only header
- ASCII art is cached to avoid regenerating on every screen
- The system automatically detects terminal resize and adjusts accordingly

---

**Enjoy your branded Deploy Tunnel experience!**
