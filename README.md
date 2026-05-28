# gitto

Un cliente de terminal (TUI) para Git enfocado en el minimalismo, la latencia cero y la modularidad.

gitto abandona las interfaces densas y saturadas en favor de un diseño limpio, vistas enfocadas (Zen Mode) y una paleta de comandos interactiva. Construido en Rust para garantizar un consumo de recursos minimo y un rendimiento nativo.

## Caracteristicas Principales

- Rendimiento Extremo: Escrito puramente en Rust utilizando libgit2. Sin recolector de basura, con un consumo de memoria optimizado.
- Filosofia Zen: Sistema de vistas modales. En lugar de dividir la pantalla en multiples paneles pequeños, gitto maximiza el espacio para lo que estas haciendo en este momento (revisar un diff, escribir un commit).
- Paleta de Comandos: Olvida la memorizacion de docenas de atajos. Presiona `:` para abrir un buscador difuso (fuzzy finder) visual con todas las acciones disponibles.
- Asincrono por Defecto: Las operaciones pesadas (fetch, push, pull) se ejecutan en segundo plano con indicadores visuales. La interfaz nunca se congela.
- Diseño Moderno y Modular: Soporte para temas base16, espaciado adaptable, componentes renderizados con doble buffer para evitar parpadeos y atajos de teclado inspirados en editores modales.

## Instalacion

Asegurate de tener Rust y Cargo instalados en tu sistema.

Clona el repositorio y compila la version de lanzamiento:

```bash
git clone [https://github.com/tu-usuario/gitto.git](https://github.com/tu-usuario/gitto.git)
cd gitto
cargo build --release

```

El binario compilado se encontrara en `target/release/gitto`. Puedes moverlo a una ruta en tu `$PATH` (por ejemplo, `~/.local/bin/`).

## Uso Rapido

Navega a cualquier directorio que sea un repositorio Git y ejecuta:

```bash
gitto

```

### Navegacion Basica

- `h`, `j`, `k`, `l` o Flechas: Navegar por listas y paneles.
- `Tab` / `Shift+Tab`: Cambiar el foco entre paneles activos.
- `Enter`: Expandir vista (ej. ver el Diff completo de un archivo seleccionado).
- `Esc`: Cerrar modales, cancelar busquedas o quitar selecciones.

### Atajos Principales

- `s`: Hacer stage/unstage del archivo o linea seleccionada.
- `c`: Abrir modal para escribir mensaje de commit.
- `P`: Push a la rama remota.
- `p`: Pull de la rama remota.
- `/`: Buscar dentro de la vista actual.
- `:`: Abrir la Paleta de Comandos.
- `?`: Abrir la capa de ayuda interactiva.
- `q`: Salir de la aplicacion.

## Configuracion

gitto es altamente personalizable. En el primer inicio, se generara un archivo de configuracion por defecto en `~/.config/gitto/config.toml`.

Desde este archivo puedes configurar:

- Temas y paletas de colores (fondos, colores de acento, alertas).
- Reasignacion de atajos de teclado.
- Disposicion inicial de los paneles.
- Comportamiento de comandos especificos.

## Arquitectura

gitto utiliza arquitectura hexagonal (Puertos y Adaptadores) para separar estrictamente el estado y la logica de Git de la capa de presentacion.

- Lenguaje: Rust
- TUI Framework: ratatui
- Eventos y Terminal: crossterm
- Git Backend: git2
- Concurrencia: tokio

## Contribucion

Las contribuciones son bienvenidas. Si deseas colaborar, por favor revisa los archivos `SPEC.md` y `DESIGN_SPEC.md` en la raiz del proyecto para entender los lineamientos arquitectonicos y de diseño antes de enviar un Pull Request.

## Licencia

Distribuido bajo la Licencia MIT.
