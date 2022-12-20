# goweb-mail
建立一个邮局，用于邮件发送

useage:

```golang
func main() {
	// PostOffice Config
	server := ""
	port := ""
	username := ""
	password := ""

	// Message info
	from := ""
	subject := ""
	body := ""
	to := ""
	PostOffice := NewPostOffice(server, port, username, password)
	message := Message{
		From:    from,
		Subject: subject,
		Body:    body,
	}
	usPostman := HirePostman(message, PostOffice)
	usPostman.SendMail(to)
}
```
