package dmpmail

import (
	. "core"
	"os/exec"
	"net/smtp"
	"strings"
)
const (
	SENDER string      = "asbuilder@eone.com"
	PASS string        = "123456"
	MAIL_SERVER string = "192.168.0.10:25"
)
func SendMail(to, subject, body, mailtype string) error {

	mailTo := to + ";liwei@eone.com"
    LogInfo("Mail:", mailTo, subject, body, mailtype)
	hp := strings.Split(MAIL_SERVER, ":")
	auth := smtp.PlainAuth("", SENDER, PASS, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + mailTo + "\r\nFrom: " + SENDER + "<" + SENDER + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(mailTo, ";")
	err := smtp.SendMail(MAIL_SERVER, auth, SENDER, send_to, msg)
	return err
}


func SendRTX(msg string, storyID string,  receiver string) {
	rtxTo := receiver + "-liwei"
	LogInfo("RTX:", msg, storyID, rtxTo)
	cmd := exec.Command("rtxproxy.cmd", msg, storyID, rtxTo, "DMP提醒")
	bytes2, err2 := cmd.Output()
	if err2 != nil {
		LogError(err2, string(bytes2))
	}
}
