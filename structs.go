package main

import "net/http"

type DiscordClient struct {
	Mail *Mail
	Client *http.Client
	Cookie string
	Fingerprint string
	Token string
	Account *Account
	Attempted bool
	Attempted2 bool
}

type Account struct {
	Token string
	Username string
	Email string
	Password string
	EmailVerified bool
	PhoneVerified bool
	Avatar string
	PhoneDone bool
	EmailDone bool
}

type RegisterPayload struct {
	Consent bool `json:"consent"`
	Fingerprint string `json:"fingerprint"`
	Captcha string `json:"captcha_key"`
	Username string `json:"username"`
	Invite string `json:"invite"`
}

type PhonePurchase struct {
	Id string
	Phone string
}

type PhoneOrder struct {
	Id string
	Phone string
	Status string
	Code string
}

type Config struct {
	EmailAPI string `json:"email_api"`
	EmailDomain string `json:"email_domain"`
	PhoneAPI string `json:"phone_api"`
	PhoneSite string `json:"phone_site"`
	PhoneCountry string `json:"phone_country"`
	PhoneOperator string `json:"phone_operator"`
	PhoneService string `json:"phone_service"`
	Max int `json:"max"`
	Threads int `json:"threads"`
	EmailVerify bool `json:"email_verify"`
	PhoneVerify bool `json:"phone_verify"`
	Invite string `json:"invite"`
}
