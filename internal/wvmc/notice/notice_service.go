package notice

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

const botUrl = "http://localhost:8085"

type NoticeService struct {
}

func NewNoticeService() *NoticeService {
	return &NoticeService{}
}

type notice struct {
	Text string `json:"text"`
}

func (s *NoticeService) Notify(text string) error {
	msg := notice{text}
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
