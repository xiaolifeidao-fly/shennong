package wechat

import (
	"common/middleware/vipper"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SessionResponse 是 jscode2session 的返回结构。
type SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Code2Session 用 wx.login 返回的临时 code 换取 openid / unionid。
// openid 是用户对当前小程序的唯一标识，无需用户额外授权即可获取。
func Code2Session(code string) (*SessionResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("微信登录 code 不能为空")
	}
	appID := strings.TrimSpace(vipper.GetString("wechat.appid"))
	secret := strings.TrimSpace(vipper.GetString("wechat.secret"))
	if appID == "" || secret == "" {
		return nil, fmt.Errorf("未配置 wechat.appid 或 wechat.secret")
	}

	values := url.Values{}
	values.Set("appid", appID)
	values.Set("secret", secret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Get("https://api.weixin.qq.com/sns/jscode2session?" + values.Encode())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var session SessionResponse
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		return nil, err
	}
	if session.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s", session.ErrMsg)
	}
	if strings.TrimSpace(session.OpenID) == "" {
		return nil, fmt.Errorf("微信未返回 openid")
	}
	return &session, nil
}
