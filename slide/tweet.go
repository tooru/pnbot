package main

import (
    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
)

const (
    consumerKey = "" // ⚠️
    consumerSecret = "" // ⚠️
    accessToken = "" // ⚠️
    accessSecret = "" // ⚠️
)    

func main() {
    config := oauth1.NewConfig(consumerKey, consumerSecret)
    token := oauth1.NewToken(accessToken, accessSecret)
    httpClient := config.Client(oauth1.NoContext, token)

    client := twitter.NewClient(httpClient)
    client.Statuses.Update("Hello Twitter", nil) // ここを繰り返す
}
