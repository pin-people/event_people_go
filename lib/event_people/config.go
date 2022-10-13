package EventPeople

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type configStruct struct {
	APP_NAME string
	TOPIC    string
	VHOST    string
	URL      string
	FULL_URL string
	Broker   AbstractBaseBroker
}

var Config configStruct

func ConfigEnvs() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	Config.APP_NAME = os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	Config.TOPIC = os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME")
	Config.VHOST = os.Getenv("RABBIT_EVENT_PEOPLE_VHOST")
	Config.URL = os.Getenv("RABBIT_URL")
	Config.FULL_URL = fmt.Sprintf("%s/%s", Config.URL, Config.VHOST)

	if Config.APP_NAME == "" || Config.TOPIC == "" || Config.VHOST == "" || Config.URL == "" {
		FailOnError(errors.New("empty envs"), "Please set environment variables")
		return
	}

	fmt.Printf("Configuration successfull executed!\nEnvs successfull loaded!\n")
}

func (config *configStruct) InitBroker(broker AbstractBaseBroker) {
	ConfigEnvs()
	config.Broker = broker
	config.Broker.Init()
}
