default:
	@just --list

build:
	go build -buildvcs=false -o cdduck .

clean:
	rm -rf cdduck release

install: build
	mkdir -p ~/.local/bin
	cp cdduck ~/.local/bin/cdduck
	@printf '%s\n' '' 'CDDuck installed to ~/.local/bin/cdduck' '' 'Add this to your shell rc:' '' '  cdd() {' '      local dir' '      dir="$(cdduck)" || return' '      if [ -n "$dir" ]; then' '          cd -- "$dir"' '      fi' '  }' '' 'For fish:' '' '  function cdd' '      set dir (cdduck)' '      if test -n "$dir"' '          cd -- "$dir"' '      end' '  end'

uninstall:
	rm -f ~/.local/bin/cdduck

version:
	@python3 -c "import re; from pathlib import Path; text = Path('version.go').read_text(); m = re.search(r'const Version = \"([^\"]+)\"', text); print(m.group(1) if m else '')"

set-version VERSION:
	@python3 -c 'from pathlib import Path; import re, sys; version = sys.argv[1]; path = Path("version.go"); text = path.read_text(); text, count = re.subn(r"^const Version = \".*\"$", f"const Version = \"{version}\"", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("version.go pattern not found")); path = Path("PKGBUILD"); text = path.read_text(); text, count = re.subn(r"^pkgver=.*$", f"pkgver={version}", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("PKGBUILD pkgver pattern not found")); text = path.read_text(); text, count = re.subn(r"^pkgrel=.*$", "pkgrel=1", text, flags=re.M); path.write_text(text) if count else (_ for _ in ()).throw(SystemExit("PKGBUILD pkgrel pattern not found")); print(f"Set version to {version} in version.go and PKGBUILD")' "{{VERSION}}"

aur-update:
	bash scripts/aur-update.sh
