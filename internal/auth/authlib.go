package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AuthlibClient talks to a Yggdrasil-compatible Authlib Injector API.
type AuthlibClient struct {
	HTTP    *http.Client
	APIRoot string // e.g. https://example.com/api/yggdrasil
}

func NewAuthlibClient(apiRoot string) *AuthlibClient {
	return &AuthlibClient{
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		APIRoot: strings.TrimRight(strings.TrimSpace(apiRoot), "/"),
	}
}

type authenticateRequest struct {
	Agent       agent  `json:"agent"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	ClientToken string `json:"clientToken,omitempty"`
	RequestUser bool   `json:"requestUser"`
}

type agent struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type authenticateResponse struct {
	AccessToken string `json:"accessToken"`
	ClientToken string `json:"clientToken"`
	SelectedProfile *profile `json:"selectedProfile"`
	AvailableProfiles []profile `json:"availableProfiles"`
	Error       string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

type profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Login performs POST {api}/authserver/authenticate.
func (c *AuthlibClient) Login(ctx context.Context, username, password, clientToken string) (Session, error) {
	if c.APIRoot == "" {
		return Session{}, fmt.Errorf("Authlib Injector API 地址为空")
	}
	if username == "" || password == "" {
		return Session{}, fmt.Errorf("用户名或密码为空")
	}
	body := authenticateRequest{
		Agent:       agent{Name: "Minecraft", Version: 1},
		Username:    username,
		Password:    password,
		ClientToken: clientToken,
		RequestUser: true,
	}
	raw, _ := json.Marshal(body)
	url := c.APIRoot + "/authserver/authenticate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return Session{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mclauncher1-go/0.1")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Session{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Session{}, err
	}
	var ar authenticateResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		return Session{}, fmt.Errorf("解析登录响应失败: %w (HTTP %d)", err, resp.StatusCode)
	}
	if ar.Error != "" || ar.ErrorMessage != "" {
		msg := ar.ErrorMessage
		if msg == "" {
			msg = ar.Error
		}
		return Session{}, fmt.Errorf("登录失败: %s", msg)
	}
	if resp.StatusCode >= 400 {
		return Session{}, fmt.Errorf("登录失败: HTTP %d: %s", resp.StatusCode, string(data))
	}
	prof := ar.SelectedProfile
	if prof == nil && len(ar.AvailableProfiles) > 0 {
		prof = &ar.AvailableProfiles[0]
	}
	if prof == nil {
		return Session{}, fmt.Errorf("登录成功但无可用角色")
	}
	uuid := formatUUID(prof.ID)
	tok := ar.ClientToken
	if tok == "" {
		tok = clientToken
	}
	return Session{
		Mode:        ModeAuthlib,
		Username:    prof.Name,
		UUID:        uuid,
		AccessToken: ar.AccessToken,
		ClientToken: tok,
		UserType:    "mojang",
		AuthlibAPI:  c.APIRoot,
	}, nil
}

func formatUUID(id string) string {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) != 32 {
		return id
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", id[0:8], id[8:12], id[12:16], id[16:20], id[20:32])
}
