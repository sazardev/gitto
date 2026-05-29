# SPEC.md - `gitto` (Zen Git TUI)

## 1. Visión General del Proyecto

`gitto` es un cliente TUI (Terminal User Interface) para Git enfocado en el minimalismo, la latencia cero y la modularidad. A diferencia de interfaces densas, `gitto` adopta una filosofía de "Vistas Enfocadas" (Zen Mode) y navegación inspirada en editores modales, reduciendo la carga cognitiva mediante el uso de espacio negativo y una paleta de comandos dinámica.

## 2. Stack Tecnológico y Dependencias Core

El proyecto está escrito puramente en **Go (Golang)**, priorizando la facilidad de distribución (binarios estáticos sin `cgo`), la concurrencia nativa y la arquitectura de estado predecible.

- **TUI Framework:** `github.com/charmbracelet/bubbletea` (Arquitectura Elm: Modelo, Vista, Actualizador).
- **Styling y Layout:** `github.com/charmbracelet/lipgloss` (para el diseño visual, bordes, márgenes y colores).
- **Git Backend:** `github.com/go-git/go-git/v5`. _Nota crítica:_ Es una implementación pura de Git en Go. No requiere `cgo` ni depender de `libgit2`. El estado debe consultarse a través de esta librería para garantizar rendimiento sin invocar `os/exec` salvo casos estrictamente necesarios.
- **Concurrencia:** Goroutines y Channels nativos, integrados con el sistema de comandos asíncronos (`tea.Cmd`) de Bubble Tea.
- **Configuración y Temas:** `github.com/spf13/viper` o `github.com/pelletier/go-toml/v2` (para leer de forma declarativa el archivo `.toml`).
- **Paleta de Comandos/Fuzzy Finder:** `github.com/charmbracelet/bubbles/list` y `github.com/charmbracelet/bubbles/textinput`.

## 3. Arquitectura del Sistema (Puertos y Adaptadores)

El código aplicará los principios de Arquitectura Hexagonal y Clean Architecture, aislando el dominio de la infraestructura y aprovechando las interfaces implícitas de Go:

- **Dominio (`internal/core/`):** Contiene el estado puro de la aplicación, entidades (ej. `Commit`, `Branch`, `FileStatus`) y casos de uso. No sabe nada de Bubble Tea, Lipgloss ni de la terminal.
- **Puertos (`internal/ports/`):** Interfaces que definen los contratos que el dominio necesita:
  - `GitProvider`: Interfaz para obtener el status, hacer commits, diffs, etc.
  - `ConfigProvider`: Interfaz para proveer la configuración del usuario y keybindings.
- **Adaptadores (`internal/adapters/`):**
  - `gogit_adapter`: Implementación de `GitProvider` utilizando `go-git/v5`.
  - `file_config`: Implementación de `ConfigProvider` utilizando `viper` para leer el archivo TOML.
- **Presentación (`internal/ui/`):** La capa de `bubbletea`. Contiene los modelos (`tea.Model`), la función `Update(tea.Msg)` y `View()`. Aquí se agrupan los componentes visuales (vistas) que consumen los casos de uso del dominio.

## 4. Diseño de la UI y Paradigmas de Interacción

- **Modularidad de Vistas:** Cada vista (Status, Diff, Log) es un sub-modelo que implementa la interfaz `tea.Model` o funciones similares (`Update`, `View`), permitiendo componer la interfaz principal delegando mensajes a los componentes activos.
- **Keybindings Nativos:** Movimiento estricto con `hjkl` mapeado a través de `key.NewBinding` del paquete `charmbracelet/bubbles/key`.
- **Paleta de Comandos (`:`):** Un modal superpuesto (overlay) que actúa como fuzzy finder para ejecutar comandos sin memorizar atajos complejos.
- **Vistas Enfocadas:** Solo se renderiza un modelo principal a la vez. Al presionar `<Enter>` sobre un archivo modificado, el modelo principal delega el renderizado al modelo `DiffView`, ocupando toda la pantalla o apareciendo como modal.

## 5. Flujo de Ejecución Principal (El Bucle de Bubble Tea)

1. **Init (`tea.Init`):** Se carga la configuración. Se detecta el repositorio mediante el adaptador de Git. Se dispara un `tea.Cmd` inicial para cargar el estado del repositorio de forma asíncrona.
2. **Event Loop:** Bubble Tea toma el control del `os.Stdin` y gestiona el bucle principal.
3. **Update (`tea.Update`):** Función pura que recibe mensajes (`tea.Msg`), como pulsaciones de teclas (`tea.KeyMsg`) o resultados asíncronos de Git. Retorna el nuevo estado del modelo y posibles nuevos comandos asíncronos (`tea.Cmd`).
4. **Concurrencia Segura:** Cualquier operación pesada (ej. `git push`) se lanza en una goroutine controlada que retorna un mensaje específico al terminar, manteniendo la UI completamente fluida.
5. **Render (`tea.View`):** Función pura que toma el modelo actual y retorna un `string` formateado con `lipgloss` para pintar en la terminal.

## 6. Comandos Principales (MVP - Alcance Inicial)

El enfoque de desarrollo inicial (versión 0.1.0) cubrirá exclusivamente estos flujos:

- **Status View (Vista por defecto):** Muestra archivos en `Staged`, `Unstaged` y `Untracked`.
- **Stage/Unstage (`s` / `u` o `Espacio`):** Mover archivos entre estados internamente en el índice del repositorio.
- **Commit View (`c`):** Abre un componente `textinput` flotante en la parte inferior para escribir el mensaje.
- **Push/Pull (`P` / `p`):** Operaciones remotas ejecutadas vía `tea.Cmd` con un componente `spinner` activo durante el proceso.
- **Log View (`l`):** Lista simple de `commits` recientes de la rama actual usando el componente `list` de Charm.
- **Diff View (Modal):** Renderizado de las diferencias del archivo seleccionado. Líneas `+` en verde (`lipgloss.Color`), líneas `-` en rojo.

## 7. Límites y Restricciones (Lo que NO cubre el MVP)

- No hay rebase interactivo gráfico.
- No hay resolución de conflictos de merge línea por línea en esta etapa.
- Ejecución enfocada a un solo repositorio a la vez por proceso.
