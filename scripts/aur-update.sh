#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_VERSION="$(make -s -C "${ROOT_DIR}" version)"
REPO_NAME="cdduck-bin"
AUR_SSH="ssh://aur@aur.archlinux.org/${REPO_NAME}.git"
WORK_DIR="$(mktemp -d /tmp/cdduck-aur-XXXXX)"

cleanup() {
  rm -rf "${WORK_DIR}"
  rm -f "${ROOT_DIR}/cdduck-${APP_VERSION}-linux-x86_64.tar.gz"
}

trap cleanup EXIT

echo "=== Updating AUR package ${REPO_NAME} to version ${APP_VERSION} ==="

cd "${ROOT_DIR}"
gh release download "v${APP_VERSION}" --repo mikelexp/CDDuck --pattern "cdduck-${APP_VERSION}-linux-x86_64.tar.gz" --clobber

HASH="$(sha256sum "cdduck-${APP_VERSION}-linux-x86_64.tar.gz" | cut -d' ' -f1)"
echo "SHA256: ${HASH}"

echo "Cloning AUR repo..."
git clone "${AUR_SSH}" "${WORK_DIR}"

cp "${ROOT_DIR}/PKGBUILD" "${WORK_DIR}/"
cp "${ROOT_DIR}/cdduck.install" "${WORK_DIR}/"

cd "${WORK_DIR}"
sed -i "s/^pkgver=.*/pkgver=${APP_VERSION}/" PKGBUILD
sed -i "s/^pkgrel=.*/pkgrel=1/" PKGBUILD
sed -i "s/^sha256sums=.*/sha256sums=('${HASH}')/" PKGBUILD

makepkg -s --noconfirm
makepkg --printsrcinfo > .SRCINFO

git add PKGBUILD .SRCINFO cdduck.install
git commit -m "bump to v${APP_VERSION}"
git push origin master

echo "=== Done ==="
