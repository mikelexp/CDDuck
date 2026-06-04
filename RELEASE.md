# Publish a New Release

## 1. Bump the version

```bash
make set-version VERSION=0.1.1
# or
just set-version 0.1.1
```

## 2. Commit and tag

```bash
git commit -am "bump to v0.1.1"
git tag v0.1.1
git push github main v0.1.1
```

## 3. GitHub Release

The `.github/workflows/release.yml` workflow builds `cdduck-${VERSION}-linux-x86_64.tar.gz` and publishes the GitHub Release when the tag is pushed.

## 4. AUR

Once the Release is published:

```bash
make aur-update
# or
just aur-update
```

The script downloads the release tarball, computes the SHA256, updates the AUR `PKGBUILD`, validates with `makepkg`, and pushes the commit.
