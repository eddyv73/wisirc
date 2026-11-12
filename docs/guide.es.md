<a id="top"></a>

# Guía de wisirc

Todo lo que hace wisirc, por tema. Volver al [README](../README.es.md) · [in English](guide.md) · `man wisirc` tiene la versión corta (en inglés).

[👋 La primera vez](#la-primera-vez) ·
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
[🗂️ Logs y menciones](#logs-y-menciones) ·
[⬆️ Actualizaciones](#actualizaciones)

<a id="la-primera-vez"></a>

## 👋 La primera vez

La primera vez que ejecutas wisirc (todavía no hay `config.toml`) se abre una configuración corta, con mouse o teclado:

1. **Idioma**: English o Español, elegido según tu sistema; la pantalla cambia mientras te mueves.
2. **Quién eres**: tu nick (sugerido a partir de tu usuario del sistema) y un nombre real opcional. Se guardan como tu identidad en `prefs.toml` y rellenan `/connect` y cada servidor nuevo.
3. **Tema**: cada tema se ve al pasar por la lista.
4. **Primer servidor**: importar tus servidores de Halloy (aparece primero si wisirc encuentra un config de Halloy), Libera.Chat entrando a `#wisirc`, tu propio servidor (abre el formulario con tu nick ya puesto) o ninguno.

**Enter** avanza, **Shift+Tab** regresa y **Esc** la salta sin guardar nada. No vuelve a aparecer: cambia lo que quieras después desde ≡ (`F1`).

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

Comandos: `/join`, `/part`, `/msg`, `/query`, `/me`, `/nick`, `/whois`, `/topic`, `/kick`, `/ban`, `/mode`, `/quit`, `/connect`, `/close`, `/quote`, `/tema`, `/lang`, `/update`, `/exportar`, `/help`. Para enviar un texto que empieza por `/`, escribe `//`.

<p align="right"><a href="#top">↑</a></p>

<a id="mouse"></a>

## 🖱️ Mouse

| Acción | Resultado |
|---|---|
| Clic en **≡** | Menú principal (conectar, servidores, unirse a un canal, logs, tema, idioma, config, ayuda, salir) |
| Clic en un canal de la barra | Cambia a ese canal |
| Clic en ▾ / ▸ | Colapsa o expande un servidor |
| Clic derecho en un servidor | Conectar/Desconectar, Unirse a canal…, Editar…, Activar/Desactivar, Color… |
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

≡ → **Servidores** lista todos los servidores con su estado (● conectado, ◌ conectando, ○ desconectado, ⊘ desactivado) y de dónde vienen:

- `servers.toml`: los guarda wisirc. Se crean, editan y borran desde esta pantalla con un formulario (nombre, host, puerto, TLS, certificado inválido, nick, usuario `usuario/red` para BNC, clave, SASL, canales y *Conectar al iniciar*).
- `config.toml`: tu archivo a mano, que wisirc nunca escribe. *Editar* abre el mismo formulario y guarda tus cambios en `servers.toml`; el servidor queda marcado con ✎ y **[Deshacer edición]** vuelve a la versión de `config.toml`. No se borran desde la pantalla, para no perder tus comentarios.
- `sin guardar`: conexiones hechas con `/connect`; se guardan con *Guardar*.

Botones: **[+ Nuevo]**, **[Editar]**, **[Conectar]** / **[Desconectar]**, **[Desactivar]** / **[Activar]**, **[Guardar]** (solo conexiones sin guardar), **[Borrar]** (**[Deshacer edición]** en los ✎), **[Cerrar]** y, si hay un config de Halloy, **[Importar de Halloy]**.

Teclas: ↑/↓ elegir · Enter editar · `n` nuevo · `c` conectar/desconectar · `x` desactivar/activar · `s` guardar · `d` borrar (pide confirmación) · `i` importar de Halloy · Esc cerrar. En el formulario: Tab/↓ siguiente campo, Espacio marca casillas, Enter guarda, Esc vuelve.

Si un nombre está en los dos archivos, gana `servers.toml`: así lo que editas aquí reemplaza a `config.toml`.

**Desactiva** un servidor que está caído o que no quieres usar por ahora: wisirc lo desconecta, deja de reintentar y no lo conecta al iniciar (aparece con ⊘ aquí y en el panel). La lista se guarda en `prefs.toml`, así que sirve para servidores de cualquiera de los dos archivos. **Activar**, o conectarlo a mano, lo vuelve a encender. `autoconnect = false` en el archivo hace lo mismo para un servidor, de forma fija.

> En macOS Terminal.app o iTerm2, activa "Usar Option como Meta" para que funcionen los atajos con `Alt`.

<p align="right"><a href="#top">↑</a></p>

<a id="venir-desde-halloy"></a>

## 📥 Venir desde Halloy

Si usas [Halloy](https://halloy.chat), wisirc encuentra su config (`~/Library/Application Support/halloy/config.toml` en macOS, `~/.config/halloy/config.toml` en Linux). En ≡ → **Servidores** aparece **[Importar de Halloy]** (o la tecla `i`): trae todos los servidores que aún no tienes, los guarda en `servers.toml` y los conecta. Los nombres que ya existen no se tocan.

Se importan nick, usuario, realname, host, puerto, TLS (en Halloy viene activado por defecto), certificados inválidos, canales, `password`, `password_file` y SASL PLAIN (también con `password_file`). Por seguridad `password_command` **no se ejecuta**; wisirc te avisa para que pongas esa clave a mano. `nick_password` y SASL EXTERNAL todavía no están soportados y también se avisan.

La primera vez que ejecutas wisirc, la [configuración inicial](#la-primera-vez) ofrece esta importación como primer servidor cuando encuentra un config de Halloy. Para arrancar directo desde un config de Halloy que está en otro lado, usa [`wisirc --config-copy`](#copiar-un-config).

Para ir en sentido contrario, `/exportar halloy` (o `/export halloy`) escribe tus servidores guardados como config de Halloy en `halloy-export.toml`, junto a tu config de wisirc, o en el archivo que indiques: `/exportar halloy ~/servidores-halloy.toml`. Incluye nick, usuario, realname, host, puerto, TLS, certificados inválidos, canales, la contraseña del servidor y SASL PLAIN; los servidores que desactivaste salen con `autoconnect = false`. Las conexiones hechas con `/connect` que nunca guardaste no se exportan. El archivo lleva tus contraseñas en texto plano, así que se crea legible solo para ti, y wisirc nunca sobrescribe un archivo que ya exista (tampoco el config de Halloy): copia tú sus tablas `[servers.*]` al `config.toml` de Halloy.

<p align="right"><a href="#top">↑</a></p>

<a id="archivo-de-configuración"></a>

## ⚙️ Archivo de configuración

El archivo está en `~/.config/wisirc/config.toml` (respeta `$XDG_CONFIG_HOME`). Para usar otro: `wisirc --config <ruta>`. La primera vez se crea sin servidores, con ejemplos comentados; los servidores que agregas van a `servers.toml`.

<a id="copiar-un-config"></a>

**Empezar con un config que ya tienes:** `wisirc --config-copy <ruta>` copia un config al lugar de wisirc (el de siempre o el que indiques con `--config`) y arranca con él, sin nada que configurar. `<ruta>` puede ser:

- una carpeta de config de wisirc, o su `config.toml`: se copian tal cual `config.toml`, `servers.toml` y `prefs.toml` (idioma, tema, identidad, colores, servidores desactivados). Sirve para llevar wisirc a otra máquina u otro usuario.
- un `config.toml` de Halloy, o la carpeta de config de Halloy: sus servidores se importan (con las mismas reglas que [Venir desde Halloy](#venir-desde-halloy)) a `servers.toml`, junto a un `config.toml` sin servidores.

```sh
wisirc --config-copy ~/respaldo/wisirc                  # una carpeta de config de wisirc
wisirc --config-copy ~/.config/halloy/config.toml       # tus servidores de Halloy
wisirc --config /tmp/prueba/config.toml --config-copy ~/.config/wisirc   # probar una copia sin tocar la tuya
```

Nunca reemplaza nada: si el destino ya tiene `config.toml`, `servers.toml` o `prefs.toml`, wisirc lo dice y no arranca; muévelo o elige otro con `--config`. Las copias quedan legibles solo para ti, porque pueden llevar contraseñas. Los logs no se copian. Una vez copiado, las siguientes veces basta con `wisirc`; como la configuración inicial se salta, el idioma, el tema y la identidad se cambian desde ≡.

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

Las claves que wisirc no conoce (por ejemplo, de un config de halloy) se ignoran y aparecen como aviso. wisirc nunca modifica tu config. Si lo cambias desde **≡ → Editar config**, se abre `$EDITOR` y al cerrarlo los cambios se aplican al momento: los servidores nuevos se conectan, los que cambiaron de dirección, puerto, TLS, nick o credenciales se reconectan con los datos nuevos, los que quitaste se desconectan, y `[theme]`, `[notifications]` y `[updates]` toman efecto (un tema elegido en el menú sigue mandando sobre `[theme]`). Volver a activar `[updates] check` después de arrancar con él apagado requiere reiniciar.

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

Cada tarjeta muestra el estado, el tiempo conectado, un gráfico animado de latencia (verde/amarillo/rojo), los canales y los mensajes sin leer y menciones; el borde usa el color del servidor en la barra lateral. Mientras el panel está abierto, la latencia se mide cada 5 segundos en vez de cada 30. Clic en una tarjeta (o Enter) va a ese servidor; `c` lo conecta o desconecta; `e` lo edita; `x` lo desactiva o activa; clic derecho abre el menú del servidor; las flechas mueven entre tarjetas; F2 o Esc lo cierra.

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

<a id="actualizaciones"></a>

## ⬆️ Actualizaciones

wisirc busca una versión nueva en GitHub al iniciar y luego una vez al día. Si la hay, la barra de estado muestra **⬆ v0.4.1 disponible · /update** y el menú ≡ suma **Actualizar a v0.4.1**. Haz clic en cualquiera de los dos, o escribe `/update` (alias `/actualizar`): wisirc descarga el binario de la versión para tu sistema, lo verifica con el `SHA256SUMS` de la release, comprueba que arranca y solo entonces reemplaza el binario en uso. La versión nueva se usa la próxima vez que abras wisirc, así no se cortan tus conexiones. `/update` sin nada anunciado busca en ese momento.

Solo se ofrecen versiones estables, nunca release candidates. Si wisirc se instaló con Homebrew no se reemplaza solo: te dice que ejecutes `brew upgrade wisirc`. Las compilaciones desde el código que no son una versión (`dev`) nunca reciben actualizaciones.

La búsqueda es una sola consulta a `api.github.com` y no envía nada sobre ti. Para apagarla, y que wisirc nunca contacte a GitHub:

```toml
[updates]
  check = false
```

<p align="right"><a href="#top">↑</a></p>
