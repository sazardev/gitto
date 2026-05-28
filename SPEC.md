# SPEC.md - `gitto` (Zen Git TUI)

## 1. Visión General del Proyecto

`gitto` es un cliente TUI (Terminal User Interface) para Git enfocado en el minimalismo, la latencia cero y la modularidad. A diferencia de interfaces densas, `gitto` adopta una filosofía de "Vistas Enfocadas" (Zen Mode) y navegación inspirada en editores modales, reduciendo la carga cognitiva mediante el uso de espacio negativo y una paleta de comandos dinámica.

## 2. Stack Tecnológico y Dependencias Core

El proyecto está escrito en **Rust** priorizando el bajo consumo de memoria y la seguridad de concurrencia.

- **TUI Framework:** `ratatui` (sucesor de `tui-rs`).
- **Terminal Backend:** `crossterm` (manejo de eventos, teclado y renderizado cross-platform).
- **Git Backend:** `git2` (bindings de `libgit2`). _Nota crítica:_ No se deben hacer llamadas al sistema usando `std::process::Command` para operaciones de lectura; el estado debe consultarse vía `libgit2` para garantizar rendimiento nativo.
- **Concurrencia y Async:** `tokio` (runtime asíncrono para operaciones de red como fetch/push y futuras integraciones de APIs).
- **Configuración y Temas:** `serde`, `serde_derive`, `toml` (para configuración declarativa estática).
- **Errores:** `anyhow` y `thiserror`.

## 3. Arquitectura del Sistema (Puertos y Adaptadores)

El código debe estructurarse aislando estrictamente el dominio de la infraestructura:

- **Dominio (`core/`):** Contiene el estado puro de la aplicación, modelos (ej. `Commit`, `Branch`, `FileStatus`) y las reglas de negocio. No sabe nada de Ratatui ni de la terminal.
- **Puertos (`ports/`):** Traits (Interfaces) que definen qué necesita el dominio para funcionar.
  - `GitProvider`: Trait para obtener el status, hacer commits, etc.
  - `ConfigProvider`: Trait para leer atajos de teclado y temas.
- **Adaptadores (`adapters/`):**
  - `git2_adapter`: Implementación de `GitProvider` usando la librería `git2`.
  - `fs_config`: Implementación de `ConfigProvider` leyendo archivos `.toml`.
- **Presentación (`ui/`):** El motor de `ratatui`. Contiene el loop principal de eventos (`crossterm`), la gestión del estado visual y los componentes (Widgets). Consume los datos del dominio.

## 4. Diseño de la UI y Paradigmas de Interacción

- **Modularidad de Widgets:** Cada vista (Status, Diff, Log) es un componente aislado que implementa un trait común (ej. `Widget` o `Component`) con métodos `update(event)` y `render(frame)`.
- **Keybindings Nativos:** Movimiento estricto con `hjkl`. Atajos de teclado configurables.
- **Paleta de Comandos (`:`):** Un widget global superpuesto (overlay) que actúa como fuzzy finder para ejecutar comandos sin memorizar atajos complejos.
- **Vistas Enfocadas:** Solo se muestra un panel principal activo a la vez (ej. Árbol de archivos). Al presionar `<Enter>` sobre un archivo modificado, la pantalla transiciona o abre un modal a pantalla completa mostrando el Diff.

## 5. Flujo de Ejecución Principal (El Bucle)

1. **Init:** Se carga `gitto.toml`. Se detecta el repositorio Git actual usando `git2::Repository::discover`.
2. **State Hydration:** Se extrae asíncronamente el estado actual (archivos modificados, rama actual, commits ahead/behind).
3. **Event Loop:** Se bloquea esperando eventos de `crossterm::event`.
4. **Update:** Al recibir un `KeyEvent`, se muta el estado global o se dispara un comando de Git en un worker asíncrono (`tokio::spawn`).
5. **Render:** Ratatui redibuja los widgets en base al nuevo estado.

## 6. Comandos Principales (MVP - Alcance Inicial)

El agente debe concentrarse exclusivamente en implementar estos flujos para la versión 0.1.0:

- **Status View (Vista por defecto):** Muestra archivos en `Staged`, `Unstaged` y `Untracked`.
- **Stage/Unstage (`s` / `u` o `Espacio`):** Mover archivos entre estados. Soporte para stage de archivos completos.
- **Commit View (`c`):** Abre un input flotante en la parte inferior para escribir el mensaje.
- **Push/Pull (`P` / `p`):** Operaciones de sincronización remota ejecutadas de forma no bloqueante con un spinner de carga en el UI.
- **Log View (`l`):** Lista simple y limpia del historial de commits recientes de la rama actual.
- **Diff View (Modal):** Renderizado de las diferencias del archivo seleccionado, con colores rojo/verde para las líneas añadidas/eliminadas.

## 7. Límites y Restricciones (Lo que NO cubre el MVP)

- No hay rebase interactivo gráfico por el momento.
- No hay resolución de conflictos de merge línea por línea (se delega a la herramienta externa configurada).
- No hay soporte para múltiples repositorios a la vez en la misma sesión.
