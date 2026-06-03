# CDDuck

TUI file browser para reemplazar `cd` en la terminal. Navegá visualmente por el filesystem y salí al directorio elegido.

## Instalación

```bash
./install.sh
```

O manualmente:

```bash
go build -buildvcs=false -o cdduck .
mkdir -p ~/.local/bin
cp cdduck ~/.local/bin/
```

## Shell wrapper

El binario imprime el path a stdout y la TUI va a `/dev/tty`. Necesitás una función shell para que haga el `cd` real.

El `install.sh` la agrega automáticamente, o la podés copiar manualmente:

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

Agregá la función a `~/.bashrc`, `~/.zshrc` o `~/.config/fish/config.fish` y recargá con `source`.

## Uso

Tipeá `cdd` en la terminal. Se abre el browser fullscreen.

### Teclas

| Tecla | Acción |
|---|---|
| `↑` / `↓` | Navegar entre items |
| `PgUp` / `PgDn` | Avanzar / retroceder una página |
| `Home` / `End` | Ir al primer / último item |
| `Enter` | Entrar al directorio seleccionado |
| `Alt+Enter` | Seleccionar carpeta y salir |
| `Backspace` | Subir un nivel (filtro vacío) / borrar carácter (filtro activo) |
| `Esc` | Limpiar el filtro (si hay texto) o salir al directorio actual |
| `Ctrl+C` / `Ctrl+Q` | Salir sin seleccionar |

### Interfaz

```
╭──────  CDDuck  ──────╮
│                       │
│  Documents/           │
│  → proyectos/         │  ← item seleccionado con highlight azul
│  file.txt             │
│                       │
╰───────────────────────╯
/home/user              ← path actual
> filtro█                ← filtro fuzzy
```

- **Highlight**: la línea seleccionada se marca con fondo azul, texto blanco y negrita.
- **Path**: entre la caja y el filtro se muestra el directorio actual completo.
- **Directorio**: se listan primero las carpetas (en cyan), orden alfabético.
- **Filtro**: escribí para filtrar con fuzzy case-insensitive en vivo. `Esc` lo limpia.

Para navegar: entrá a una carpeta con `Enter` y cuando estés donde quieras, `Alt+Enter` para salir a ese directorio. También podés usar `Esc` con el filtro vacío para salir al directorio actual.

## Build desde fuente

Requiere Go 1.21+.

```bash
git clone ...
cd cdduck
go build -buildvcs=false -o cdduck .
```
