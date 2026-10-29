<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/brand/banner-dark.svg">
    <img src="docs/brand/banner-light.svg" alt="wisirc: IRC en tu terminal" width="760">
  </picture>
</p>

<p align="center">
  <a href="https://github.com/eddyv73/wisirc/releases/latest"><img src="https://img.shields.io/github/v/release/eddyv73/wisirc?style=flat-square&color=3fb950&label=release" alt="última versión"></a>
  <a href="https://github.com/eddyv73/wisirc/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/eddyv73/wisirc/ci.yml?branch=main&style=flat-square&label=tests" alt="pruebas"></a>
  <a href="https://github.com/eddyv73/wisirc/actions/workflows/install.yml"><img src="https://img.shields.io/github/actions/workflow/status/eddyv73/wisirc/install.yml?branch=main&style=flat-square&label=installer" alt="instalador"></a>
  <img src="https://img.shields.io/github/go-mod/go-version/eddyv73/wisirc?style=flat-square&color=4fc3d9" alt="versión de Go">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-6ea8ff?style=flat-square" alt="macOS y Linux">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/eddyv73/wisirc?style=flat-square&color=b38cff" alt="licencia MIT"></a>
</p>

<p align="center">
  <a href="#instalación">Instalación</a> ·
  <a href="#funciones">Funciones</a> ·
  <a href="#capturas">Capturas</a> ·
  <a href="#documentación">Documentación</a> ·
  <a href="README.md">English</a>
</p>

<p align="center">
  <img src="docs/screenshots/chat.png" alt="wisirc: servidores y canales a la izquierda, el chat al centro, los usuarios a la derecha y una barra de estado en vivo" width="900">
</p>

**wisirc** es un cliente IRC para la terminal que se siente como una app de escritorio. Los servidores y canales viven en una barra lateral, el chat queda al centro y los usuarios a la derecha. Todo funciona con el mouse (clics, rueda, menús contextuales) y todo funciona con el teclado, al estilo irssi. Un solo binario, sin dependencias y sin necesidad de configurar nada para empezar.

## Instalación

```sh
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh
```

Luego ejecuta `wisirc`. El instalador elige el binario para tu sistema (macOS o Linux, Intel o ARM), lo verifica con `SHA256SUMS`, lo deja en `~/.local/bin` y te avisa si esa carpeta no está en tu `PATH`. Ejecútalo otra vez para actualizar. Antes de que te llegue, cada versión se instala así en Linux y macOS de forma automática.

| Otras formas | Cómo |
|---|---|
| **Arch Linux / Garuda** | `makepkg -si` en [`packaging/aur`](packaging/aur), con man page y entrada en el menú (paquete en AUR próximamente) |
| **Go** | `go install github.com/eddyv73/wisirc/cmd/wisirc@latest` |
| **Binarios** | [último release](https://github.com/eddyv73/wisirc/releases/latest): `darwin-arm64`, `darwin-amd64`, `linux-amd64`, `linux-arm64` |
| **Desde el código** | `make build` (Go 1.27) |

<details>
<summary><b>Opciones del instalador</b></summary>

Las opciones van después de `sh -s --`:

```sh
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --version v0.3.1
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --add-to-path
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --source     # compilarlo tú (git + Go)
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --uninstall
```

| Opción / variable | Qué hace |
|---|---|
| `--version vX.Y.Z` | instala ese release en vez del último |
| `--source` | compila desde el código: este checkout o un clon nuevo |
| `--add-to-path` | agrega la carpeta de instalación a tu perfil (zsh, bash o fish), una sola vez |
| `--uninstall` | quita wisirc; tu configuración y tus logs se quedan |
| `WISIRC_PREFIX=/usr/local` | instala en otro lugar (`$WISIRC_PREFIX/bin`) |
| `WISIRC_REPO=dueño/nombre` | instala desde un fork |
| `GITHUB_TOKEN` | para un fork privado (o inicia sesión con `gh auth login`) |

Desde un clon, `./install.sh` acepta las mismas opciones.

</details>

<details>
<summary><b>Instalación manual y compilar desde el código</b></summary>

Descarga el binario para tu sistema desde `dist/` (o compílalo) y ponlo en tu `PATH`:

| Sistema | Binario |
|---|---|
| macOS Apple Silicon | `wisirc-darwin-arm64` |
| macOS Intel | `wisirc-darwin-amd64` |
| Linux x86-64 | `wisirc-linux-amd64` |
| Linux ARM64 | `wisirc-linux-arm64` |

```sh
install -m 0755 dist/wisirc-darwin-arm64 ~/.local/bin/wisirc
```

Compilar desde el código (requiere Go 1.27):

```sh
make build            # ./wisirc
make install          # copia a ~/.local/bin (PREFIX=/usr/local para otro destino)
make release VERSION=v0.1.0   # los 4 binarios en dist/
make test
```

</details>

## Funciones

<table>
<tr>
<td width="50%" valign="top">

**🖱️ Mouse y teclado, los dos completos**<br>
Clic en los canales, clic derecho en servidores, canales y nicks para sus menús, arrastra los bordes para cambiar el tamaño, selecciona texto para copiarlo. O no toques el mouse nunca.

</td>
<td width="50%" valign="top">

**📊 Panel en vivo**<br>
Una tarjeta por servidor con su estado, tiempo conectado, una gráfica animada de lag, los canales y los mensajes sin leer. Se abre al iniciar y con F2 cuando quieras.

</td>
</tr>
<tr>
<td valign="top">

**🎨 20 temas con vista previa**<br>
Dracula, Catppuccin, Tokyo Night, Rosé Pine, Kanagawa, monitores CRT verde y ámbar… todos revisados para que el texto se lea bien (WCAG AA). Cada servidor puede tener su propio color.

</td>
<td valign="top">

**🔐 Varios servidores, TLS y SASL**<br>
Conéctate a varias redes a la vez, con SASL PLAIN, contraseñas de servidor y de bouncer (ZNC) y certificados autofirmados cuando tú lo permites.

</td>
</tr>
<tr>
<td valign="top">

**🧭 Sin archivo de configuración**<br>
`/connect irc.libera.chat -nick tunick -join #linux` y guárdalo con un clic. Los servidores se administran desde una pantalla con formulario, y si usas Halloy importas todo en un clic.

</td>
<td valign="top">

**🌎 Español e inglés**<br>
Toda la interfaz sigue el idioma de tu sistema, y puedes cambiarlo en vivo desde el menú o con `/lang`.

</td>
</tr>
<tr>
<td valign="top">

**🔔 Menciones que te encuentran**<br>
Líneas resaltadas, canales marcados, un contador de menciones en la barra de estado y una notificación del sistema cuando no estás mirando.

</td>
<td valign="top">

**🗂️ Logs e historial**<br>
Un archivo por canal y por día; al abrir un canal vuelve su historial reciente, y el log completo está a un clic.

</td>
</tr>
</table>

## Primeros pasos

1. Ejecuta `wisirc`. Abre en el panel; presiona **Esc** para ir al chat.
2. Presiona **F1** (o haz clic en **≡**) para el menú principal: conectar, servidores, temas, idioma, ayuda.
3. Conéctate con `/connect irc.libera.chat -nick tunick -join #linux` y responde **Sí** cuando wisirc ofrezca guardarlo.
4. Cambia de canal con un clic o **Alt+1…9**; clic derecho en cualquier cosa para su menú.

## Capturas

<table>
<tr>
<td width="50%"><img src="docs/screenshots/dashboard.png" alt="Panel"><br><sub><b>Panel</b> (F2): gráficas de lag en vivo, tiempo conectado, canales y mensajes sin leer por servidor</sub></td>
<td width="50%"><img src="docs/screenshots/menu.png" alt="Menú principal"><br><sub><b>Menú ≡</b> (F1 o clic): todo a un clic</sub></td>
</tr>
<tr>
<td><img src="docs/screenshots/themes.png" alt="Selector de temas"><br><sub><b>Temas</b>: con las flechas se previsualizan en vivo (aquí catppuccin-mocha)</sub></td>
<td><img src="docs/screenshots/theme-papel.png" alt="Tema papel"><br><sub>El tema claro <b>papel</b></sub></td>
</tr>
<tr>
<td><img src="docs/screenshots/servers.png" alt="Pantalla de servidores"><br><sub><b>≡ → Servidores</b>: conectar, editar, borrar, importar de Halloy</sub></td>
<td><img src="docs/screenshots/server-form.png" alt="Formulario de servidor"><br><sub>Formulario de servidor, pensado para BNC (<code>usuario/red</code>)</sub></td>
</tr>
<tr>
<td><img src="docs/screenshots/connect.png" alt="Pantalla de conexión"><br><sub><code>/connect</code> con ejemplos listos para adaptar</sub></td>
<td><img src="docs/screenshots/spanish.png" alt="Interfaz en español"><br><sub>Interfaz en español (<code>/lang es</code>) y el menú de clic derecho de un nick</sub></td>
</tr>
</table>

Una grabación de 38 segundos está en [`docs/demo.cast`](docs/demo.cast): `asciinema play docs/demo.cast`.

## Documentación

`man wisirc` tiene la versión corta (en inglés). Todo lo demás, por tema:

<details>
<summary><b>⌨️ Atajos de teclado y comandos</b></summary>

| Tecla | Acción |
|---|---|
| `F1` | Menú principal |
| `Alt+1..9` | Ir al buffer N |
| `Ctrl+N` / `Ctrl+P` (o `Alt+→` / `Alt+←`) | Buffer siguiente / anterior |
| `Alt+A` | Siguiente buffer con actividad (primero las menciones) |
| `PgUp` / `PgDn` | Desplazar el historial |
| `↑` / `↓` | Historial de lo que escribiste |
| `Tab` | Completar nick |
| `Esc` | Cerrar menú o panel |
| `Ctrl+C` | Salir |

Comandos: `/join`, `/part`, `/msg`, `/query`, `/me`, `/nick`, `/whois`, `/topic`, `/kick`, `/ban`, `/mode`, `/quit`, `/connect`, `/close`, `/quote`, `/tema`, `/lang`, `/help`. Para enviar un texto que empieza por `/`, escribe `//`.

</details>

<details>
<summary><b>🖱️ Mouse</b></summary>

| Acción | Resultado |
|---|---|
| Clic en **≡** | Menú principal (conectar, servidores, unirse a un canal, logs, tema, idioma, config, ayuda, salir) |
| Clic en un canal de la barra | Cambia a ese canal |
| Clic en ▾ / ▸ | Colapsa o expande un servidor |
| Clic derecho en un servidor | Conectar/Desconectar, Unirse a canal…, Color… |
| Clic derecho en un canal | Salir (part), Cerrar, Marcar como leído, Ver topic |
| Clic (o clic derecho) en un nick | Whois, Mensaje privado, Op/Voice, Kick, Ban |
| Arrastrar sobre el chat o un panel de ayuda | Selecciona y copia al portapapeles (pbcopy, wl-copy/xclip/xsel; por SSH, OSC 52) |
| Rueda | Desplaza el historial, la barra o la nicklist |
| Arrastrar un borde `│` | Cambia el ancho de la barra o de la nicklist |

</details>

<details>
<summary><b>🔌 Conectarse a un servidor</b></summary>

`/connect` (o `/conectar`) sin nada, o ≡ → *Conectar servidor*, abre una pantalla con ejemplos. Sintaxis:

```
/connect <nombre>                          servidor guardado en el config
/connect <host> [puerto] [opciones]
```

| Opción | Para qué |
|---|---|
| `-nick <nick>` | tu nick |
| `-user <usuario>` | usuario (en un BNC/ZNC: `usuario/red`) |
| `-pass <clave>` | clave del servidor o del BNC |
| `-sasl <usuario:clave>` | login SASL |
| `-join <#a,#b>` | canales al conectar |
| `-tls` / `-notls` | forzar TLS (por defecto solo en el puerto 6697) |
| `-insecure` | aceptar certificados inválidos |
| `-name <nombre>` | nombre en la barra lateral |

Ejemplos:

```
/connect 192.0.2.20 7000 -notls -nick pepe -user pepe/libera -pass miClaveZnc
/connect irc.libera.chat -nick pepe -sasl pepe:miClave123 -join #linux
/connect irc.ejemplo.org 6667 -nick pepe -pass claveDelServidor
```

Cuando el servidor responde, wisirc pregunta **¿Guardar &lt;nombre&gt; permanentemente?** (S/N o clic). Si dices que sí, se guarda en `~/.config/wisirc/servers.toml` y se conecta solo cada vez que abres wisirc.

</details>

<details>
<summary><b>🗄️ Pantalla Servidores</b></summary>

≡ → **Servidores** lista todos los servidores con su estado (● conectado, ◌ conectando, ○ desconectado) y de dónde vienen:

- `servers.toml`: los guarda wisirc. Se crean, editan y borran desde esta pantalla con un formulario (nombre, host, puerto, TLS, certificado inválido, nick, usuario `usuario/red` para BNC, clave, SASL, canales y *Conectar al iniciar*).
- `config.toml`: tu archivo a mano. *Editar* abre el editor; no se borran desde la pantalla, para no perder tus comentarios.
- `sin guardar`: conexiones hechas con `/connect`; se guardan con *Guardar*.

Botones: **[+ Nuevo]**, **[Editar]**, **[Conectar]** / **[Desconectar]**, **[Guardar]** (solo conexiones sin guardar), **[Borrar]**, **[Cerrar]** y, si hay un config de Halloy, **[Importar de Halloy]**.

Teclas: ↑/↓ elegir · Enter editar · `n` nuevo · `c` conectar/desconectar · `s` guardar · `d` borrar (pide confirmación) · `i` importar de Halloy · Esc cerrar. En el formulario: Tab/↓ siguiente campo, Espacio marca casillas, Enter guarda, Esc vuelve.

Si un nombre está en los dos archivos, gana `config.toml`. Con `autoconnect = false` un servidor queda guardado pero no se conecta al iniciar.

> En macOS Terminal.app o iTerm2, activa "Usar Option como Meta" para que funcionen los atajos con `Alt`.

</details>

<details>
<summary><b>📥 Venir desde Halloy</b></summary>

Si usas [Halloy](https://halloy.chat), wisirc encuentra su config (`~/Library/Application Support/halloy/config.toml` en macOS, `~/.config/halloy/config.toml` en Linux). En ≡ → **Servidores** aparece **[Importar de Halloy]** (o la tecla `i`): trae todos los servidores que aún no tienes, los guarda en `servers.toml` y los conecta. Los nombres que ya existen no se tocan.

Se importan nick, usuario, realname, host, puerto, TLS (en Halloy viene activado por defecto), certificados inválidos, canales, `password`, `password_file` y SASL PLAIN (también con `password_file`). Por seguridad `password_command` **no se ejecuta**; wisirc te avisa para que pongas esa clave a mano. `nick_password` y SASL EXTERNAL todavía no están soportados y también se avisan.

La primera vez que abres wisirc sin config, si hay un Halloy en el equipo, arranca directamente en esa pantalla.

</details>

<details>
<summary><b>⚙️ Archivo de configuración</b></summary>

El archivo está en `~/.config/wisirc/config.toml` (respeta `$XDG_CONFIG_HOME`). Para usar otro: `wisirc --config <ruta>`. La primera vez que ejecutas wisirc se crea un ejemplo.

```toml
[servers.examplenet]
  nickname = "alice"
  username = "alice"
  server = "irc.example.net"
  port = 6697
  channels = ["#lobby"]
  dangerously_accept_invalid_certs = true   # solo para certificados autofirmados

  [servers.examplenet.sasl.plain]
  username = "alice"
  password = "hunter2"

[notifications]
  enabled = true

[theme]
  name = "oscuro"   # oscuro, claro, nocturno, clasico, alto-contraste, bosque, glaciar, papel, ...
```

Campos por servidor: `nickname`, `username`, `realname`, `server`, `port`, `use_tls` (por defecto activado en el puerto 6697), `channels`, `dangerously_accept_invalid_certs`, `password` (PASS del servidor), `sasl.plain` y `autoconnect`.

</details>

<details>
<summary><b>🎨 Temas</b></summary>

≡ → **Tema / colores** abre un submenú con los temas; el actual lleva ●. Al moverte con ↑/↓ el tema se aplica en vivo; Enter lo deja puesto, Esc vuelve al anterior. Con el mouse, un clic en el tema lo aplica. También `/tema <nombre>` (o `/theme`), y `/tema` solo abre el submenú.

Los nombres de los temas son identificadores y no cambian con el idioma (`dark` y `light` se aceptan como alias de `oscuro` y `claro`).

| Tema | Estilo |
|---|---|
| `oscuro` | colores de tu terminal (el de siempre) |
| `claro` | para terminales de fondo claro |
| `nocturno` | carbón con acento cian, estilo herdr |
| `clasico` | homenaje a irssi, barra azul |
| `alto-contraste` | negro y blanco puros, accesible |
| `bosque` | oscuro cálido |
| `glaciar` | oscuro frío |
| `papel` | claro tipo papel |
| `dracula` | oscuro violáceo con acentos neón |
| `catppuccin-mocha` | oscuro pastel, malva y azul |
| `catppuccin-latte` | claro pastel |
| `tokyo-night` | azul marino con acentos eléctricos |
| `one-dark` | gris carbón estilo Atom |
| `solarized-oscuro` | azul petróleo, tonos Solarized |
| `rose-pine` | púrpura nocturno, rosa y dorado |
| `kanagawa` | tinta japonesa, tonos cálidos |
| `everforest` | verde bosque suave |
| `monokai` | oliva oscuro, colores vivos |
| `fosforo` | monitor CRT verde |
| `ambar` | monitor CRT ámbar |

Los colores de los esquemas conocidos se aclararon donde hacía falta para que todo el texto tenga un contraste de al menos 4.5:1 (WCAG AA); hay un test que lo comprueba. Los temas con fondo propio cambian el fondo de la terminal mientras wisirc está abierto y lo restauran al salir. El tema elegido en el menú se guarda en `~/.config/wisirc/prefs.toml` y gana sobre `[theme]` del config.

Las claves que wisirc no conoce (por ejemplo, de un config de halloy) se ignoran y aparecen como aviso. wisirc nunca modifica tu config. Si lo cambias desde **≡ → Editar config**, se abre `$EDITOR` y al cerrarlo se recarga el archivo y se conectan los servidores nuevos.

> La contraseña se guarda en texto plano. Protege el archivo con `chmod 600 ~/.config/wisirc/config.toml`.

</details>

<details>
<summary><b>🌈 Colores por servidor</b></summary>

Cada servidor tiene su color en la barra lateral: la fila seleccionada se rellena con él, y la ▾ del servidor y sus canales se escriben en ese tono, así cada grupo se distingue de un vistazo. Por defecto los servidores reciben tonos de verde en el orden de la barra. Clic derecho en un servidor → **Color…** para elegir otro (verde, verde oscuro, lima, jade, cian, azul, morado, rosa, rojo, naranja, ámbar, amarillo, gris) o **Automático**; las flechas lo muestran en vivo, Enter lo deja y Esc lo deshace. Se guarda por servidor en `prefs.toml`. Los canales con mensajes sin leer van en negrita del color del servidor; las menciones mantienen el color de alerta del tema.

</details>

<details>
<summary><b>📊 Panel</b></summary>

wisirc arranca en el panel; Esc o F2 te lleva al chat. **F2** (o ≡ → **Panel**, o `/panel` / `/dashboard`) lo vuelve a abrir: una vista en vivo de todos tus servidores, una tarjeta por servidor:

```
 ● LIVE  servidores: 2 · conectados: 2 · promedio 120ms
 ╔═ znc ══════════════════════════╗ ╭─ libera ───────────────────────╮
 ║ ● conectado            ⏱ 2h 14m ║ │ ○ desconectado                 │
 ║ ▂▃▂▄▂▃▂  ● 42ms                ║ │ sin mediciones de latencia aún │
 ║ #lobby #code                   ║ │ sin canales                    │
 ║ 3 sin leer · 1 mención         ║ │ todo leído                     │
 ╚════════════════════════════════╝ ╰────────────────────────────────╯
```

Cada tarjeta muestra el estado, el tiempo conectado, un gráfico animado de latencia (verde/amarillo/rojo), los canales y los mensajes sin leer y menciones; el borde usa el color del servidor en la barra lateral. Mientras el panel está abierto, la latencia se mide cada 5 segundos en vez de cada 30. Clic en una tarjeta (o Enter) va a ese servidor; clic derecho (o `c`) lo conecta o desconecta; las flechas mueven entre tarjetas; F2 o Esc lo cierra.

</details>

<details>
<summary><b>🌎 Idioma</b></summary>

La interfaz está en inglés y en español. Por defecto sigue el idioma del sistema (`LC_ALL`, luego `LC_MESSAGES`, luego `LANG`): por ejemplo, `LANG=es_MX.UTF-8` la pone en español y cualquier otro valor, en inglés.

Para cambiarlo, abre ≡ → **Language / Idioma**: al moverte con ↑/↓ ves el idioma en vivo, Enter lo deja puesto y Esc vuelve al anterior. También puedes escribir `/lang en` o `/lang es` (alias `/idioma` y `/language`). La elección se guarda en `~/.config/wisirc/prefs.toml`.

</details>

<details>
<summary><b>🗂️ Logs y menciones</b></summary>

Se guarda un archivo por día y canal en `~/.local/share/wisirc/logs/<servidor>/<canal>/<AAAA-MM-DD>.log` (respeta `$XDG_DATA_HOME`). Al abrir un canal se cargan sus últimas 200 líneas como historial, y **≡ → Ver logs** muestra el registro completo del canal actual.

Cuando alguien escribe tu nick o te manda un privado, la línea se resalta y el canal se marca en la barra. Si no estás viendo ese canal, o la terminal no tiene el foco, también recibes una notificación del sistema: en macOS con `osascript` y en Linux con `notify-send`. Si no hay ninguno de los dos, suena la campana de la terminal.

</details>

## Contribuir

Los issues y pull requests son bienvenidos: lee [CONTRIBUTING.md](CONTRIBUTING.md). Lo que cambió en cada versión está en [CHANGELOG.md](CHANGELOG.md).

## Licencia

[MIT](LICENSE) © 2026 eddyv73
