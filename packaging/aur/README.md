# AUR package

`PKGBUILD` builds wisirc from the release tag with Go, following the Arch Go
packaging guidelines (PIE, trimpath, read-only modules), runs the tests and
installs the binary, man page, desktop entry, license and README.

For each release:

```sh
cd packaging/aur
sed -i "s/^pkgver=.*/pkgver=X.Y.Z/; s/^pkgrel=.*/pkgrel=1/" PKGBUILD
updpkgsums                         # replaces SKIP with the tarball's sha256
makepkg -si                        # build, test and install locally
makepkg --printsrcinfo > .SRCINFO  # metadata the AUR needs
```

Then copy `PKGBUILD` and `.SRCINFO` into the AUR git repository
(`ssh://aur@aur.archlinux.org/wisirc.git`) and push.
