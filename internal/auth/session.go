package auth

// Mode constants.
const (
	ModeOffline = "offline"
	ModeAuthlib = "authlib"
	// ModeMicrosoft reserved — not implemented yet.
	ModeMicrosoft = "microsoft"
)

// Session holds credentials passed to the game process.
type Session struct {
	Mode        string
	Username    string
	UUID        string
	AccessToken string
	ClientToken string
	UserType    string // "legacy" / "mojang" / "msa"
	AuthlibAPI  string // yggdrasil API URL for -javaagent
}

// DisplayName returns the in-game player name.
func (s Session) DisplayName() string {
	if s.Username != "" {
		return s.Username
	}
	return "Player"
}
