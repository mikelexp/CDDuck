#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="${HOME}/.local/bin"
BIN_PATH="${BIN_DIR}/cdduck"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "==> Compilando cdduck..."
cd "$SCRIPT_DIR"
go build -buildvcs=false -o cdduck .

echo "==> Instalando en ${BIN_PATH}..."
mkdir -p "$BIN_DIR"
cp cdduck "$BIN_PATH"

# Detectar shell y preparar snippet de función
case "${SHELL}" in
  *zsh)
    RC_FILE="${HOME}/.zshrc"
    FUNC_SNIPPET='# --- CDDuck ---
cdd() {
    local dir
    dir="$(cdduck)" || return
    if [ -n "$dir" ]; then
        cd -- "$dir"
    fi
}'
    ;;
  *bash)
    RC_FILE="${HOME}/.bashrc"
    FUNC_SNIPPET='# --- CDDuck ---
cdd() {
    local dir
    dir="$(cdduck)" || return
    if [ -n "$dir" ]; then
        cd -- "$dir"
    fi
}'
    ;;
  *fish)
    RC_FILE="${HOME}/.config/fish/config.fish"
    FUNC_SNIPPET='# --- CDDuck ---
function cdd
    set dir (cdduck)
    if test -n "$dir"
        cd -- "$dir"
    end
end'
    ;;
  *)
    RC_FILE=""
    FUNC_SNIPPET='# --- CDDuck ---
cdd() {
    local dir
    dir="$(cdduck)" || return
    if [ -n "$dir" ]; then
        cd -- "$dir"
    fi
}'
    ;;
esac

echo ""
echo "===== Instalación completa ====="
echo ""
echo "Binario: ${BIN_PATH}"

# Verificar si ~/.local/bin está en PATH
case ":$PATH:" in
  *:"${BIN_DIR}":*) 
    echo "PATH:   ~/.local/bin ya está en PATH ✓"
    ;;
  *)
    echo "PATH:   ~/.local/bin NO está en PATH"
    echo ""
    echo "=> Agregando ~/.local/bin al PATH en ${RC_FILE}..."
    echo "" >> "$RC_FILE"
    echo '# CDDuck: agregar ~/.local/bin al PATH' >> "$RC_FILE"
    echo 'export PATH="${PATH}:${HOME}/.local/bin"' >> "$RC_FILE"
    echo "   ✓ Listo (se agregó al final de ${RC_FILE})"
    ;;
esac

# Verificar si la función cdd ya está definida en el RC
if [ -n "$RC_FILE" ] && grep -q "cdd()" "$RC_FILE" 2>/dev/null; then
  echo "Función: ya está definida en ${RC_FILE}"
else
  echo ""
  echo "=> Agregando función cdd a ${RC_FILE}..."
  echo "" >> "$RC_FILE"
  echo "$FUNC_SNIPPET" >> "$RC_FILE"
  echo "   ✓ Listo"
fi

echo ""
echo "===== Recargá la shell con: ====="
echo ""
echo "    source ${RC_FILE}"
echo ""
echo "Y después usá:"
echo ""
echo "    cdd"
