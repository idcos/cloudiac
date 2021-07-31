// Copyright 2021 CloudJ Company Limited. All rights reserved.

package mail

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/utils/logs"
	"gopkg.in/gomail.v2"
	"mime"
	"net"
	"strconv"
)

func SendMail(tos []string, subject, content string) e.Error {
	logs.Get().Infof("send mail:\n%s\n%s\n%s", tos, subject, content)

	srv := configs.Get().SMTPServer
	srvHost, srvPortStr, _ := net.SplitHostPort(srv.Addr)
	srvPort, _ := strconv.Atoi(srvPortStr)

	from := srv.From
	if from == "" {
		from = srv.UserName
	}

	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", from, srv.FromName)
	msg.SetHeader("To", tos...)
	msg.SetHeader("Subject", mime.BEncoding.Encode("utf-8", subject))
	msg.SetBody("text/html", content)

	conn := gomail.NewDialer(srvHost, srvPort, srv.UserName, srv.Password)
	if err := conn.DialAndSend(msg); err != nil {
		logs.Get().Errorf("send mail error: %v", err)
		return e.New(e.MailServerError, err)
	}
	return nil
}
