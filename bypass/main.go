package bypass

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

const (
	Base      = "https://hcaptcha.com"
	Sec       = "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"96\", \"Google Chrome\";v=\"96\""
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:94.0) Gecko/20100101 Firefox/94.0"
)

func Setup() {
	_ = exec.Command("node", "bypass\\browser\\index.js").Run()
}

func MakeClient(proxy string) *http.Client {
	if !strings.HasPrefix(proxy, "http://") {
		proxy = "http://" + proxy
	}

	p, _ := url.Parse(proxy)

	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			MaxVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				0x0a0a, 0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030,
				0xcca9, 0xcca8, 0xc013, 0xc014, 0x009c, 0x009d, 0x002f, 0x0035,
			},
			CurvePreferences: []tls.CurveID{
				tls.CurveID(0x0a0a),
				tls.X25519,
				tls.CurveP256,
				tls.CurveP384,
			},
			InsecureSkipVerify: true,
		},
	}

	if proxy != "" && proxy != "http://" {
		t.Proxy = http.ProxyURL(p)
	}

	return &http.Client{
		Transport: t,
		Timeout:   60 * time.Second,
	}
}

func CheckSiteConfig(client *http.Client, host string, sitekey string) (*SiteConfig, error) {
	req, err := http.NewRequest("GET", Base+"/checksiteconfig?v=e7ef6a8&host="+host+"&sitekey="+sitekey+"&sc=1&swa=1", nil)

	if err != nil {
		return nil, err
	}

	for k, v := range map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "en-GB,en;q=0.5",
		"Cache-Control":   "no-cache",
		"Connection":      "keep-alive",
		"Content-Type":    "application/json; charset=utf-8",
		"Origin":          "https://newassets.hcaptcha.com",
		"Referer":         "https://newassets.hcaptcha.com/",
		"sec-fetch-dest":  "empty",
		"sec-fetch-mode":  "cors",
		"sec-fetch-site":  "same-site",
		"TE":              "trailers",
		"user-agent":      UserAgent,
	} {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	data := &SiteConfig{}

	err = json.Unmarshal(body, data)

	if err != nil {
		return nil, err
	}

	return data, err
}

func SolveN(req string) (string, error) {
	resp, err := http.Get("http://193.34.166.182:3030/n?req=" + req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

func m() string {
	return `{"st":1638833457976,"mm":[[129,20,1638833467772],[129,20,1638833467814]],"mm-mp":21,"md":[[129,20,1638833467986]],"md-mp":0,"mu":[[129,20,1638833467988]],"mu-mp":0,"v":1,"topLevel":{"st":1638833457105,"sc":{"availWidth":1920,"availHeight":1080,"width":1920,"height":1080,"colorDepth":24,"pixelDepth":24,"top":360,"left":2560,"availTop":360,"availLeft":2560,"mozOrientation":"landscape-primary","onmozorientationchange":null},"nv":{"permissions":{},"doNotTrack":"1","maxTouchPoints":0,"mediaCapabilities":{},"oscpu":"Linux x86_64","vendor":"","vendorSub":"","productSub":"20100101","cookieEnabled":true,"buildID":"20181001000000","mediaDevices":{},"credentials":{},"clipboard":{},"mediaSession":{},"webdriver":false,"hardwareConcurrency":16,"geolocation":{},"appCodeName":"Mozilla","appName":"Netscape","appVersion":"5.0 (X11)","platform":"Linux x86_64","userAgent":"Mozilla/5.0 (X11; Linux x86_64; rv:94.0) Gecko/20100101 Firefox/94.0","product":"Gecko","language":"en-US","languages":["en-US","en"],"onLine":true,"storage":{},"plugins":[]},"dr":"https://discord.com/","inv":false,"exec":false,"wn":[[1353,967,1,1638833457107]],"wn-mp":0,"xy":[[0,0,1,1638833457107]],"xy-mp":0,"mm":[[992,230,1638833467043],[989,244,1638833467063],[980,267,1638833467084],[966,299,1638833467105],[956,313,1638833467126],[956,315,1638833467147],[956,316,1638833467174],[956,328,1638833467195],[964,345,1638833467216],[1153,409,1638833467410],[961,496,1638833467542],[914,508,1638833467564],[871,519,1638833467585],[815,534,1638833467605]],"mm-mp":17.03030303030303},"session":[],"widgetList":["03e37er47cjx"],"widgetId":"03e37er47cjx","href":"https://discord.com/login","prev":{"escaped":false,"passed":false,"expiredChallenge":false,"expiredResponse":false}}`
}

func GetKey(client *http.Client, sitekey string, config *SiteConfig, n string) (string, error) {
	c, err := json.Marshal(config.C)

	if err != nil {
		return "", err
	}

	b := url.Values{}

	for k, v := range map[string]string{
		"v":          "e7ef6a8",
		"sitekey":    sitekey,
		"host":       "discord.com",
		"hl":         "en",
		"motionData": m(),
		"n":          n,
		"c":          string(c),
	} {
		b.Set(k, v)
	}

	req, err := http.NewRequest("POST", Base+"/getcaptcha?s="+sitekey, strings.NewReader(b.Encode()))

	if err != nil {
		return "", err
	}

	for k, v := range map[string]string{
		"Accept":             "application/json",
		"Accept-Language":    "en-GB,en-US;q=0.9",
		"Content-Type":       "application/x-www-form-urlencoded",
		"Origin":             "https://newassets.hcaptcha.com",
		"Referer":            "https://newassets.hcaptcha.com/",
		"sec-ch-ua":          Sec,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         UserAgent,
	} {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if !strings.Contains(string(body), "generated_pass_UUID") {
		return string(body), errors.New("failed to get pass uuid")
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", err
	}

	return data["generated_pass_UUID"].(string), nil
}

func TrySolve(proxy string) (string, error) {
	client := MakeClient(proxy)

	config, err := CheckSiteConfig(client, "discord.com", "f5561ba9-8f1e-40ca-9b5b-a0b3f719ef34")

	if err != nil {
		return "", err
	}

	n, err := SolveN(config.C.Req)

	if err != nil {
		return "", err
	}

	captcha, err := GetKey(client, "f5561ba9-8f1e-40ca-9b5b-a0b3f719ef34", config, n)

	if err != nil {
		return "", err
	}

	return captcha, nil
}

func SolveCaptcha(proxy string) (string, error) {
	for {
		key, err := TrySolve(proxy)

		if err != nil {
			return "", err
		}

		return key, nil
	}
}

func SolveCaptchaForce(proxy string) string {
	for {
		key, err := TrySolve(proxy)

		if err != nil {
			continue
		}

		return key
	}
}
