# CDDuck

A TUI file browser that replaces `cd` in the terminal. Browse the filesystem visually and exit into the selected directory.

## Install

```bash
make install
```

The installer prints the shell setup snippet after copying the binary.

## Shell Wrapper

The binary prints the selected path to stdout and renders the TUI on `/dev/tty`. You still need a shell function to perform the actual `cd`.

### Bash / Zsh

```bash
cdd() {
    local dir
    dir="$(cdduck)" || return
    if [ -n "$dir" ]; then
        cd -- "$dir"
    fi
}
```

### Fish

```fish
function cdd
    set dir (cdduck)
    if test -n "$dir"
        cd -- "$dir"
    end
end
```

Add the function to `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`, then reload your shell with `source`.

To check the installed version, run `cdduck --version`.

## Usage

Run `cdd` in the terminal. The browser opens fullscreen.

### Keys

| Key | Action |
|---|---|
| `↑` / `↓` | Move between items |
| `PgUp` / `PgDn` | Page up / down |
| `Home` / `End` | First / last item |
| `Enter` | Open the selected directory |
| `Alt+Enter` | Select the current folder and exit |
| `Backspace` | Go up a level (empty filter) / delete a character (active filter) |
| `Esc` | Clear the filter (if there is text) or exit into the current directory |
| `Ctrl+H` | Jump to your home directory |
| `Ctrl+C` / `Ctrl+Q` | Quit without selecting |

### UI

- **Highlight**: the selected line uses a blue background, white text, and bold weight.
- **Path**: the full current directory is shown below the box, after the `Current path:` label.
- **Directories**: folders are listed first, in cyan, sorted alphabetically.
- **Frame**: the box uses a double blue border with the title centered at the top.
- **Labels**: `Current path:` and `Filter:` are rendered in blue.
- **Filter**: type to fuzzy-filter items case-insensitively in real time. `Esc` clears it.

To navigate, open a folder with `Enter` and press `Alt+Enter` when you want to exit into that directory. You can also press `Esc` with an empty filter to exit into the current directory.

## Build From Source

Requires Go 1.26.3+.

```bash
git clone ...
cd cdduck
go build -buildvcs=false -o cdduck .
```

## License

MIT
