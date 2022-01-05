package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	Base = "http://api.kopeechka.store"
	DevId = "19"
)

type Client struct {
	APIKey string
	Client *http.Client
}

type MailRequest struct {
	Site string
	Type string
	Sender string
	Regex string
	Investor string
	NoSearch string
	Subject string
	Clear string
}

type Mail struct {
	Id string
	Address string
	Activated bool
}

func (c Client) GetBalance () (string, error) {
	resp, err := c.Client.Get(Base + "/user-balance?token=" + c.APIKey + "&type=JSON&api=2.0")

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if body == nil {
		return c.GetBalance()
	}

	if string(body)[0] == '<' {
		time.Sleep(time.Millisecond * 500)

		return c.GetBalance()
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", err
	}

	switch data["status"].(string) {
	case "OK":
		return fmt.Sprintf("%v", data["balance"]), nil
	case "ERROR":
		return "", errors.New(fmt.Sprintf("%v", data["value"]))
	default:
		return "", errors.New(string(body))
	}
}

func (c Client) RequestMail (r MailRequest) (*Mail, error) {
	uri := Base + "/mailbox-get-email?site=" + r.Site + "&mail_type=" + r.Type + "&sender=" + r.Sender +
		"&regex=" + r.Regex + "&token=" + c.APIKey + "&soft=" + DevId + "&investor=" + r.Investor +
		"&no_search=" + r.NoSearch + "&type=JSON&subject=" + r.Subject + "&clear=" + r.Clear + "&api=2.0"

	resp, err := c.Client.Get(uri)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if body == nil {
		return c.RequestMail(r)
	}

	if string(body)[0] == '<' {
		time.Sleep(time.Millisecond * 500)

		return c.RequestMail(r)
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}

	switch data["status"].(string) {
	case "OK":
		return &Mail{
			Id: data["id"].(string),
			Address: data["mail"].(string),
			Activated: r.NoSearch != "1",
		}, nil
	case "ERROR":
		return nil, errors.New(data["value"].(string))
	default:
		return nil, errors.New(string(body))
	}
}

func (c Client) CancelMail (m *Mail) error {
	resp, err := c.Client.Get(Base + "/mailbox-cancel?id=" + m.Id + "&token=" + c.APIKey +  "&type=JSON&api=2.0")

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if len(string(body)) > 1 {
		if string(body)[0] == '<' {
			time.Sleep(time.Millisecond * 500)

			return c.CancelMail(m)
		}
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return err
	}

	switch data["status"].(string) {
	case "OK":
		return nil
	case "ERROR":
		return errors.New(data["value"].(string))
	default:
		return errors.New(string(body))
	}
}

func (c Client) ActivateSearch (m *Mail) error {
	if m.Activated {
		return nil
	}

	resp, err := c.Client.Get(Base + "/mailbox-activate-post?id=" + m.Id + "&token=" + c.APIKey + "&type=JSON&api=2.0")

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if len(string(body)) > 1 {
		if string(body)[0] == '<' {
			time.Sleep(time.Millisecond * 500)

			return c.ActivateSearch(m)
		}
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return err
	}

	switch data["status"].(string) {
	case "OK":
		return nil
	case "ERROR":
		return errors.New(data["value"].(string))
	default:
		return errors.New(string(body))
	}
}

func (c Client) GetMessage (m *Mail) (string, error) {
	if !m.Activated {
		err := c.ActivateSearch(m)

		if err != nil {
			return "", err
		}
	}

	resp, err := c.Client.Get(Base + "/mailbox-get-message?full=0&id=" + m.Id + "&token=" + c.APIKey + "&type=JSON&api=2.0")

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if len(string(body)) > 1 {
		if string(body)[0] == '<' {
			time.Sleep(time.Millisecond * 500)

			return c.GetMessage(m)
		}
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", err
	}

	switch data["status"].(string) {
	case "OK":
		return data["value"].(string), nil
	case "ERROR":
		value := data["value"].(string)

		if value == "WAIT_LINK" {
			return "", nil
		}

		return "", errors.New(value)
	default:
		return "", errors.New(string(body))
	}
}

func (c Client) WaitForMessage (m *Mail, delay time.Duration, max int) (string, error) {
	for i := 0; i < max; i++ {
		message, err := c.GetMessage(m)

		if err != nil {
			return "", err
		}

		if message != "" {
			return message, nil
		}

		time.Sleep(delay)
	}

	return "", errors.New("failed to receive message in time")
}
