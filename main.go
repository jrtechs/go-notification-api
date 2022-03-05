package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jrtechs/go-notification-api/conf"
	"github.com/jrtechs/go-notification-api/email"

	"github.com/joho/godotenv"
)

var ValidToken string

type EmailRequest struct {
	Token   string
	Message string
	Subject string
	Email   string
}

func sendEmail(res http.ResponseWriter, req *http.Request) {

	var postData EmailRequest
	err := json.NewDecoder(req.Body).Decode(&postData)
	if err != nil {
		panic(err)
	}

	if postData.Token == ValidToken {
		conf.Logger.Println("Token accepted on request")

		if postData.Email != "" && postData.Message != "" && postData.Subject != "" {
			conf.Logger.Printf("Sending email with subject: '%s' and message: '%s'", postData.Subject, postData.Message)
			email.SendEmail(postData.Email, postData.Subject, postData.Message)
		}
	} else {
		conf.Logger.Println("Invalid token")
	}
}

func main() {

	if err := conf.InitLogger(); err != nil {
		fmt.Errorf("Error initializing loggers! %s", err)
	}

	errEnv := godotenv.Load(".env")
	if errEnv != nil {
		conf.Logger.Fatal("Error loading .env file")
	}

	if errEmail := email.InitConfig(); errEmail != nil {
		conf.Logger.Fatal("Failed to initialize email configuration")
	}

	//email.SendEmail("Test subject custom", "Test message from golang")

	token, envExists := os.LookupEnv("token")
	if !envExists {
		conf.Logger.Panic("Unable to fetch 'token' environment variable")
	}
	ValidToken = token
	conf.Logger.Println("Loaded auth token")

	port, envExists := os.LookupEnv("port")
	if !envExists {
		conf.Logger.Panic("Unable to fetch 'port' environment variable")
	}

	http.HandleFunc("/sendEmail", sendEmail)

	conf.Logger.Print("Starting server on port: ", port)
	conf.Logger.Fatal(http.ListenAndServe(":"+port, nil))
}
