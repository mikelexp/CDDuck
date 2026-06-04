BIN := cdduck
BIN_DIR := $(HOME)/.local/bin

.PHONY: help build clean install uninstall version set-version aur-update

help:
	@echo "Targets:"
	@echo "  build        Build the binary"
	@echo "  clean        Remove build artifacts"
	@echo "  install      Build and install to ~/.local/bin"
	@echo "  uninstall    Remove the installed binary"
	@echo "  version      Print the current app version"
	@echo "  set-version  Set Version in version.go and PKGBUILD (use VERSION=...)"
	@echo "  aur-update   Update the AUR package"

build:
	go build -buildvcs=false -o $(BIN) .

clean:
	rm -rf $(BIN) release

install: build
	mkdir -p "$(BIN_DIR)"
	cp $(BIN) "$(BIN_DIR)/$(BIN)"

uninstall:
	rm -f "$(BIN_DIR)/$(BIN)"

version:
	@python3 -c "import re; from pathlib import Path; text = Path('version.go').read_text(); m = re.search(r'const Version = \"([^\"]+)\"', text); print(m.group(1) if m else '')"

set-version:
	@test -n "$(VERSION)" || (echo "Usage: make set-version VERSION=x.y.z"; exit 1)
	@python3 -c 'from pathlib import Path; import re, sys; version = sys.argv[1]; path = Path("version.go"); text = path.read_text(); text, count = re.subn(r"^const Version = \".*\"$$", f"const Version = \"{version}\"", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("version.go pattern not found")); path = Path("PKGBUILD"); text = path.read_text(); text, count = re.subn(r"^pkgver=.*$$", f"pkgver={version}", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("PKGBUILD pkgver pattern not found")); text = path.read_text(); text, count = re.subn(r"^pkgrel=.*$$", "pkgrel=1", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("PKGBUILD pkgrel pattern not found")); print(f"Set version to {version} in version.go and PKGBUILD")' "$(VERSION)"

aur-update:
	bash scripts/aur-update.sh
