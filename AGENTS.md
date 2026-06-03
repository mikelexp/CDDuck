# CDDuck

TUI file browser que reemplaza `cd`. Escrito en Go con [bubbletea](https://github.com/charmbracelet/bubbletea) y [lipgloss](https://github.com/charmbracelet/lipgloss).

## Build

```bash
go build -buildvcs=false -o cdduck .
```

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

El cursor se pinta con **códigos ANSI 16 crudos** (`\033[44;37;1m` = fondo azul, texto blanco, negrita), no con lipgloss. Esto es porque algunos terminales no soportan colores 256 ni true color para backgrounds. Los directorios usan lipgloss con color ANSI 16 (`"6"` = cyan).

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
- `github.com/charmbracelet/lipgloss` — estilos (solo para directorios)
