package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type UserInfo struct {
	Epoch     int64  `json:"epoch_second"`
	ProblemID string `json:"problem_id"`
	ContestID string `json:"contest_id"`
	UserID    string `json:"user_id"`
	Language  string `json:"language"`
	Result    string `json:"result"`
}

type SlackRequestBody struct {
	Text string `json:"text"`
}

func main() {
	users := usersFromFile("users.txt")
	for _, i := range users {
		result := fetchNewAC(i, time.Now().Unix()-86400)
		text := formatResult(result)
		postSlack(text)
	}
}

func fetchNewAC(users string, bound int64) [][]string {
	var result [][]string
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodGet, "https://kenkoooo.com/atcoder/atcoder-api/results", nil)
	if err != nil {
		log.Fatal(err)
	}
	query := req.URL.Query()
	query.Set("user", users)
	req.URL.RawQuery = query.Encode()

	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	byteArray, _ := ioutil.ReadAll(response.Body)
	data := new([]UserInfo)

	if err := json.Unmarshal(byteArray, data); err != nil {
		log.Fatal(err)
	}

	for _, i := range *data {
		if i.Epoch > bound && i.Result == "AC" {
			result = append(result, []string{i.UserID, i.ProblemID, i.ContestID, i.Language})
		}
	}
	fmt.Println(result)
	return result
}

func formatResult(result [][]string) string {
	var text string
	for _, i := range result {
		url := "<https://atcoder.jp/contests/" + i[2] + "/tasks/" + i[1] + "|" + i[1] + ">"
		text = text + "\n" + i[0] + "さんが" + url + "を" + i[3] + "でACしました！"
	}
	return text
}

func postSlack(text string) {
	if text != "" {
		client := new(http.Client)
		webhookurl, err := ioutil.ReadFile("webhook.txt")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(text)
		slackBody, _ := json.Marshal(SlackRequestBody{Text: text})
		req, _ := http.NewRequest(http.MethodPost,
			string(webhookurl),
			bytes.NewBuffer(slackBody))
		req.Header.Add("Content-Type", "application/json")
		resp, _ := client.Do(req)
		fmt.Println(resp.Status)
	}
}

func usersFromFile(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	var users []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		users = append(users, scanner.Text())
	}

	return users
}
