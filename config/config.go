package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"net/url"
)

var Config struct {
	Mode     string `env:"MODE" envDefault:"dev"`
	DbUrl    string `env:"DB_URL,required"`
	KongUrl  string `env:"KONG_URL,required"`
	RedisUrl string `env:"REDIS_URL"`
	// sending email config
	EmailServerNoReplyUrl url.URL `env:"EMAIL_SERVER_NO_REPLY_URL,required"`
	EmailDomain           string  `env:"EMAIL_DOMAIN,required"`
	SiteName              string  `env:"SITE_NAME" envDefault:"OpenTreeHole"`

	VerificationCodeExpires int `env:"VERIFICATION_CODE_EXPIRES" envDefault:"10"`

	HoleFloorSize int `env:"HOLE_FLOOR_SIZE" envDefault:"10"`

	// file secrets
	EmailList []string `env:"EMAIL_LIST,file" envDefault:"/var/run/secrets/email_list"` // an email list in json array format
}

func InitConfig() {
	var err error
	if err = env.Parse(&Config); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", &Config)
}
