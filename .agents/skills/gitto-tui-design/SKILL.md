# Gitto TUI Design System: Modern, Visual & Modular

## Core Philosophy

Las interfaces de terminal tradicionales asumen que el usuario prefiere la densidad de texto sobre la estética. Este sistema rechaza esa noción. El diseño debe sentirse como una aplicación web moderna o un editor de código de última generación, viviendo dentro de la terminal.

**Pilares del Diseño:**

1. **Visualmente Explicativo:** El usuario nunca debe adivinar qué está pasando. Las animaciones, los estados de carga y las ayudas visuales son ciudadanos de primera clase.
2. **Modular y Acomodable:** La interfaz responde a las necesidades del usuario, permitiendo layouts dinámicos y personalizables similares a un _tiling window manager_.
3. **Poder Amigable:** Atajos potentes para el veterano, pero con una capa de descubrimiento visual e interactiva para el principiante.
4. **Espacio Negativo (Zen):** La interfaz respira. Se prefieren los márgenes y el _padding_ sobre las líneas divisorias duras.

---

## 1. El Paradigma de Layout: "Tiling Modular"

La interfaz abandona el concepto de paneles fijos e inamovibles. El layout base es limpio, pero altamente personalizable mediante configuración.

| Tipo de Vista          | Comportamiento                                                                  | Uso Principal                                     |
| :--------------------- | :------------------------------------------------------------------------------ | :------------------------------------------------ |
| **Zen Mode (Default)** | Un solo panel central dominando el 80% de la pantalla. Márgenes amplios.        | Vista de Estado (Status), Diff enfocado.          |
| **Split Dinámico**     | División de pantalla (Vertical/Horizontal) invocada por el usuario.             | Comparar el Diff mientras se navega el historial. |
| **Overlay (Flotante)** | Ventanas emergentes centradas con efecto de desenfoque/oscurecimiento de fondo. | Paleta de Comandos, Alertas, Input de Commits.    |

- **Regla de Acomodo:** El usuario puede definir en su `config.toml` dónde quiere el panel de ramas (ej. `branches_position = "left" | "right" | "floating"`).
- **Bordes Modernos:** Uso estricto de bordes redondeados (`╭`, `╰`, `╮`, `╯`). Los paneles inactivos tienen bordes atenuados; el panel activo usa el color de acento y brilla sutilmente.

---

## 2. El Corazón de la Navegación: La Paleta de Comandos

Para evitar que el usuario memorice 50 atajos de teclado, la herramienta se centra en una Paleta de Comandos altamente visual.

- **Invocación:** Se abre instantáneamente con `:` o `Ctrl+P`.
- **Estética:** Un modal flotante en el centro superior de la pantalla.
- **Fuzzy Search en Tiempo Real:** Filtra acciones, ramas y archivos mientras se escribe, resaltando las coincidencias con el color de acento.
- **Diseño Explicativo:** Cada comando en la paleta muestra a la derecha su atajo de teclado correspondiente (si lo tiene) y un ícono descriptivo.

> **Ejemplo Visual de la Paleta:**
> `[🔍 ] ch...`
> `❯ 🌿 checkout branch    [c]`
> `↩️ cherry-pick commit [C]`

---

## 3. Feedback Visual y Animaciones

Una TUI moderna debe sentirse viva. Ninguna operación que tarde más de 50ms debe ocurrir sin feedback visual.

### Transiciones y Estados de Carga

- **Skeleton Loaders:** En lugar de dejar la pantalla en blanco mientras se lee un repositorio grande o se hace un _fetch_, mostrar un esqueleto del layout con texto atenuado (`██████░░░`).
- **Spinners Explicativos:** Todo proceso asíncrono (push, pull, rebase) muestra un spinner animado a 60fps en la barra de estado inferior, acompañado de texto claro: `⠧ Sincronizando con origin/main...`
- **Micro-Animaciones:** Al hacer _stage_ de un archivo, este no debe desaparecer de golpe; el color debe parpadear en verde (éxito) antes de moverse a la sección de _Staged_.

### Sistema de Notificaciones (Toasts)

Alertas flotantes no bloqueantes en la esquina inferior derecha.

- **Éxito (Verde):** `✓ Commit "feat: UI" creado.` (Desaparece en 3s).
- **Error (Rojo):** `✗ Conflicto en main.rs. Requiere resolución manual.` (Requiere presionar `Esc`).

---

## 4. Teclado, Accesibilidad y Ayuda Continua

El diseño asume que el usuario es inteligente, pero puede olvidar cosas. La curva de aprendizaje debe ser plana gracias a la interfaz.

### El Footer Contextual (Bottom Bar)

Una barra dinámica en la parte inferior que cambia según el panel enfocado, mostrando _solamente_ lo que se puede hacer en ese instante.

`[Espacio] Stage  |  [c] Commit  |  [:] Comandos  |  [?] Ayuda Visual`

### El Sistema de Ayuda Interactivo (`?`)

Al presionar `?`, en lugar de mostrar un bloque de texto plano, la pantalla actual se oscurece y aparecen "Tooltips" o flechas apuntando a los paneles actuales explicando qué hace cada tecla en ese contexto específico.

### Navegación Universal

- `hjkl` / Flechas: Movimiento fluido.
- `Tab`: Salto entre paneles visibles.
- `Esc`: Botón de pánico universal. Cierra modales, cancela búsquedas, quita selecciones.

---

## 5. Tipografía, Iconografía y Temas

La TUI depende en gran medida del uso correcto de fuentes Nerd Fonts para romper la monotonía del texto plano.

| Elemento               | Regla de Diseño                                                                                                             |
| :--------------------- | :-------------------------------------------------------------------------------------------------------------------------- |
| **Íconos de Archivos** | Cada archivo muestra el ícono de su extensión (ej. 🦀 para `.rs`, 🐹 para `.go`).                                           |
| **Padding**            | Todo panel debe tener al menos un espacio en blanco (1 celda) de separación entre su borde y el texto.                      |
| **Highlights**         | Al seleccionar una línea en un Diff, no solo se cambia el fondo, la línea entera se pone en **Bold** para máximo contraste. |

### Paleta de Colores Semántica y "Aesthetic"

Se debe evitar los colores primarios chillones de la terminal ANSI clásica. Los temas por defecto deben basarse en paletas pastel y modernas.

- `bg_base`: #1E1E2E (Fondo oscuro, sin contraste extremo).
- `accent_primary`: #CBA6F7 (Púrpura/Lavanda para indicar foco de navegación).
- `success`: #A6E3A1 (Verde pastel para archivos _staged_ o líneas añadidas).
- `danger`: #F38BA8 (Rojo suave para archivos borrados o errores).
- `muted`: #6C7086 (Gris oscuro para metadata, atajos de teclado inactivos).

---

## 6. Rendimiento Gráfico (Anti-Flicker)

Para soportar las animaciones, el TUI debe implementar un renderizado estricto:

1. **Buffer Doble:** Nunca borrar la pantalla; sobreescribir solo las celdas que cambiaron.
2. **60 FPS Limit:** Limitar el refresco de animaciones (spinners) para no consumir más del 5% de la CPU en reposo.
3. **Renderizado Asíncrono:** La capa de UI y la capa de peticiones a `libgit2` deben vivir en hilos (threads) separados para garantizar que el cursor y las animaciones nunca se congelen.
