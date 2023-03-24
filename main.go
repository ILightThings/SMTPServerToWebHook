package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"os"
	"strings"

	"github.com/mhale/smtpd"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ListenIP   string `yaml:"ListenIP"`
	ListenPort int    `yaml:"ListenPort"`
	Username   string `yaml:"Username"`
	Password   string `yaml:"Password"`
	WebhookURL string `yaml:"WebhookURL"`
	Parameters map[string]string `yaml:"Parameters"`
}

const CONFIGFILENAME = "config.yaml"

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

func makeWebhookRequest(msg string, config *Config) error {

	jsonStr, err := json.Marshal(config.Parameters)
	if err != nil{
		log.Panic(err)
	}

	req, err := http.NewRequest(http.MethodPost, config.WebhookURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	client := &http.Client{}
	bby, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(bby.Body)
	if err != nil {
		return err
	}
	fmt.Println(body)

	fmt.Println("Webhook Sent")
	return nil

}

func authHandler(remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared []byte, config *Config) (bool, error) {

	if string(username) == config.Username && string(password) == config.Password {
		log.Printf("Auth Sucess from %s\n", remoteAddr.String())
		return true, nil
	} else {
		log.Printf("Auth Fail from %s\n", remoteAddr.String())
		return false, nil

	}

}

func mailHandler(origin net.Addr, from string, to []string, data []byte, config *Config) error {
	msg, _ := mail.ReadMessage(bytes.NewReader(data))
	subject := msg.Header.Get("Subject")
	log.Printf("Received mail from %s for %s with subject %s", from, to[0], subject)
	buf := new(strings.Builder)
	_, err := io.Copy(buf, msg.Body)
	if err != nil {
		return err
	}
	err = makeWebhookRequest(subject, config)
	if err != nil {
		fmt.Printf("There is an error sending webhook %s\n", err.Error())
		return err
	}

	return nil
}


func ListenAndServe(config *Config) error {
	mechs := map[string]bool{"PLAIN": true}
	srv := &smtpd.Server{
		Addr:         fmt.Sprintf("%s:%d", config.ListenIP, config.ListenPort),
		Handler:      func (origin net.Addr, from string, to []string, data []byte) error { return mailHandler(origin, from, to, data, config) },
		Appname:      "MyServerApp",
		Hostname:     "",
		AuthMechs:    mechs,
		AuthHandler:  func (remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared[]byte)(bool, error){ return authHandler(remoteAddr, mechanism, username, password, shared, config) },
		AuthRequired: config.Username != "",
	}
	return srv.ListenAndServe()
}

func main() {
	c, err := ReadConfig(CONFIGFILENAME)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server Running.....")
	if c.Username != "" {
		log.Println("Authentication required")
	}else{
		log.Println("Anonymous access permitted")
	}
	ListenAndServe(c)

}
