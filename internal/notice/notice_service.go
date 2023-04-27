package notice

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

const botUrl = "http://localhost:8085"

type Service struct {
}

func NewNoticeService() *Service {
	return &Service{}
}

type notifyRequest struct {
	Text string `json:"text"`
}

type addIPToWLRequest struct {
	IP4      string `json:"ip4"`
	UserName string `json:"user_name"`
	Comment  string `json:"comment"`
}

func (s *Service) Notify(text string) error {
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

func (s *Service) AddIPToWL(userName, ip4, comment string) {
	msg := addIPToWLRequest{
		IP4:      ip4,
		UserName: userName,
		Comment:  comment,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return
	}

	body := bytes.NewReader(b)

	w, err := http.Post(botUrl+"/wl", "application/json", body)
	if err != nil {
		return
	}

	if w.StatusCode != http.StatusOK {
		log.Println("failed to add ip to wl, status code: ", w.StatusCode)
	}
}
