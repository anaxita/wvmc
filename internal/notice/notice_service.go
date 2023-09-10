package notice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const botUrl = "http://localhost:8085" // TODO: move to config.

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

func (s *Service) AddIPToWL(userName, ip4, comment string) error {
	msg := addIPToWLRequest{
		IP4:      ip4,
		UserName: userName,
		Comment:  comment,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	w, err := http.Post(botUrl+"/wl", "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("post request: %w", err)
	}

	defer w.Body.Close()

	if w.StatusCode != http.StatusOK {
		return fmt.Errorf("error send message: %s", w.Status)
	}

	return nil
}
