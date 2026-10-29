<a id="top"></a>

# Guía de wisirc

Todo lo que hace wisirc, por tema. Volver al [README](../README.es.md) · [in English](guide.md) · `man wisirc` tiene la versión corta (en inglés).

[⌨️ Atajos de teclado y comandos](#atajos-de-teclado-y-comandos) ·
[🖱️ Mouse](#mouse) ·
[🔌 Conectarse a un servidor](#conectarse-a-un-servidor) ·
[🗄️ Pantalla Servidores](#pantalla-servidores) ·
[📥 Venir desde Halloy](#venir-desde-halloy) ·
[⚙️ Archivo de configuración](#archivo-de-configuración) ·
[🎨 Temas](#temas) ·
[🌈 Colores por servidor](#colores-por-servidor) ·
[📊 Panel](#panel) ·
[🌎 Idioma](#idioma) ·
[🗂️ Logs y menciones](#logs-y-menciones)

<a id="atajos-de-teclado-y-comandos"></a>

## ⌨️ Atajos de teclado y comandos

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

<p align="right"><a href="#top">↑</a></p>

<a id="mouse"></a>

## 🖱️ Mouse

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

<p align="right"><a href="#top">↑</a></p>

<a id="conectarse-a-un-servidor"></a>

## 🔌 Conectarse a un servidor

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

<p align="right"><a href="#top">↑</a></p>

<a id="pantalla-servidores"></a>

## 🗄️ Pantalla Servidores

≡ → **Servidores** lista todos los servidores con su estado (● conectado, ◌ conectando, ○ desconectado) y de dónde vienen:

- `servers.toml`: los guarda wisirc. Se crean, editan y borran desde esta pantalla con un formulario (nombre, host, puerto, TLS, certificado inválido, nick, usuario `usuario/red` para BNC, clave, SASL, canales y *Conectar al iniciar*).
- `config.toml`: tu archivo a mano. *Editar* abre el editor; no se borran desde la pantalla, para no perder tus comentarios.
- `sin guardar`: conexiones hechas con `/connect`; se guardan con *Guardar*.

Botones: **[+ Nuevo]**, **[Editar]**, **[Conectar]** / **[Desconectar]**, **[Guardar]** (solo conexiones sin guardar), **[Borrar]**, **[Cerrar]** y, si hay un config de Halloy, **[Importar de Halloy]**.

Teclas: ↑/↓ elegir · Enter editar · `n` nuevo · `c` conectar/desconectar · `s` guardar · `d` borrar (pide confirmación) · `i` importar de Halloy · Esc cerrar. En el formulario: Tab/↓ siguiente campo, Espacio marca casillas, Enter guarda, Esc vuelve.

Si un nombre está en los dos archivos, gana `config.toml`. Con `autoconnect = false` un servidor queda guardado pero no se conecta al iniciar.

> En macOS Terminal.app o iTerm2, activa "Usar Option como Meta" para que funcionen los atajos con `Alt`.

<p align="right"><a href="#top">↑</a></p>

<a id="venir-desde-halloy"></a>

## 📥 Venir desde Halloy

Si usas [Halloy](https://halloy.chat), wisirc encuentra su config (`~/Library/Application Support/halloy/config.toml` en macOS, `~/.config/halloy/config.toml` en Linux). En ≡ → **Servidores** aparece **[Importar de Halloy]** (o la tecla `i`): trae todos los servidores que aún no tienes, los guarda en `servers.toml` y los conecta. Los nombres que ya existen no se tocan.

Se importan nick, usuario, realname, host, puerto, TLS (en Halloy viene activado por defecto), certificados inválidos, canales, `password`, `password_file` y SASL PLAIN (también con `password_file`). Por seguridad `password_command` **no se ejecuta**; wisirc te avisa para que pongas esa clave a mano. `nick_password` y SASL EXTERNAL todavía no están soportados y también se avisan.

La primera vez que abres wisirc sin config, si hay un Halloy en el equipo, arranca directamente en esa pantalla.

<p align="right"><a href="#top">↑</a></p>

<a id="archivo-de-configuración"></a>

## ⚙️ Archivo de configuración

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

<p align="right"><a href="#top">↑</a></p>

<a id="temas"></a>

## 🎨 Temas

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

<p align="right"><a href="#top">↑</a></p>

<a id="colores-por-servidor"></a>

## 🌈 Colores por servidor

Cada servidor tiene su color en la barra lateral: la fila seleccionada se rellena con él, y la ▾ del servidor y sus canales se escriben en ese tono, así cada grupo se distingue de un vistazo. Por defecto los servidores reciben tonos de verde en el orden de la barra. Clic derecho en un servidor → **Color…** para elegir otro (verde, verde oscuro, lima, jade, cian, azul, morado, rosa, rojo, naranja, ámbar, amarillo, gris) o **Automático**; las flechas lo muestran en vivo, Enter lo deja y Esc lo deshace. Se guarda por servidor en `prefs.toml`. Los canales con mensajes sin leer van en negrita del color del servidor; las menciones mantienen el color de alerta del tema.

<p align="right"><a href="#top">↑</a></p>

<a id="panel"></a>

## 📊 Panel

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

<p align="right"><a href="#top">↑</a></p>

<a id="idioma"></a>

## 🌎 Idioma

La interfaz está en inglés y en español. Por defecto sigue el idioma del sistema (`LC_ALL`, luego `LC_MESSAGES`, luego `LANG`): por ejemplo, `LANG=es_MX.UTF-8` la pone en español y cualquier otro valor, en inglés.

Para cambiarlo, abre ≡ → **Language / Idioma**: al moverte con ↑/↓ ves el idioma en vivo, Enter lo deja puesto y Esc vuelve al anterior. También puedes escribir `/lang en` o `/lang es` (alias `/idioma` y `/language`). La elección se guarda en `~/.config/wisirc/prefs.toml`.

<p align="right"><a href="#top">↑</a></p>

<a id="logs-y-menciones"></a>

## 🗂️ Logs y menciones

Se guarda un archivo por día y canal en `~/.local/share/wisirc/logs/<servidor>/<canal>/<AAAA-MM-DD>.log` (respeta `$XDG_DATA_HOME`). Al abrir un canal se cargan sus últimas 200 líneas como historial, y **≡ → Ver logs** muestra el registro completo del canal actual.

Cuando alguien escribe tu nick o te manda un privado, la línea se resalta y el canal se marca en la barra. Si no estás viendo ese canal, o la terminal no tiene el foco, también recibes una notificación del sistema: en macOS con `osascript` y en Linux con `notify-send`. Si no hay ninguno de los dos, suena la campana de la terminal.

<p align="right"><a href="#top">↑</a></p>
