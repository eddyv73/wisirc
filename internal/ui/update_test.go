package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/eddyv73/wisirc/internal/update"
)

type fakeUpdater struct {
	latest  string
	applied []string
	err     error
}

func (f *fakeUpdater) Latest(context.Context) (string, error) { return f.latest, nil }
func (f *fakeUpdater) Apply(_ context.Context, tag string) (string, error) {
	f.applied = append(f.applied, tag)
	return "/home/u/.local/bin/wisirc", f.err
}

func (h *harness) withUpdater(version, latest string) *fakeUpdater {
	f := &fakeUpdater{latest: latest}
	h.m.d.Version, h.m.d.Updates, h.m.updater = version, f, f
	return f
}

func TestUpdateNoticeAndApply(t *testing.T) {
	h := newHarness(t)
	f := h.withUpdater("v0.4.0", "v0.4.1")
	h.send(h.m.checkUpdate(false)())
	if !strings.Contains(h.statusRow(), "v0.4.1") {
		t.Fatalf("no notice in the status bar: %q", h.statusRow())
	}
	if g := h.m.globalMenu(); g.Items[0].Action != ActUpdate {
		t.Fatalf("no Update item: %+v", g.Items[0])
	}
	cmd := h.m.startUpdate()
	if cmd == nil {
		t.Fatal("/update did nothing")
	}
	h.send(cmd())
	if len(f.applied) != 1 || f.applied[0] != "v0.4.1" {
		t.Fatalf("applied %v", f.applied)
	}
	if v := h.view(); !strings.Contains(v, "reopen wisirc") {
		t.Fatalf("no restart hint:\n%s", v)
	}
	if h.m.globalMenu().Items[0].Action == ActUpdate {
		t.Fatal("still offering the installed update")
	}
}

func TestUpdateUpToDate(t *testing.T) {
	h := newHarness(t)
	f := h.withUpdater("v0.4.1", "v0.4.1")
	h.send(h.m.checkUpdate(false)())
	if strings.Contains(h.statusRow(), "/update") {
		t.Fatal("notice for the same version")
	}
	h.send(h.m.startUpdate()()) // manual check
	if len(f.applied) != 0 || !strings.Contains(h.view(), "up to date") {
		t.Fatalf("applied %v:\n%s", f.applied, h.view())
	}
}

func TestUpdateManualInstallsRightAway(t *testing.T) {
	h := newHarness(t)
	f := h.withUpdater("v0.4.0", "v0.5.0")
	check := h.m.startUpdate()
	apply := h.m.updateChecked(check().(updateCheckedMsg))
	h.send(apply())
	if len(f.applied) != 1 {
		t.Fatalf("/update with nothing known did not install: %v", f.applied)
	}
}

func TestUpdateHomebrew(t *testing.T) {
	h := newHarness(t)
	f := h.withUpdater("v0.4.0", "v0.4.1")
	f.err = update.ErrHomebrew
	h.send(h.m.checkUpdate(false)())
	h.send(h.m.startUpdate()())
	if !strings.Contains(h.view(), "installed with Homebrew") || !strings.Contains(h.statusRow(), "brew upgrade wisirc") {
		t.Fatalf("no Homebrew hint:\n%s", h.view())
	}
}

func TestUpdateOff(t *testing.T) {
	h := newHarness(t)
	if h.m.checkUpdate(false) != nil {
		t.Fatal("checks without an updater")
	}
	if h.m.startUpdate() != nil || !strings.Contains(h.view(), "check = false") {
		t.Fatalf("/update with checks off:\n%s", h.view())
	}
}
