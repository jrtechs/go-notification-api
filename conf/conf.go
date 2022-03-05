package conf

import (
	"io"
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() error {
	file, error := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if error != nil {
		return error
	}

	multiWriter := io.MultiWriter(os.Stdout, file)

	Logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)

	Logger.Println("Logger initialized")

	return nil
}
