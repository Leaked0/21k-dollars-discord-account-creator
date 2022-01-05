package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
	"woen_gen/bypass"
)

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, length)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func (c *DiscordClient) GetCookies() error {
	req, err := http.NewRequest("GET", "https://discord.com", nil)

	if err != nil {
		return err
	}

	for k, v := range map[string]string {
		"accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"accept-encoding": "gzip, deflate, br",
		"accept-language": "en-US,en;q=0.9",
		"sec-fetch-dest": "document",
		"sec-fetch-mode": "navigate",
		"sec-fetch-site": "none",
		"sec-fetch-user": "?1",
		"upgrade-insecure-requests": "1",
		"user-agent": UserAgent,
	} {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	cookies := resp.Cookies()
	cookie := ""

	for i, v := range cookies {
		cookie += fmt.Sprintf("%v=%v%v", v.Name, v.Value, func() string {
			if i == len(cookies) - 1 {
				return ""
			}

			return "; "
		}())
	}

	c.Cookie = cookie

	return nil
}

func (c *DiscordClient) GetFingerprint() error {
	req, err := http.NewRequest("GET", "https://discord.com/api/v9/experiments", nil)

	if err != nil {
		return err
	}

	for k, v := range map[string]string {
		"accept": "*/*",
		"accept-encoding": "hi",
		"accept-language": "en-US,en;q=0.9",
		"cookie": c.Cookie,
		"referer": "https://discord.com/",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
		"user-agent": UserAgent,
		"x-track": XTrack,
	} {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "fingerprint") {
		return errors.New("failed to get fingerprint")
	}

	c.Fingerprint = data["fingerprint"].(string)

	return nil
}

func (c *DiscordClient) GetConsent() (string, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v9/auth/location-metadata", nil)

	if err != nil {
		return "", err
	}

	for k, v := range map[string]string {
		"accept": "*/*",
		"accept-encoding": "hi",
		"accept-language": "en-US,en;q=0.9",
		"cookie": c.Cookie,
		"referer": "https://discord.com/",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
		"user-agent": UserAgent,
		"x-track": XTrack,
	} {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", err
	}

	if !strings.Contains(string(body), "consent_required") {
		return "true", nil
	}

	return fmt.Sprint(data["consent_required"]), nil
}

func (c *DiscordClient) Register(username string, proxy string, checkConsent bool, invite string) error {
	if c.Cookie == "" {
		err := c.GetCookies()

		if err != nil {
			return err
		}
	}

	if c.Fingerprint == "" {
		err := c.GetFingerprint()

		if err != nil {
			return err
		}
	}

	consent := "true"

	if checkConsent {
		co, err := c.GetConsent()

		if err != nil {
			return err
		}

		consent = co
	}

	ckey, err := bypass.SolveCaptcha(proxy)

	if err != nil {
		return err
	}

	payload := &RegisterPayload{
		Fingerprint: c.Fingerprint,
		Consent: consent == "true",
		Username: username,
		Captcha: ckey,
		Invite: invite,
	}

	p, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://discord.com/api/v9/auth/register", bytes.NewReader(p))

	if err != nil {
		return err
	}

	for k, v := range map[string]string {
		"accept": "*/*",
		"accept-encoding": "hi",
		"accept-language": "en-US,en;q=0.9",
		"content-type": "application/json",
		"cookie": c.Cookie,
		"origin": "https://discord.com",
		"referer": "https://discord.com/",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
		"user-agent": UserAgent,
		"x-fingerprint": c.Fingerprint,
		"x-debug-options": "bugReporterEnabled",
		"x-track": XTrack,
	} {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var data map[string]interface{}

	if !strings.Contains(string(body), "token") {
		return errors.New(string(body))
	}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return err
	}

	c.Token = data["token"].(string)

	c.Account = &Account{
		Username:      username,
		EmailVerified: false,
		PhoneVerified: false,
		Token: c.Token,
	}

	return nil
}

func (c *DiscordClient) ClaimAccount(mail *Mail) error {
	password := RandomString(12)

	for i := 0; i < 2; i++ {
		body := ""

		if i == 0 {
			body = `{"date_of_birth":"1981-11-12"}`
		} else {
			body = `{"email":"` + mail.Address + `","password":"` + password + `"}`
		}

		req, err := http.NewRequest("PATCH", "https://discord.com/api/v9/users/@me", strings.NewReader(body))

		if err != nil {
			return err
		}

		for k, v := range map[string]string {
			"accept": "*/*",
			"accept-encoding": "hi",
			"accept-language": "en-US",
			"authorization": c.Token,
			"content-type": "application/json",
			"cookie": c.Cookie,
			"origin": "https://discord.com",
			"referer": "https://discord.com/channels/@me",
			"sec-fetch-dest": "empty",
			"sec-fetch-mode": "cors",
			"sec-fetch-site": "same-origin",
			"user-agent": UserAgent,
			"x-debug-options": "bugReporterEnabled",
			"x-fingerprint": c.Fingerprint,
			"x-super-properties": SuperProperties,
		} {
			req.Header.Set(k, v)
		}

		resp, err := c.Client.Do(req)

		if err != nil {
			return err
		}

		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		if !strings.Contains(string(b), "token") {
			return errors.New(string(b))
		}

		var data map[string]interface{}

		err = json.Unmarshal(b, &data)

		if err != nil {
			return err
		}

		c.Token = data["token"].(string)

		if i == 0 {
			mail, err = kc.RequestMail(MailRequest{
				Site:     "discord.com",
				Type:     "mail.com,email.com,OUTLOOK",
				Sender:   "noreply@discord.com",
				Regex:    "",
				Investor: "0",
				NoSearch: "1",
				Subject:  "",
				Clear:    "1",
			})

			if err != nil {
				return err
			}

			c.Mail = mail
		}
	}

	c.Account.Email = mail.Address
	c.Account.Password = password

	return nil
}

func (c *DiscordClient) ConnectToWebsocket(proxy string) error {
	p, _ := url.Parse("http://" + proxy)

	dialer := websocket.Dialer{
		Proxy: http.ProxyURL(p),
	}

	ws, _, err := dialer.Dial("wss://gateway.discord.gg/?encoding=json&v=9&compress=zlib-stream", http.Header{
		"Origin": []string{"https://discord.com"},
		"User-Agent": []string{UserAgent},
	})

	if err != nil {
		return err
	}

	_, _, _ = ws.ReadMessage()

	message := `{"op":2,"d":{"token":"` + c.Token + `","capabilities":125,"properties":{"os":"Windows","browser":"Chrome","device":"","system_locale":"en-GB","browser_user_agent":"Mozilla/5.0 (X11; CrOS x86_64 10066.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36","browser_version":"96.0.4664.45","os_version":"","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":104785,"client_event_source":null},"presence":{"status":"online","since":0,"activities":[],"afk":false},"compress":false,"client_state":{"guild_hashes":{},"highest_last_message_id":"0","read_state_version":0,"user_guild_settings_version":-1,"user_settings_version":-1}}}`

	err = ws.WriteMessage(websocket.TextMessage, []byte(message))

	if err != nil {
		return err
	}

	_, _, _ = ws.ReadMessage()
	_, _, _ = ws.ReadMessage()

	ws.Close()

	return nil
}

func (c *DiscordClient) EmailVerify(mail *Mail, proxy string) error {
	defer func() {
		c.Account.EmailDone = true
	}()

	uri, err := kc.WaitForMessage(mail, 2 * time.Second, 30)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)

	if err != nil {
		return err
	}

	t := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := t.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	location, err := resp.Location()

	if err != nil {
		return err
	}

	token := strings.Split(location.String(), "token=")[1]

	for i := 0; i < 2; i++ {
		var b string

		if i == 0 {
			b = `{"token":"` + token + `","captcha_key":null}`
		} else {
			b = `{"token":"` + token + `","captcha_key":"` + bypass.SolveCaptchaForce(proxy) + `"}`
		}

		req, err := http.NewRequest(http.MethodPost, "https://discord.com/api/v9/auth/verify", bytes.NewReader([]byte(b)))

		if err != nil {
			return err
		}

		for k, v := range map[string]string {
			"accept": "*/*",
			"accept-encoding": "none",
			"accept-language": "en-GB",
			"content-type": "application/json",
			"cookie": c.Cookie,
			"origin": "https://discord.com",
			"referer": "https://discord.com/verify",
			"sec-fetch-dest": "empty",
			"sec-fetch-mode": "cors",
			"sec-fetch-site": "same-origin",
			"user-agent": UserAgent,
			"x-debug-options": "bugReporterEnabled",
			"x-fingerprint": c.Fingerprint,
			"x-super-properties": SuperProperties,
		} {
			req.Header.Set(k, v)
		}

		resp, err := c.Client.Do(req)

		if err != nil {
			return err
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		if resp.StatusCode == 200 {
			var data map[string]interface{}

			err = json.Unmarshal(body, &data)

			if err != nil {
				return err
			}

			c.Account.EmailVerified = true

			if strings.Contains(string(body), "token") {
				c.Token = data["token"].(string)
				c.Account.Token = c.Token

				return nil
			} else {
				return nil
			}
		}

		if !strings.Contains(string(body), "captcha") {
			return errors.New("email verify: " + string(body))
		}
	}

	return errors.New("failed to email verify")
}

func (c *DiscordClient) PhoneVerify(Buy func() (*PhonePurchase, error), Check func(purchase *PhonePurchase) (*PhoneOrder, error), Change func(order *PhoneOrder, status string) error) error {
	defer func() {
		c.Account.PhoneDone = true
	}()

	order, err := Buy()

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://discord.com/api/v9/users/@me/phone", strings.NewReader(fmt.Sprintf(`{"phone":"%v"}`, order.Phone)))

	if err != nil {
		return err
	}

	for k, v := range map[string]string {
		"accept": "*/*",
		"accept-encoding": "none",
		"accept-language": "en-GB",
		"authorization": c.Token,
		"content-type": "application/json",
		"cookie": c.Cookie,
		"origin": "https://discord.com",
		"referer": "https://discord.com/channels/@me",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
		"user-agent": UserAgent,
		"x-debug-options": "bugReporterEnabled",
		"x-super-properties": SuperProperties,
	} {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		if strings.Contains(string(body), "Invalid") || strings.Contains(string(body), "VoIP") {
			return c.PhoneVerify(Buy, Check, Change)
		}

		return errors.New(string(body))
	}

	var o *PhoneOrder

	for i := 0; i < 45; i++ {
		o, err = Check(order)

		if err != nil {
			return err
		}

		if o.Status == "READY" {
			break
		}

		time.Sleep(2 * time.Second)
	}

	if o.Code == "" {
		err = Change(o, "8")

		if err != nil {
			return err
		}

		if c.Attempted2 == false {
			if c.Attempted {
				c.Attempted2 = true
			} else {
				c.Attempted = true
			}

			return c.PhoneVerify(Buy, Check, Change)
		} else {
			return errors.New("failed to get code")
		}
	}

	req, err = http.NewRequest("POST", "https://discord.com/api/v9/phone-verifications/verify", strings.NewReader(fmt.Sprintf(`{"phone":"%v","code":"%v"}`, order.Phone, o.Code)))

	if err != nil {
		return err
	}

	for k, v := range map[string]string {
		"accept": "*/*",
		"accept-encoding": "none",
		"accept-language": "en-GB",
		"authorization": c.Token,
		"content-type": "application/json",
		"cookie": c.Cookie,
		"origin": "https://discord.com",
		"referer": "https://discord.com/channels/@me",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
		"user-agent": UserAgent,
		"x-debug-options": "bugReporterEnabled",
		"x-super-properties": SuperProperties,
	} {
		req.Header.Set(k, v)
	}

	resp, err = c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "token") {
		return errors.New(string(body))
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return err
	}

	token := data["token"].(string)

	req, err = http.NewRequest("POST", "https://discord.com/api/v9/users/@me/phone", strings.NewReader(fmt.Sprintf(`{"phone_token":"%v","password":"%v"}`, token, c.Account.Password)))

	if err != nil {
		return err
	}

	for k, v := range map[string]string {
		"accept": "*/*",
		"accept-encoding": "none",
		"accept-language": "en-GB",
		"authorization": c.Token,
		"content-type": "application/json",
		"cookie": c.Cookie,
		"origin": "https://discord.com",
		"referer": "https://discord.com/channels/@me",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
		"user-agent": UserAgent,
		"x-debug-options": "bugReporterEnabled",
		"x-super-properties": SuperProperties,
	} {
		req.Header.Set(k, v)
	}

	resp, err = c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		return errors.New(string(body))
	}

	c.Account.PhoneVerified = true

	_ = Change(o, "6")

	return nil
}

func (c *DiscordClient) AddPFP(avatar string) error {
	req, err := http.NewRequest("PATCH", "https://discord.com/api/v9/users/@me", strings.NewReader(fmt.Sprintf(`{"avatar":"%v"}`, avatar)))

	if err != nil {
		return err
	}

	for k, v := range map[string]string {
		"accept": "*/*",
		"accept-encoding": "none",
		"accept-language": "en-GB",
		"authorization": c.Token,
		"content-type": "application/json",
		"cookie": c.Cookie,
		"origin": "https://discord.com",
		"referer": "https://discord.com/channels/@me",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
		"user-agent": UserAgent,
		"x-debug-options": "bugReporterEnabled",
		"x-fingerprint": c.Fingerprint,
		"x-super-properties": SuperProperties,
	} {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
