package utils

import (
	"crypto/tls"
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
	"treehole_backend/config"
)

func SendCodeEmail(code, receiver string) error {
	emailUsername := config.Config.EmailServerNoReplyUrl.User.Username() + "@" + config.Config.EmailDomain
	emailPassword, _ := config.Config.EmailServerNoReplyUrl.User.Password()
	e := &email.Email{
		To:      []string{receiver},
		From:    emailUsername,
		Subject: fmt.Sprintf("%s 邮箱验证", config.Config.SiteName),
		Text:    []byte(fmt.Sprintf("您的验证码是 %s，10分钟内有效\n如果您意外收到此邮件，请忽略\n", code)),
	}

	return e.SendWithTLS(
		config.Config.EmailServerNoReplyUrl.Host,
		smtp.PlainAuth(
			"",
			emailUsername,
			emailPassword,
			config.Config.EmailServerNoReplyUrl.Hostname(),
		),
		&tls.Config{
			ServerName: config.Config.EmailServerNoReplyUrl.Hostname(),
		},
	)
}
