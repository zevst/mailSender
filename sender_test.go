package mailSender

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mylockerteam/mailSender/mocks"
	"gopkg.in/gomail.v2"
)

const success = "success"
const failure = "failure"

const essExample = "smtp.gmail.com:465;test@example.com;qwerty123"
const emailTemplateSuccess = "<div>Hi, {{ .Name }}!</div>"

func mockSelectorGetCloser(name string, mockCtrl *gomock.Controller) *mocks.Mockdealer {
	mockdealer := mocks.NewMockdealer(mockCtrl)
	switch name {
	case success:
		mockdealer.EXPECT().Dial().Return(&mocks.MockSendCloser{}, nil).AnyTimes()
		return mockdealer
	case failure:
		mockdealer.EXPECT().Dial().Return(&mocks.MockSendCloser{}, errors.New("it's bad news")).AnyTimes()
		return mockdealer
	default:
		fmt.Println(fmt.Errorf("name %s not found", name))
	}
	return &mocks.Mockdealer{}
}

func Test_getCloser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	type args struct {
		d dealer
	}
	tests := []struct {
		name string
		args args
		want gomail.SendCloser
	}{
		{
			name: success,
			args: args{
				d: mockSelectorGetCloser(success, mockCtrl),
			},
			want: &mocks.MockSendCloser{},
		},
		{
			name: failure,
			args: args{
				d: mockSelectorGetCloser(failure, mockCtrl),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCloser(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCloser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEss(t *testing.T) {
	type args struct {
		ess string
	}
	tests := []struct {
		name         string
		args         args
		wantHost     string
		wantPort     int
		wantUsername string
		wantPassword string
	}{
		{
			name: success,
			args: args{
				ess: essExample,
			},
			wantHost:     "smtp.gmail.com",
			wantPort:     465,
			wantUsername: "test@example.com",
			wantPassword: "qwerty123",
		},
		{
			name: failure,
			args: args{
				ess: "smtp.gmail.com;test@example.com;qwerty123",
			},
			wantHost:     "",
			wantPort:     0,
			wantUsername: "",
			wantPassword: "",
		},
		{
			name: failure,
			args: args{
				ess: "",
			},
			wantHost:     "",
			wantPort:     0,
			wantUsername: "",
			wantPassword: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, gotUsername, gotPassword := ParseEss(tt.args.ess)
			if gotHost != tt.wantHost {
				t.Errorf("ParseEss() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("ParseEss() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
			if gotUsername != tt.wantUsername {
				t.Errorf("ParseEss() gotUsername = %v, want %v", gotUsername, tt.wantUsername)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("ParseEss() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
		})
	}
}

func Test_prepareMsg(t *testing.T) {
	type args struct {
		msg Message
	}
	m := gomail.NewMessage()
	tpl, _ := template.New(success).Parse(emailTemplateSuccess)
	tests := []struct {
		name string
		args args
		want *gomail.Message
	}{
		{
			name: success,
			args: args{
				msg: Message{
					Message: m,
					Data: EmailData{
						"Name": "Tester",
					},
					Template: tpl,
				},
			},
			want: m,
		},
		{
			name: failure,
			args: args{

				msg: Message{
					Message: m,
					Data: EmailData{
						"name": "Tester",
					},
					Template: template.New(failure),
				},
			},
			want: m,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareMsg(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sender_SendAsync(t *testing.T) {
	type fields struct {
		channel chan Message
		closer  gomail.SendCloser
	}
	type args struct {
		message Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: success,
			fields: fields{
				channel: make(chan Message, 1),
			},
			args: args{
				message: Message{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sender{
				Channel: tt.fields.channel,
				Closer:  tt.fields.closer,
			}
			s.SendAsync(tt.args.message)
		})
	}
}

func Test_sender_asyncSender(t *testing.T) {
	type fields struct {
		channel chan Message
		closer  gomail.SendCloser
	}
	m := gomail.NewMessage()
	tpl, _ := template.New(success).Parse(emailTemplateSuccess)
	messages := make(chan Message, 1)
	messages <- Message{
		Message: m,
		Data: EmailData{
			"Name": "Tester",
		},
		Template: tpl,
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			fields: fields{
				channel: messages,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sender{
				Channel: tt.fields.channel,
				Closer:  tt.fields.closer,
			}
			go s.asyncSender()
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		s *Sender
	}
	dialer := gomail.NewDialer(ParseEss(essExample))
	object := &Sender{
		Channel: make(chan Message, 1),
		Closer:  GetCloser(dialer),
	}
	tests := []struct {
		name string
		args args
		want AsyncSender
	}{
		{
			args: args{
				s: object,
			},
			want: object,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Create(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
