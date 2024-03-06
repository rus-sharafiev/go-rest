package common

import (
	"slices"
)

type config struct {
	Db                *string   `json:"db"`
	Port              *string   `json:"port"`
	StaticDir         *string   `json:"staticDir"`
	UploadPath        *string   `json:"uploadPath"`
	UploadDir         *string   `json:"uploadDir"`
	FormdataWhitelist *[]string `json:"formdataWhitelist"`
	FormdataBlacklist *[]string `json:"formdataBlacklist"`
	JwtKey            *string   `json:"jwtKey"`
	RecaptchaSecret   *string   `json:"recaptchaSecret"`
	MailLogin         *string   `json:"mailLogin"`
	MailPassword      *string   `json:"mailPassword"`
	MailHost          *string   `json:"mailHost"`
}

var Config config

func (c config) IsNotValid() bool {
	mandatotyFields := []*string{
		c.Port,
		c.StaticDir,
		c.UploadPath,
		c.UploadDir,
		c.JwtKey,
	}
	return slices.Contains(mandatotyFields, nil)
}
