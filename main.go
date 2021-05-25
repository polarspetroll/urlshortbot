package main

import (
	"app"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

type whook struct {
	Description string `json: "description"`
}

func main() {
	token := os.Getenv("TOKEN")
	webhookurl := "https://" + os.Getenv("DOMAIN") + "/bot/" + token + "&allow_updates=" + `["message"]`
	res, err := http.Get("https://api.telegram.org/bot" + token + "/setWebhook?url=" + webhookurl)
	app.CheckErr(err)
	defer res.Body.Close()
	j, err := ioutil.ReadAll(res.Body)
	app.CheckErr(err)
	var stat whook
	err = json.Unmarshal(j, &stat)
	app.CheckErr(err)
	println(stat.Description)
	http.HandleFunc("/bot/"+token, app.WebHook)
	http.HandleFunc("/u/", app.GetURL)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
