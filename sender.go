package mailSender

import (
	"bufio"
	"bytes"
	"errors"
	"html/template"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

type dealer interface {
	Dial() (gomail.SendCloser, error)
}

type AsyncSender interface {
	SendAsync(message Message)
}

type EmailData map[interface{}]interface{}

type Message struct {
	Message  *gomail.Message
	Template *template.Template
	Data     EmailData
}

type Sender struct {
	AsyncSender
	Channel chan Message
	Closer  gomail.SendCloser
}

var tryConnCnt uint = 0

func ParseEss(ess string) (host string, port int, username string, password string) {
	str := strings.Split(ess, ";")
	if len(str) < 3 {
		err := errors.New("minimum match not found")
		log.Println(err)
		return "", 0, "", ""
	}
	host, portStr, err := net.SplitHostPort(str[0])
	if err != nil {
		log.Println(err)
		return "", 0, "", ""
	}
	port, _ = strconv.Atoi(portStr)
	username, password = str[1], str[2]
	return
}

// Create method expects Email Sender Server parameter in format %s:%s;%s;%s
func Create(s *Sender) AsyncSender {
	go s.asyncSender()
	return s
}

func GetCloser(d dealer) gomail.SendCloser {
	closer, err := d.Dial()
	if err == nil {
		return closer
	}
	tryConnCnt++
	if tryConnCnt < 5 {
		GetCloser(d)
	}
	return nil
}

func (s *Sender) asyncSender() {
	for msg := range s.Channel {
		message := prepareMsg(msg)
		if err := gomail.Send(s.Closer, message); err != nil {
			_, err := message.WriteTo(bufio.NewWriter(os.Stdout))
			log.Println(err)
		}
	}
}

func prepareMsg(msg Message) *gomail.Message {
	buffer := new(bytes.Buffer)
	if err := msg.Template.Execute(buffer, msg.Data); err != nil {
		log.Println(err)
	}
	msg.Message.SetBody("text/html", buffer.String())
	return msg.Message
}

func (s *Sender) SendAsync(message Message) {
	s.Channel <- message
}
