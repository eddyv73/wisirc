export GO111MODULE := on
export CGO_ENABLED := 0

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
PREFIX ?= $(HOME)/.local
LDFLAGS := -s -w -X main.Version=$(VERSION)
TARGETS := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64

.PHONY: build test release install clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o wisirc ./cmd/wisirc

test:
	go test ./...

release:
	@mkdir -p dist
	@for t in $(TARGETS); do \
		os=$${t%/*}; arch=$${t#*/}; \
		echo "-> dist/wisirc-$$os-$$arch"; \
		GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" -o dist/wisirc-$$os-$$arch ./cmd/wisirc || exit 1; \
	done

# PREFIX=/usr DESTDIR=$pkgdir for packages; the default installs for the user
install: build
	install -Dm0755 wisirc $(DESTDIR)$(PREFIX)/bin/wisirc
	install -Dm0644 docs/wisirc.1 $(DESTDIR)$(PREFIX)/share/man/man1/wisirc.1
	install -Dm0644 packaging/wisirc.desktop $(DESTDIR)$(PREFIX)/share/applications/wisirc.desktop
	install -Dm0644 LICENSE $(DESTDIR)$(PREFIX)/share/licenses/wisirc/LICENSE

clean:
	rm -rf dist wisirc
