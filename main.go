package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/mail"
	"os"

	"github.com/mhale/smtpd"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ListenIP   string `yaml:"ListenIP"`
	ListenPort int    `yaml:"ListenPort"`
	Username   string `yaml:"Username"`
	Password   string `yaml:"Password"`
	WebhookURL string `yaml:"WebhookURL"`
	Parameters string `yaml:"Parameters"`
}

const CONFIGFILENAME = "config.yaml"

var globalconfig Config

func ReadConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func authHandler(remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared []byte) (bool, error) {

	//Yeah I know. I hate this too.... I can't pass config as a parameter.
	usernameConf := globalconfig.Username
	passwordConf := globalconfig.Password
	if string(username) == usernameConf && string(password) == passwordConf {
		log.Printf("Auth Sucess from %s\n", remoteAddr.String())
		return true, nil
	} else {
		log.Printf("Auth Fail from %s\n", remoteAddr.String())
		return false, nil

	}

}

func ListenAndServe(handler smtpd.Handler, authHandler smtpd.AuthHandler) error {
	mechs := map[string]bool{"PLAIN": true}
	srv := &smtpd.Server{
		Addr:         fmt.Sprintf("%s:%d", globalconfig.ListenIP, globalconfig.ListenPort),
		Handler:      handler,
		Appname:      "MyServerApp",
		Hostname:     "",
		AuthMechs:    mechs,
		AuthHandler:  authHandler,
		AuthRequired: true,
	}
	return srv.ListenAndServe()
}

func mailHandler(origin net.Addr, from string, to []string, data []byte) error {
	msg, _ := mail.ReadMessage(bytes.NewReader(data))
	subject := msg.Header.Get("Subject")
	log.Printf("Received mail from %s for %s with subject %s", from, to[0], subject)
	return nil
}
func main() {
	c, err := ReadConfig(CONFIGFILENAME)
	if err != nil {
		log.Fatal(err)
	}
	globalconfig = *c

	log.Println("Server Running.....")
	ListenAndServe(mailHandler, authHandler)

}
