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

type backupPayload struct {
	Settings struct {
		Data string `json:"data"`
	} `json:"settings"`
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
	err := os.Mkdir(config.BackupPath, 0755)
	errAlreadyExists := fmt.Sprintf("mkdir %s: Cannot create a file when that file already exists.", config.BackupPath)
	if err.Error() != errAlreadyExists {
		log.Fatal(err)
	}

	slugs := make(map[string]string)
	content, err := ioutil.ReadFile(config.DataFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(content, &slugs)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for k, v := range slugs {
		wg.Add(1)
		go func(k, v string) {
			defer wg.Done()

			token, err := getAuthToken(k, v)
			if err != nil {
				log.Errorf("%s %v", k, err)
				return
			}

			err = saveBackup(token, k, v)
			if err != nil {
				log.Errorf("%s %v", k, err)
				return
			}

			log.Infof("%s backup works successfully!", k)
		}(k, v)
	}

	wg.Wait()
	log.Info("Finished!")
}

func getAuthToken(k, v string) (string, error) {
	loginURL := fmt.Sprintf(config.BaseURL+"%s", k, "/v1/users/login")
	req, err := http.NewRequest("POST", loginURL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(config.Username, v)
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if res.Status != "200 OK" {
		return "", errors.New("status: " + res.Status)
	}
	bodyText, err := ioutil.ReadAll(res.Body)
	loginPay := loginPayload{}
	err = json.Unmarshal(bodyText, &loginPay)
	if err != nil {
		return "", err
	}

	if len(loginPay.Users) < 1 {
		return "", errors.New("No users")
	}

	return loginPay.Users[0].Token, nil
}

func saveBackup(token, k, v string) error {
	requestBody := strings.NewReader(fmt.Sprintf(`{"password": "%s"}`, v))
	backupURL := fmt.Sprintf(config.BaseURL+"%s", k, "/v1/settings/backup")
	req, err := http.NewRequest("POST", backupURL, requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if res.Status != "200 OK" {
		return errors.New("status: " + res.Status)
	}
	bodyText, err := ioutil.ReadAll(res.Body)
	backupPay := backupPayload{}
	err = json.Unmarshal(bodyText, &backupPay)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf(config.BackupPath+"/%s", k), []byte(fmt.Sprint(backupPay.Settings.Data)), 0644)
	if err != nil {
		return err
	}

	return nil
}
