package goweb_mail

import (
	"crypto/tls"
	"fmt"
	gomail "gopkg.in/gomail.v2"
	"os"
	"strings"
	"time"
)

type MailClient struct {
	Account    string
	Password   string
	StartTime  string //开通时间，申请邮箱6个月密码过期
	MailServer string
	Port       int
	Debug      bool
	Admins     []string
	dialer     *gomail.Dialer
}

func NewMailClient(account, password, mailServer, startTime string, port int, debug bool, admins []string) (*MailClient, error) {
	_, err := time.ParseInLocation("2006-01-02", startTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("mail account start_time(%s) must like 2006-01-02", startTime)
	}

	dialer := gomail.NewDialer(mailServer, port, account, password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	cl, err := dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("can't dial(%s:%d) with account(%s) and password(%s)",
			mailServer, port, account, password)
	}
	cl.Close()

	client := &MailClient{
		Account:    account,
		Password:   password,
		StartTime:  startTime,
		MailServer: mailServer,
		Port:       port,
		Debug:      debug,
		Admins:     admins,
		dialer:     dialer,
	}

	client.Admins = client.formatUserMail(admins)
	if len(client.Admins) < 1 {
		return nil, fmt.Errorf("no admins")
	}

	go client.checkPasswdExpired()
	return client, nil
}

func (mc *MailClient) checkPasswdExpired() {
	startTime, _ := time.ParseInLocation("2006-01-02", mc.StartTime, time.Local)
	warnTime := startTime.AddDate(0, 0, 165)
	tick := time.NewTicker(time.Duration(24) * time.Hour)
	for {
		<-tick.C
		now := time.Now()
		if warnTime.Before(now) {
			content := fmt.Sprintf("%s密码小于15天将失效！！！", mc.Account)
			mc.SendEmailToAdmin(content, "邮箱密码即将过期")
		}
	}
}

func (mc *MailClient) SendEmailToAdmin(content string, subject ...string) error {
	sub := "程序异常，请尽快处理！！！"
	if len(subject) > 0 {
		sub = fmt.Sprintf("%s", subject[0])
	}
	hostname, _ := os.Hostname()
	content = fmt.Sprintf("Send To Admin\nFrom:%s \n\n%s", hostname, content)
	return mc.send(sub, content, mc.Admins)
}

func (mc *MailClient) SendEmailToUser(users []string, subject string, content string, contentType ...string) error {
	if mc.Debug {
		adminContent := fmt.Sprintf("To:%s\n\nSubject:%s\n\nContent:%s",
			strings.Join(users, ","), subject, content)
		adminSubject := "[Debug模式]SendEmailToUser"
		return mc.SendEmailToAdmin(adminContent, adminSubject)
	}

	user_mails := mc.formatUserMail(users)

	mail_type := "text/plain"
	if len(contentType) > 0 {
		mail_type = contentType[0]
	}
	return mc.send(subject, content, user_mails, mail_type)
}

func (mc *MailClient) formatUserMail(users []string) []string {
	userMails := make([]string, 0)
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}

		if strings.Contains(user, "@") {
			userMails = append(userMails, user)
		} else {
			user = fmt.Sprintf("%s", user)
			userMails = append(userMails, user)
		}
	}
	return userMails
}

func (mc *MailClient) send(subject, content string, users []string, contentType ...string) error {
	if len(users) < 1 {
		return fmt.Errorf("no mail recipients")
	}
	if len(content) < 1 {
		return fmt.Errorf("no mail content")
	}
	if len(subject) < 1 {
		return fmt.Errorf("no mail subject")
	}
	if mc.dialer == nil {
		return fmt.Errorf("no dialer, can't send mail")
	}

	mailType := "text/plain"
	if len(contentType) > 0 {
		mailType = contentType[0]
	}
	m := gomail.NewMessage()
	// @后可以是不同邮箱公司地址，例： qq.com; 163.com;gmail.com
	m.SetHeader("From", mc.Account+"@"+"gmail.com")
	m.SetHeader("To", users...)
	m.SetHeader("Subject", subject)
	m.SetBody(mailType, content)
	err := mc.dialer.DialAndSend(m)
	return err
}
