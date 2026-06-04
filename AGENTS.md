# CDDuck

TUI file browser que reemplaza `cd`. Escrito en Go con [bubbletea](https://github.com/charmbracelet/bubbletea).

## Build

```bash
go build -buildvcs=false -o cdduck .
```

## Install / Release

- `make install` o `just install` copian el binario a `~/.local/bin` y muestran el snippet de shell.
- También se puede instalar directo desde AUR como `cdduck-bin`.
- `make set-version VERSION=x.y.z` y `just set-version x.y.z` actualizan `version.go` y `PKGBUILD`.
- `make aur-update` y `just aur-update` publican el release en AUR desde el tarball de GitHub Releases.

## Arquitectura

- `main.go` — todo en un archivo, modelo bubbletea (~370 líneas)
- **Modelo**: `Model` struct con el directorio actual, items, cursor, offset, filtro, etc.
- **Items**: estructura plana (no tree), estilo Midnight Commander con `..` para subir
- **Filtro**: fuzzy match case-insensitive en vivo, `Esc` lo limpia o sale al directorio actual si ya está vacío

## Detalle clave: salida TUI vs resultado

El TUI de bubbletea escribe a `/dev/tty` (no a stdout) mediante `tea.WithOutput(tty)`. El path seleccionado se imprime en stdout con `fmt.Println`. Esto permite que el binario se use con `$()` en una función shell sin que la interfaz se pierda:

```zsh
cdd() {
    local dir
    dir="$(cdduck)" || return
    if [ -n "$dir" ]; then
        cd -- "$dir"
    fi
}
```

## Detalle clave: cursor highlight

El cursor se pinta con **códigos ANSI crudos** (`\033[44;37;1m` = fondo azul, texto blanco, negrita), no con lipgloss. Los directorios usan ANSI 16 (`\033[36;1m` = cyan). Los labels y bordes usan azul ANSI (`\033[34;1m`).

## Teclas

| Tecla | Acción |
|---|---|
| `↑` / `↓` | Navegar |
| `PgUp` / `PgDn` | Página |
| `Home` / `End` | Primero / último |
| `Enter` | Entrar al directorio |
| `Alt+Enter` | Seleccionar y salir |
| `Backspace` | Subir (filtro vacío) / borrar (filtro activo) |
| `Esc` | Limpiar filtro / salir al directorio actual |
| `Ctrl+C` / `Ctrl+Q` | Salir sin seleccionar |

## Dependencias

- `github.com/charmbracelet/bubbletea` — TUI framework
