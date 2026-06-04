package app

import "testing"

func TestBundleExposesCMSModuleAndDefaultProfile(t *testing.T) {
	bundle := Bundle()
	if bundle.Manifest().ModuleID != "cms" {
		t.Fatalf("bundle module = %q, want cms", bundle.Manifest().ModuleID)
	}
	if bundle.Module().Manifest().ID != "cms" {
		t.Fatalf("module manifest = %q, want cms", bundle.Module().Manifest().ID)
	}
	profile := bundle.DefaultProfile()
	if profile.ID != "gocms-admin" || len(profile.Workspaces) != 1 {
		t.Fatalf("unexpected profile %#v", profile)
	}
	if got := profile.Workspaces[0].Modules; len(got) != 1 || got[0] != "cms" {
		t.Fatalf("profile modules = %v, want cms", got)
	}
}
