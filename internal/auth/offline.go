package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// OfflineSession builds an offline (cracked) session.
// UUID follows Minecraft's OfflinePlayer UUID v3 (MD5 name-based) scheme.
func OfflineSession(username string) Session {
	name := strings.TrimSpace(username)
	if name == "" {
		name = "Player"
	}
	return Session{
		Mode:        ModeOffline,
		Username:    name,
		UUID:        OfflineUUID(name),
		AccessToken: "0",
		ClientToken: "0",
		UserType:    "legacy",
	}
}

// OfflineUUID returns the offline-mode UUID for a player name (no dashes in storage form with dashes for game).
func OfflineUUID(username string) string {
	// Java: UUID.nameUUIDFromBytes(("OfflinePlayer:" + name).getBytes(UTF_8))
	h := md5.Sum([]byte("OfflinePlayer:" + username))
	h[6] = (h[6] & 0x0f) | 0x30 // version 3
	h[8] = (h[8] & 0x3f) | 0x80 // IETF variant
	hexs := hex.EncodeToString(h[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexs[0:8], hexs[8:12], hexs[12:16], hexs[16:20], hexs[20:32])
}
