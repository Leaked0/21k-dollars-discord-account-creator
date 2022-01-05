package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
	"woen_gen/bypass"
)

var kc *Client
var c *Config

func SetTitle(title string) (int, error) {
	handle, err := syscall.LoadLibrary("Kernel32.dll")

	if err != nil {
		return 0, err
	}

	defer syscall.FreeLibrary(handle)

	proc, err := syscall.GetProcAddress(handle, "SetConsoleTitleW")

	if err != nil {
		return 0, err
	}

	r, _, err := syscall.Syscall(proc, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)

	return int(r), err
}

func Contains(x string, y ...string) bool {
	for _, z := range y {
		if strings.Contains(x, z) {
			return true
		}
	}

	return false
}

func Create(username string, avatar string, proxy string) (*Account, error) {
	p, _ := url.Parse(func() string {
		if !strings.Contains(proxy, "http://") {
			return "http://" + proxy
		}

		return proxy
	}())

	client := &DiscordClient{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MaxVersion: tls.VersionTLS12,
					CipherSuites: []uint16 {
						0x0a0a, 0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030,
						0xcca9, 0xcca8, 0xc013, 0xc014, 0x009c, 0x009d, 0x002f, 0x0035,
					},
					InsecureSkipVerify: true,
					CurvePreferences: []tls.CurveID{
						tls.CurveID(0x0a0a),
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
					},
				},
				Proxy: http.ProxyURL(p),
			},
		},
	}

	err := client.Register(username, proxy, false, c.Invite)

	if err != nil {
		return nil, err
	}

	err = client.ConnectToWebsocket(proxy)

	if err != nil {
		return nil, err
	}

	kc = &Client{
		APIKey: c.EmailAPI,
		Client: &http.Client{},
	}

	if err != nil {
		return nil, err
	}

	err = client.ClaimAccount(nil)

	if err != nil {
		return nil, err
	}

	var ee error
	var pe error

	if c.EmailVerify {
		go func() {
			ee = client.EmailVerify(client.Mail, proxy)
		}()
	} else {
		client.Account.EmailDone = true
	}

	if c.PhoneVerify {
		go func() {
			pe = client.PhoneVerify(func() (*PhonePurchase, error) {
				return Buy(c.PhoneCountry, c.PhoneOperator, c.PhoneService)
			}, Check, ChangeStatus)
		}()
	} else {
		client.Account.PhoneDone = true
	}

	for !client.Account.PhoneDone || !client.Account.EmailDone {
		time.Sleep(time.Millisecond * 150)
	}

	if ee != nil || pe != nil {
		return nil, func() error {
			if ee != nil {
				return ee
			} else {
				return pe
			}
		}()
	}

	go client.AddPFP(avatar)

	return client.Account, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	go bypass.Setup()

	r, err := ioutil.ReadFile("config.json")

	if err != nil {
		fmt.Println(err)

		return
	}

	err = json.Unmarshal(r, &c)

	if err != nil {
		fmt.Println(err)

		return
	}

	PhoneBase = c.PhoneSite
	PhoneKey = c.PhoneAPI
	kc = &Client{
		Client: &http.Client{},
		APIKey: c.EmailAPI,
	}

	fully, _ := os.OpenFile("fully.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	email, _ := os.OpenFile("email.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	unclaimed, _ := os.OpenFile("unclaimed.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	info, _ := os.OpenFile("info.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)

	var usernames []string
	var proxies []string

	f, _ := os.Open("usernames.txt")

	s := bufio.NewScanner(f)

	for s.Scan() {
		usernames  = append(usernames, s.Text())
	}

	f, _ = os.Open("proxies.txt")

	s = bufio.NewScanner(f)

	for s.Scan() {
		proxies = append(proxies, s.Text())
	}

	var avatars []string

	wd, _ := os.Getwd()

	err = filepath.Walk(wd + "\\avatars", func(path string, info fs.FileInfo, err error) error {
		img, err := os.Open(path)

		if err != nil {
			return err
		}

		f, _ := img.Stat()
		buf := make([]byte, f.Size())

		reader := bufio.NewReader(img)
		_, _ = reader.Read(buf)

		if string(buf) == "" {
			return nil
		}

		b64 := base64.StdEncoding.EncodeToString(buf)
		s := strings.Split(path, ".")

		avatars = append(avatars, "data:image/" + s[len(s) - 1] + ";base64," + b64)

		return nil
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("Loaded %v avatars & %v usernames.", len(avatars), len(usernames)))

	wg := sync.WaitGroup{}

	max := c.Max
	total := 0
	fails := 0
	start := time.Now()

	go func() {
		for {
			_, _ = SetTitle(fmt.Sprintf("Tokens Created: %v/%v; Fails: %v; Time %v seconds", total, max, fails, (time.Now().UnixNano() - start.UnixNano()) / 1e+9))
		}
	}()

	for i := 0; i < c.Threads; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				acc, err := Create(usernames[rand.Intn(len(usernames))], avatars[rand.Intn(len(avatars))], proxies[rand.Intn(len(proxies))])

				if err != nil {
					if !Contains(err.Error(), "EOF") {
						fails++

						fmt.Println(err)
					}

					continue
				}

				total++

				fmt.Println(acc.Token)

				if acc.PhoneVerified {
					_, _ = fully.Write([]byte(acc.Token + "\n"))
				} else if acc.EmailVerified {
					_, _ = email.Write([]byte(acc.Token + "\n"))
				} else {
					_, _ = unclaimed.Write([]byte(acc.Token + "\n"))
				}

				_, _ = info.Write([]byte(fmt.Sprintf("%v:%v:%v\n", acc.Email, acc.Password, acc.Token)))

				if total >= max {
					return
				}
			}
		}()
	}

	wg.Wait()
}
