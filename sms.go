package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	PhoneBase = "https://smshub.org"
	PhoneKey = ""
	PhoneClient = &http.Client{}
)

func Buy (country string, operator string, service string) (*PhonePurchase, error) {
	resp, err := PhoneClient.Get(PhoneBase + "/stubs/handler_api.php?api_key=" + PhoneKey + "&action=getNumber&service=" +
		service + "&operator=" + operator + "&country=" + country)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(body), "ACCESS_NUMBER") {
		if string(body) == "" || strings.Contains(string(body), "429") {
			time.Sleep(time.Second * 2)

			return Buy(country, operator, service)
		}

		return nil, errors.New(string(body))
	}

	t := strings.Split(string(body), ":")

	return &PhonePurchase{
		Id: t[1],
		Phone: "+" + t[2],
	}, nil
}

func Check (o *PhonePurchase) (*PhoneOrder, error) {
	resp, err := PhoneClient.Get(PhoneBase + "/stubs/handler_api.php?api_key=" + PhoneKey + "&action=getStatus&id=" + o.Id)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(body), "STATUS") {
		if string(body) == "" || strings.Contains(string(body), "429") {
			time.Sleep(time.Second * 2)

			return Check(o)
		}

		return nil, errors.New(string(body))
	}

	status := func() string {
		if strings.Contains(string(body), "WAIT") {
			return "WAIT"
		} else if strings.Contains(string(body), "CANCEL"){
			return "CANCELLED"
		} else {
			return "READY"
		}
	}()

	return &PhoneOrder{
		Id: o.Id,
		Phone: o.Phone,
		Status: status,
		Code: func() string {
			if status == "READY" {
				if PhoneBase == "https://activation.pw" {
					temp := strings.Split(string(body), " ")

					return temp[len(temp) - 1]
				} else {
					return strings.Split(string(body), ":")[1]
				}
			} else {
				return ""
			}
		}(),
	}, nil
}

func ChangeStatus (o *PhoneOrder, status string) error {
	resp, err := PhoneClient.Post(PhoneBase + "/stubs/handler_api.php?api_key=" + PhoneKey + "&action=setStatus&status=" + status + "&id=" + o.Id, "application/json", nil)

	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}
