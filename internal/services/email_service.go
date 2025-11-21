package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailService struct {
	apiToken string
	apiURL   string
}

func NewEmailService(apiToken string) *EmailService {
	return &EmailService{
		apiToken: apiToken,
		apiURL:   "https://send.api.mailtrap.io/api/send",
	}
}

type mailtrapSender struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type mailtrapRecipient struct {
	Email string `json:"email"`
}

type mailtrapPayload struct {
	From     mailtrapSender      `json:"from"`
	To       []mailtrapRecipient `json:"to"`
	Subject  string              `json:"subject"`
	Text     string              `json:"text"`
	Category string              `json:"category"`
}

func (s *EmailService) SendEmail(toEmail, subject, body string) error {
	if s.apiToken == "" {
		fmt.Println("EmailService: API token is empty")
		return fmt.Errorf("API token is empty")
	}

	payload := mailtrapPayload{
		From: mailtrapSender{
			Email: "hello@demomailtrap.co",
			Name:  "Obsonarium Support",
		},
		To: []mailtrapRecipient{
			{Email: toEmail},
		},
		Subject:  subject,
		Text:     body,
		Category: "Query Resolution",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	fmt.Printf("EmailService: Sending email to %s with subject '%s'\n", toEmail, subject)

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+s.apiToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer res.Body.Close()

	fmt.Printf("EmailService: Response status: %s\n", res.Status)

	if res.StatusCode >= 400 {
		return fmt.Errorf("email service returned status: %s", res.Status)
	}

	return nil
}
