package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	{
		for _, i := range users {
			result := fetchNewAC(i, time.Now().Unix()-(60*60*24))
			text := formatTextForSlack(result)
			postSlack(text)
		}
	}
}

func fetchNewAC(user string, bound int64) []*UserInfo {
	var req *http.Request
	{
		req = createRequest(http.MethodGet, "https://kenkoooo.com/atcoder/atcoder-api/results", bytes.NewBuffer(nil))
		query := req.URL.Query()
		query.Set("user", user)
		req.URL.RawQuery = query.Encode()
	}

	client := new(http.Client)
	response, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	rb, _ := ioutil.ReadAll(response.Body)
	return formatResult(rb, bound)

}
func formatResult(fetchResult []byte, bound int64) []*UserInfo {
	var data []*UserInfo
	{
		if err := json.Unmarshal(fetchResult, &data); err != nil {
			log.Fatal(err)
		}
	}

	var result []*UserInfo
	{
		for _, i := range data {
			if i.Epoch > bound && i.Result == "AC" {
				result = append(result, i)
			}
		}
	}

	return result
}

func formatTextForSlack(result []*UserInfo) string {
	var text string
	{
		for _, i := range result {
			url := "<https://atcoder.jp/contests/" + i.ContestID + "/tasks/" + i.ProblemID + "|" + i.ProblemID + ">"
			text += "\n" + i.UserID + "さんが" + url + "を" + i.Language + "でACしました！"
		}
	}
	return text
}

func postSlack(text string) {
	if text == "" {
		return
	}

	var req *http.Request
	{
		webhookURL, err := ioutil.ReadFile("webhook.txt")
		if err != nil {
			log.Fatal(err)
		}

		slackBody, _ := json.Marshal(SlackRequestBody{Text: text})
		req = createRequest(http.MethodPost,
			string(webhookURL),
			bytes.NewBuffer(slackBody))

		req.Header.Add("Content-Type", "application/json")
	}

	var client *http.Client
	{
		client = new(http.Client)
		_, _ = client.Do(req)
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

func createRequest(method, url string, body *bytes.Buffer) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatal(err)
	}

	return req
}
