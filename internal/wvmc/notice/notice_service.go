package notice

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/anaxita/logit"
	"net/http"
)

const botUrl = "http://localhost:8085"

type KMSBOT struct {
}

func NewNoticeService() *KMSBOT {
	return &KMSBOT{}
}

type notifyRequest struct {
	Text string `json:"text"`
}

type addIPToWLRequest struct {
	IP4      string `json:"ip4"`
	UserName string `json:"user_name"`
	Comment  string `json:"comment"`
}

func (s *KMSBOT) Notify(text string) error {
	msg := notifyRequest{text}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	body := bytes.NewReader(b)

	w, err := http.Post(botUrl+"/send", "application/json", body)
	if err != nil {
		return err
	}

	if w.StatusCode != http.StatusOK {
		return errors.New("error send message")
	}

	return nil
}

func (s *KMSBOT) AddIPToWL(userName, ip4, comment string) {
	msg := addIPToWLRequest{
		IP4:      ip4,
		UserName: userName,
		Comment:  comment,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		logit.Log("failed to marshal to json request for adding ip to wl: ", err)
		return
	}

	body := bytes.NewReader(b)

	w, err := http.Post(botUrl+"/wl", "application/json", body)
	if err != nil {
		logit.Log("failed to send request for adding ip to wl: ", err)
		return
	}

	if w.StatusCode != http.StatusOK {
		logit.Log("failed to add ip to wl, status code: ", w.StatusCode)
	}
}
