package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v2"
)

const (
	Sender  = "Glitchd <no-reply@glitchd.io>"
	CharSet = "UTF-8"
)

func SendMail(recipient string, subject string, body string, textBody string) (bool, error) {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("RESEND_API_KEY")

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Glitchd <no-reply@glitchd.io>",
		To:      []string{recipient},
		Subject: subject,
		Html:    body,
	}

	sent, err := client.Emails.Send(params)

	if err != nil {
		fmt.Println("something went wrong when sending an email: ", err)
		return false, err
	}

	fmt.Println("Successfully sent email with id ", sent.Id)

	return true, nil
}
