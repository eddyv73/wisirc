#!/usr/bin/env python3
"""A tiny fake IRC server, only for README screenshots.

It speaks just enough IRC for wisirc (CAP, NICK/USER, PING/PONG, JOIN,
NAMES, topics) and fills the channels with a scripted, friendly tech chat
plus a slow trickle of live messages. Each port is a different "network".

    python3 -I fakeircd.py            # listens on 127.0.0.1:16601-16603

Standard library only. Never exposed beyond 127.0.0.1.
"""
import asyncio
import datetime as dt
import random

HOST = "127.0.0.1"
B, C, I, R = "\x02", "\x03", "\x1d", "\x0f"   # bold, color, italic, reset
ACT = "\x01ACTION "

NETWORKS = {
    16601: {
        "name": "libera.fake",
        "lag": (0.035, 0.11),
        "channels": {
            "#go": {
                "topic": "The Go programming language · go1.27 is out! · paste at play.golang.org · be kind",
                "nicks": ["@gopherette", "@nullptr", "+byteknight", "+lambda_l", "segfault", "rackmount",
                          "pixelpine", "dotfiles", "quokka", "zephyr", "cobalt", "mossy", "vimtastic",
                          "ferrite", "owlbear", "tinygo_tim"],
                "backlog": [
                    ("join", "pixelpine"),
                    ("msg", "gopherette", "morning all ☕"),
                    ("msg", "byteknight", "o/ anyone tried the new iterator funcs in a real codebase yet?"),
                    ("msg", "nullptr", "yep, range-over-func cleaned up our pagination code a lot"),
                    ("msg", "nullptr", "for item := range client.Pages(ctx) { ... } just reads nicer"),
                    ("msg", "segfault", "the error handling inside the iterator is the tricky part though"),
                    ("msg", "lambda_l", "we yield (T, error) pairs, works fine with iter.Seq2"),
                    ("act", "rackmount", "is rebuilding the CI cache for the third time today"),
                    ("msg", "dotfiles", "rackmount: GOCACHE on a tmpfs helped us a ton"),
                    ("msg", "quokka", B + "PSA:" + B + " go vet now catches the loop-var capture we all hit in 2019 😅"),
                    ("msg", "gopherette", "kestrel: your bubbletea PR looked great btw, merged it this morning"),
                    ("msg", "zephyr", "is there a good way to profile allocations in a test?"),
                    ("msg", "byteknight", "zephyr: go test -bench . -benchmem -memprofile mem.out, then pprof -http :8080"),
                    ("msg", "zephyr", "oh nice, the flame graph view is exactly what I needed"),
                    ("part", "owlbear"),
                    ("msg", "cobalt", C + "03✔ build green" + R + " on " + C + "11linux/arm64" + R + ", " + C + "11darwin/arm64" + R + " and " + C + "11windows/amd64" + R),
                    ("msg", "mossy", "generics + iterators + slog… Go in 2026 is a comfy place"),
                    ("msg", "vimtastic", "still waiting for sum types ;)"),
                    ("msg", "nullptr", "vimtastic: aren't we all"),
                ],
                "live": [
                    ("msg", "ferrite", "TIL: errors.Join keeps every error, and errors.Is walks all of them"),
                    ("msg", "pixelpine", "anyone using sqlc with pgx v5? happy with it so far"),
                    ("msg", "lambda_l", "sqlc + goose migrations is our default stack now"),
                    ("msg", "segfault", "the race detector found a bug in my mutex wrapper, humbling"),
                    ("act", "quokka", "hands out virtual gopher stickers"),
                    ("msg", "byteknight", "kestrel: does wisirc support server-time? the scrollback looks right"),
                    ("msg", "dotfiles", "context.AfterFunc is underrated"),
                ],
            },
            "#linux": {
                "topic": "Linux help & chat · ask, don't ask to ask · pastebins welcome",
                "nicks": ["@kernelpanic", "+tux_fan", "zephyr", "mossy", "dotfiles", "btrfs_ben", "ziggy",
                          "nebula", "rackmount", "pixelpine"],
                "backlog": [
                    ("msg", "tux_fan", "upgraded to kernel 6.17, suspend finally works on this laptop"),
                    ("msg", "kernelpanic", "nice, which distro?"),
                    ("msg", "tux_fan", "fedora, with a hand-tuned zram setup"),
                    ("msg", "btrfs_ben", "snapper saved me again after a botched update 🙏"),
                    ("msg", "nebula", "systemd-run --user --scope is great for limiting a rogue build"),
                ],
                "live": [
                    ("msg", "ziggy", "pipewire + wireplumber has been rock solid lately"),
                    ("msg", "mossy", "journalctl -b -1 -p err is my favourite debugging command"),
                    ("msg", "kernelpanic", "remember to keep a live USB around, folks"),
                ],
            },
            "#homelab": {
                "topic": "Homelabs, racks & blinkenlights · post your setups",
                "nicks": ["@rackmount", "+blinkenlight", "cobalt", "ferrite", "nebula", "quokka", "owlbear", "ziggy"],
                "backlog": [
                    ("msg", "blinkenlight", "new mini PC arrived: 64 GB RAM, 2x NVMe, 9 W idle"),
                    ("msg", "rackmount", "proxmox or bare metal k3s?"),
                    ("msg", "blinkenlight", "proxmox, with a k3s VM for the fun stuff"),
                ],
                "live": [
                    ("msg", "cobalt", "my UPS just did its monthly self-test, everything survived"),
                    ("msg", "ferrite", "tailscale subnet router on the NAS is a game changer"),
                    ("msg", "owlbear", "finally labelled every cable in the rack"),
                ],
            },
        },
    },
    16602: {
        "name": "darkstar.lan",
        "lag": (0.012, 0.045),
        "channels": {
            "#lobby": {
                "topic": "darkstar · main hangout · weekly game night on fridays",
                "nicks": ["@alice", "@ana", "+bob", "nimbus", "pixelpine", "mossy", "quill", "ember", "juniper"],
                "backlog": [
                    ("msg", "alice", "who's in for game night?"),
                    ("msg", "ana", "me! bringing snacks"),
                    ("msg", "nimbus", "count me in, I'll set up the voice server"),
                ],
                "live": [
                    ("msg", "bob", "the new map pack is out"),
                    ("msg", "ember", "brb, coffee"),
                    ("msg", "quill", "kestrel: are you joining tonight?"),
                ],
            },
            "#code": {
                "topic": "dev talk · CI is green unless proven otherwise",
                "nicks": ["@alice", "+juniper", "quill", "nimbus", "ember"],
                "backlog": [
                    ("msg", "juniper", "pushed the new release tag, binaries are building"),
                    ("msg", "quill", "changelog looks good 👍"),
                ],
                "live": [
                    ("msg", "nimbus", "flaky test in the parser again, looking into it"),
                    ("msg", "juniper", "deployed to staging"),
                ],
            },
        },
    },
    16603: {
        "name": "znc.home",
        "lag": (0.11, 0.24),
        "channels": {
            "#selfhosted": {
                "topic": "Self-hosting everything · backups first, then fun",
                "nicks": ["@owlbear", "+nebula", "dotfiles", "btrfs_ben", "cobalt", "mossy", "ziggy", "quokka",
                          "ferrite", "lambda_l", "vimtastic", "tinygo_tim"],
                "backlog": [
                    ("msg", "owlbear", "restic + rclone to object storage, tested restores monthly"),
                    ("msg", "nebula", "the tested restores part is what everybody skips"),
                    ("msg", "dotfiles", "running my own matrix bridge to IRC now, works great"),
                ],
                "live": [
                    ("msg", "btrfs_ben", "caddy auto-TLS for internal services via DNS challenge, love it"),
                    ("msg", "ziggy", "uptime kuma dashboard is all green today"),
                    ("msg", "cobalt", "who else is running their own git forge?"),
                ],
            },
            "#irc": {
                "topic": "Everything IRC · clients, bouncers, IRCv3",
                "nicks": ["@nebula", "+lambda_l", "vimtastic", "quokka", "segfault", "pixelpine"],
                "backlog": [
                    ("msg", "lambda_l", "IRCv3 chathistory support is spreading nicely"),
                    ("msg", "vimtastic", "terminal clients with mouse support are underrated"),
                ],
                "live": [
                    ("msg", "quokka", "ZNC playback module + server-time = perfect scrollback"),
                    ("msg", "segfault", "SASL everywhere, please"),
                ],
            },
        },
        "query": ("nebula", "hey kestrel, the bouncer update went fine, thanks for the tip!"),
    },
}


def stamp(t):
    return t.astimezone(dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.") + "%03dZ" % (t.microsecond // 1000)


def bare(n):
    return n.lstrip("@+")


def mask(n):
    n = bare(n)
    return "%s!%s@%s.users.fake" % (n, n[:8], n)


class Conn:
    def __init__(self, net, reader, writer):
        self.net, self.r, self.w = net, reader, writer
        self.nick, self.user, self.cap, self.reg = None, False, False, False
        self.joined = []
        self.srv = net["name"]

    def send(self, line):
        self.w.write((line + "\r\n").encode())

    def num(self, code, *params, t=None):
        *mid, last = params
        tag = "@time=%s " % stamp(t) if t else ""
        self.send("%s:%s %s %s %s:%s" % (tag, self.srv, code, self.nick, "".join(p + " " for p in mid), last))

    def line(self, ev, ch, t=None):
        tag = "@time=%s " % stamp(t) if t else ""
        kind, who = ev[0], ev[1]
        if kind == "msg":
            self.send("%s:%s PRIVMSG %s :%s" % (tag, mask(who), ch, ev[2]))
        elif kind == "act":
            self.send("%s:%s PRIVMSG %s :%s%s\x01" % (tag, mask(who), ch, ACT, ev[2]))
        elif kind == "join":
            self.send("%s:%s JOIN %s" % (tag, mask(who), ch))
        elif kind == "part":
            self.send("%s:%s PART %s :see you later" % (tag, mask(who), ch))

    async def register(self):
        if self.reg or not (self.nick and self.user) or self.cap:
            return
        self.reg = True
        self.num("001", "Welcome to the %s IRC network, %s" % (self.srv, self.nick))
        self.num("002", "Your host is %s, running fakeircd" % self.srv)
        self.num("005", "CHANTYPES=# PREFIX=(ov)@+ NETWORK=%s" % self.srv.split(".")[0], "are supported by this server")
        self.num("375", "- %s message of the day" % self.srv)
        self.num("372", "- Be excellent to each other.")
        self.num("376", "End of /MOTD command.")
        q = self.net.get("query")
        if q:
            asyncio.get_running_loop().call_later(4, lambda: self.send(":%s PRIVMSG %s :%s" % (mask(q[0]), self.nick, q[1])))

    def join(self, ch):
        info = self.net["channels"].get(ch)
        if not info:
            self.send(":%s!%s@wisirc.users.fake JOIN %s" % (self.nick, self.nick, ch))
            self.num("353", "=", ch, self.nick)
            self.num("366", ch, "End of /NAMES list.")
            return
        self.joined.append(ch)
        back = info["backlog"]
        now = dt.datetime.now().astimezone()
        start = now - dt.timedelta(minutes=2 + len(back) * 1.4)
        # replay the JOIN with an old server-time so the backlog reads in order
        self.send("@time=%s :%s!%s@wisirc.users.fake JOIN %s" % (stamp(start), self.nick, self.nick, ch))
        self.num("332", ch, info["topic"], t=start)
        self.num("333", ch, "gopherette", str(int(dt.datetime.now().timestamp()) - 86400 * 3), t=start)
        self.num("353", "=", ch, " ".join(info["nicks"] + [self.nick]))
        self.num("366", ch, "End of /NAMES list.")
        t = start
        for ev in back:
            t += dt.timedelta(seconds=random.randint(25, 110))
            if t > now:
                t = now
            self.line(ev, ch, t)

    async def live(self):
        await asyncio.sleep(8)
        pools = []
        for ch, info in self.net["channels"].items():
            for ev in info.get("live", []):
                pools.append((ch, ev))
        random.shuffle(pools)
        for ch, ev in pools:   # each live line is sent once, so nothing repeats
            await asyncio.sleep(random.uniform(5, 12))
            if ch in self.joined:
                self.line(ev, ch)
                await self.w.drain()

    async def run(self):
        task = asyncio.create_task(self.live())
        try:
            while True:
                raw = await self.r.readline()
                if not raw:
                    break
                text = raw.decode(errors="replace").rstrip("\r\n")
                if text.startswith("@"):
                    text = text.split(" ", 1)[1]
                parts = text.split(" ")
                cmd = parts[0].upper()
                if cmd == "CAP":
                    sub = parts[1].upper() if len(parts) > 1 else ""
                    if sub == "LS":
                        self.cap = True
                        self.send(":%s CAP * LS :server-time multi-prefix" % self.srv)
                    elif sub == "REQ":
                        self.send(":%s CAP * ACK :%s" % (self.srv, text.split(":", 1)[1]))
                    elif sub == "END":
                        self.cap = False
                        await self.register()
                elif cmd == "NICK":
                    self.nick = parts[1].lstrip(":")
                    await self.register()
                elif cmd == "USER":
                    self.user = True
                    await self.register()
                elif cmd == "PING":
                    tok = text.split(" ", 1)[1].lstrip(":") if " " in text else ""
                    lo, hi = self.net.get("lag", (0.02, 0.12))
                    delay = random.uniform(lo, hi)
                    if random.random() < 0.15:
                        delay = min(0.3, delay * 2.5)   # an occasional spike
                    loop = asyncio.get_running_loop()
                    loop.call_later(delay, lambda tok=tok: self.send(":%s PONG %s :%s" % (self.srv, self.srv, tok)))
                elif cmd == "JOIN":
                    for ch in parts[1].lstrip(":").split(","):
                        self.join(ch)
                elif cmd == "PART":
                    ch = parts[1]
                    if ch in self.joined:
                        self.joined.remove(ch)
                    self.send(":%s!%s@wisirc.users.fake PART %s" % (self.nick, self.nick, ch))
                elif cmd == "QUIT":
                    break
                await self.w.drain()
        finally:
            task.cancel()
            self.w.close()


async def main():
    servers = []
    for port, net in NETWORKS.items():
        async def handler(r, w, net=net):
            await Conn(net, r, w).run()
        servers.append(await asyncio.start_server(handler, HOST, port))
        print("listening on %s:%d (%s)" % (HOST, port, net["name"]), flush=True)
    await asyncio.gather(*(s.serve_forever() for s in servers))


if __name__ == "__main__":
    asyncio.run(main())
