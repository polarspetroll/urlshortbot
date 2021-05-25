package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)
/////////////////////////////////////////////////////////
var DB = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDISADDR") + ":6379",
	Password: os.Getenv("REDISPWD"),
	DB:       0,
})
var (
	token  = os.Getenv("TOKEN")
	domain = os.Getenv("DOMAIN")
	help   = `
  *URL Shortner Bot*

  send your long URLs and get short URL
  EG : https://example.com

  _ URL expires after 7 days_`
)

type Res struct {
	Message struct {
		Chat struct {
			FirstName string `json:"first_name"`
			ID        int64  `json:"id"`
			Type      string `json:"type"`
			Username  string `json:"username"`
		} `json:"chat"`
		Date     int64 `json:"date"`
		Entities []struct {
			Length int64  `json:"length"`
			Offset int64  `json:"offset"`
			Type   string `json:"type"`
		} `json:"entities"`
		From struct {
			FirstName    string `json:"first_name"`
			ID           int64  `json:"id"`
			IsBot        bool   `json:"is_bot"`
			LanguageCode string `json:"language_code"`
			Username     string `json:"username"`
		} `json:"from"`
		MessageID int64  `json:"message_id"`
		Text      string `json:"text"`
	} `json:"message"`
	UpdateID int64 `json:"update_id"`
}
/////////////////////////////////////////////////////////

func WebHook(w http.ResponseWriter, r *http.Request) {
	update, err := ioutil.ReadAll(r.Body)
	var out Res
	err = json.Unmarshal(update, &out)
	CheckErr(err)
	url := out.Message.Text
	if len(out.Message.Entities) == 0 {
		sendMessage("*invalid URL*", "Markdown", out.Message.Chat.ID)
		return
	}
	if out.Message.Entities[0].Type == "bot_command" {
		sendMessage(help, "Markdown", out.Message.Chat.ID)
		return
	} else if out.Message.Entities[0].Type != "url" {
		sendMessage("<b>invalid URL</b>", "HTML", out.Message.Chat.ID)
		return
	}
	path := "https://" + domain + "/u/"
	path += Insert(url)
	sendMessage(fmt.Sprintf("[Click](%v)", path), "Markdown", out.Message.Chat.ID)
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	res := Get(r.URL.Path[3:])
	if res == "" {
		http.Error(w, "Not Found", 404)
		return
	}
	http.Redirect(w, r, res, 302)
}

func Insert(url string) string {
	a := make([]byte, 7)
	rand.Read(a)
	path := hex.EncodeToString(a)
	exp, err := time.ParseDuration("168h")
	CheckErr(err)
	err = DB.Set(path, url, exp).Err()
	CheckErr(err)
	return path
}
func Get(path string) (u string) {
	u, err := DB.Get(path).Result()
	switch {
	case err == redis.Nil:
		u = ""
	case err != nil:
		log.Fatal(err)
	case u == "":
		u = ""
	}
	return u
}

func sendMessage(message, parse string, chat int64) {
	u := fmt.Sprintf("bot%v/sendMessage?chat_id=%v&text=%v&parse_mode=%v", token, chat, url.QueryEscape(message), parse)
	_, err := http.Get("https://api.telegram.org/" + u)
	CheckErr(err)
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
