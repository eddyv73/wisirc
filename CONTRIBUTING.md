# Contributing to wisirc

Thanks for helping! wisirc is a small Go program; everything needed is in
this repository.

## Build and test

```sh
make build      # ./wisirc
make test       # go test ./...
go test -race ./...
gofmt -l .      # must print nothing
```

Go 1.27 or newer. The UI is built with Bubble Tea v2 and Lip Gloss v2.

## Ground rules

- **Tests first.** Every change comes with a test that fails without it.
  UI behavior is tested through `internal/ui`'s harness (keys, mouse clicks,
  rendered view), not by hand.
- **Every user-visible string is translated.** Write it in English inside
  `i18n.T("...")` (a string literal) and add the Spanish text to a catalog in
  `internal/i18n/es_*.go`. `TestCatalogsComplete` fails on missing or unused
  entries. Adding a language is one more catalog file.
- **Nothing from the network reaches the terminal raw.** Text from servers
  goes through `RenderIRC` / `sanitize`, which drop control characters.
- **Bounded memory.** Keep the existing limits (line length, lines per
  buffer, private windows) when touching the IRC client or the state.
- **Example data only.** Use `alice`, `bob`, `irc.example.net` and the
  `192.0.2.0/24` documentation range in tests, docs and screenshots — never
  real servers, addresses or credentials.

## Pull requests

Keep them focused, describe what a user will notice, and make sure CI is
green (gofmt, vet, `go test -race` on Linux and macOS).
