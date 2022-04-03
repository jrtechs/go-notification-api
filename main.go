package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/jrtechs/go-notification-api/conf"
	"github.com/jrtechs/go-notification-api/email"

	"github.com/joho/godotenv"
)

var ValidToken string

var Debug bool = false

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
		conf.Logger.Print("Error decoding response", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if postData.Token == ValidToken {
		conf.Logger.Println("Token accepted on request")
		if postData.Email != "" && postData.Message != "" && postData.Subject != "" {
			if !Debug {
				emailErr := email.SendEmail(postData.Email, postData.Subject, postData.Message)
				if emailErr != nil {
					res.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				conf.Logger.Printf("\nSending email to %s\t\nwith subject: '%s'\t\n and message: '%s'", postData.Email, postData.Subject, postData.Message)
			}
		}
	} else {
		conf.Logger.Println("Invalid token")
		res.WriteHeader(http.StatusUnauthorized)
	}
}

const recaptchaServerName = "https://www.google.com/recaptcha/api/siteverify"

// loaded in main from captcha env variable
var recaptchaPrivateKey string

//maps sites URL to email to send form input to
var sites map[string]string

type RecaptchaResponse struct {
	Success bool `json:"success"`
}

func verifyCaptcha(captcha string) bool {
	resp, err := http.PostForm(recaptchaServerName,
		url.Values{"secret": {recaptchaPrivateKey}, "response": {captcha}})
	if err != nil {
		conf.Logger.Printf("Post error: %s\n", err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		conf.Logger.Printf("Read error: could not read body: %s", err)
		return false
	}
	r := RecaptchaResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		conf.Logger.Printf("Read error: got invalid JSON: %s", err)
		return false
	}
	return r.Success
}

func contact(res http.ResponseWriter, req *http.Request) {
	captcha := req.PostFormValue("g-recaptcha-response")
	targetWebsite := req.PostFormValue("site")
	destinationEmail, ok := sites[targetWebsite]

	if !ok {
		res.Write([]byte("Invalid form site parameter."))
		res.Header().Set("Content-Type", "application/text")
		res.WriteHeader(http.StatusInternalServerError)
	}

	visiterEmail := req.PostFormValue("email")
	visiterName := req.PostFormValue("name")
	visiterMessage := req.PostFormValue("message")

	subject := targetWebsite + " form submission -- from " + visiterEmail
	message := fmt.Sprintf("Message from contact form on %s \n email: %s \n name: %s \n message:\n %s", targetWebsite, visiterEmail, visiterName, visiterMessage)

	if visiterEmail != "" && visiterName != "" && visiterMessage != "" && captcha != "" {
		if verifyCaptcha(captcha) {
			conf.Logger.Println("Valid captcha recieved.")
			emailErr := email.SendEmail(destinationEmail, subject, message)
			if emailErr != nil {
				res.WriteHeader(http.StatusInternalServerError)
				res.Header().Set("Content-Type", "application/text")
				res.Write([]byte("Unable to send email."))
			}
			http.Redirect(res, req, "https://"+targetWebsite, 303)
		} else {
			conf.Logger.Println("Invalid captcha recieved.")
			res.Write([]byte("Invalid captcha recieved."))
			res.Header().Set("Content-Type", "application/text")
			res.WriteHeader(http.StatusUnauthorized)
		}
	} else {
		res.Write([]byte("Invalid form parameters."))
		res.Header().Set("Content-Type", "application/text")
		res.WriteHeader(http.StatusInternalServerError)
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

	debugEnv, envExists := os.LookupEnv("debug")
	if envExists {
		if debugEnv == "true" {
			Debug = true
		}
	}

	recaptcha, envExists := os.LookupEnv("captcha")
	if envExists {
		recaptchaPrivateKey = recaptcha
	} else {
		conf.Logger.Panic("Unable to load 'captcha' environment variable that stores captcha secret")
	}

	file, err := ioutil.ReadFile("sites.json")
	if err != nil {
		conf.Logger.Panic("Unable to read sites.json file")
	}
	json.Unmarshal([]byte(file), &sites)

	conf.Logger.Println(sites)

	http.HandleFunc("/sendEmail", sendEmail)
	http.HandleFunc("/contact", contact)

	conf.Logger.Print("Starting server on port: ", port)
	conf.Logger.Fatal(http.ListenAndServe(":"+port, nil))
}
