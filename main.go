package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type loginPayload struct {
	Users []struct {
		Token string `json:"token"`
	} `json:"users"`
}

type data struct {
	URL      string `json:"url"`
	FileName string `json:"filename"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05",
	})
}

// /v1/users/login
func main() {
	err := os.Mkdir("./backups", 0755)
	errAlreadyExists := "mkdir ./backups: Cannot create a file when that file already exists."
	if err.Error() != errAlreadyExists {
		log.Fatal(err)
	}

	channelData := []data{}
	content, err := ioutil.ReadFile("./data.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(content, &channelData)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, c := range channelData {
		wg.Add(1)
		go func(c data) {
			defer wg.Done()

			token, err := getAuthToken(c)
			if err != nil {
				log.Errorf("%s - getting authtoken - %v", c.FileName, err)
				return
			}

			err = saveBackup(token, c)
			if err != nil {
				log.Errorf("%s - saving backup - %v", c.FileName, err)
				return
			}

			log.Infof("%s backup works successfully!", c.FileName)
		}(c)
	}

	wg.Wait()
	log.Info("Finished!")
}

func getAuthToken(c data) (string, error) {
	loginURL := c.URL + "/v1/users/login"
	req, err := http.NewRequest("POST", loginURL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.Status != "200 OK" {
		return "", errors.New("status: " + res.Status)
	}
	bodyText, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	loginPay := loginPayload{}
	if err = json.Unmarshal(bodyText, &loginPay); err != nil {
		return "", err
	}

	if len(loginPay.Users) < 1 {
		return "", errors.New("No users")
	}

	return loginPay.Users[0].Token, nil
}

func saveBackup(token string, c data) error {
	requestBody := strings.NewReader(fmt.Sprintf(`{"password": "%s"}`, c.Password))
	backupURL := c.URL + "/v1/settings/backup"
	req, err := http.NewRequest("POST", backupURL, requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.Status != "200 OK" {
		return errors.New("status: " + res.Status)
	}
	bodyText, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fmt.Sprintf("./backups/%s", c.FileName), bodyText, 0644); err != nil {
		return err
	}

	return nil
}
