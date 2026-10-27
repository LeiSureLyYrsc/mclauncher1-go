package auth

import "testing"

func TestOfflineUUID(t *testing.T) {
	// Known: OfflinePlayer:Player -> classic offline uuid
	u := OfflineUUID("Player")
	if len(u) != 36 {
		t.Fatalf("uuid len: %q", u)
	}
	s := OfflineSession("")
	if s.Username != "Player" || s.Mode != ModeOffline {
		t.Fatalf("%+v", s)
	}
	s2 := OfflineSession("Steve")
	if s2.UUID == s.UUID {
		t.Fatal("different names should differ")
	}
}
