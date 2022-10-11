package Config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	Utils "github.com/pinpeople/event_people_go/lib/event_people/utils"
)

var (
	APP_NAME = ""
	TOPIC    = ""
	VHOST    = ""
	URL      = ""
	FULL_URL = ""
)

func ConfigEnvs() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	APP_NAME = os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	TOPIC = os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME")
	VHOST = os.Getenv("RABBIT_EVENT_PEOPLE_VHOST")
	URL = os.Getenv("RABBIT_URL")
	FULL_URL = fmt.Sprintf("%s/%s", URL, VHOST)

	if APP_NAME == "" || TOPIC == "" || VHOST == "" || URL == "" {
		Utils.FailOnError(errors.New("empty envs"), "Please set environment variables")
		return
	}

	fmt.Printf("Configuration successfull executed!\nEnvs successfull loaded!\n")
}
