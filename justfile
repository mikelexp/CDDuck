# List available recipes
default: help

# List available recipes
help:
	@just --list

# Build the binary
build:
	go build -buildvcs=false -o cdduck .

# Remove build artifacts
clean:
	rm -rf cdduck release

# Build and install to ~/.local/bin
install: build
	mkdir -p ~/.local/bin
	cp cdduck ~/.local/bin/cdduck
	@printf '%s\n' '' 'CDDuck installed to ~/.local/bin/cdduck' '' 'Add this to your shell rc:' '' '  cdd() {' '      if [ "$1" = "--version" ] || [ "$1" = "-V" ]; then' '          cdduck "$@"' '          return' '      fi' '      local dir' '      dir="$(cdduck "$@")" || return' '      if [ -n "$dir" ]; then' '          cd -- "$dir"' '      fi' '  }' '' 'For fish:' '' '  function cdd' '      if test "$argv[1]" = --version; or test "$argv[1]" = -V' '          cdduck $argv' '          return' '      end' '      set dir (cdduck $argv)' '      if test -n "$dir"' '          cd -- "$dir"' '      end' '  end'

# Remove the installed binary
uninstall:
	rm -f ~/.local/bin/cdduck

# Print the current app version
version:
	@python3 -c "import re; from pathlib import Path; text = Path('version.go').read_text(); m = re.search(r'const Version = \"([^\"]+)\"', text); print(m.group(1) if m else '')"

# Set Version in version.go and PKGBUILD
set-version VERSION:
	@python3 -c 'from pathlib import Path; import re, sys; version = sys.argv[1]; path = Path("version.go"); text = path.read_text(); text, count = re.subn(r"^const Version = \".*\"$", f"const Version = \"{version}\"", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("version.go pattern not found")); path = Path("PKGBUILD"); text = path.read_text(); text, count = re.subn(r"^pkgver=.*$", f"pkgver={version}", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("PKGBUILD pkgver pattern not found")); text = path.read_text(); text, count = re.subn(r"^pkgrel=.*$", "pkgrel=1", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("PKGBUILD pkgrel pattern not found")); print(f"Set version to {version} in version.go and PKGBUILD")' "{{VERSION}}"

# Update the AUR package
aur-update:
	bash scripts/aur-update.sh
