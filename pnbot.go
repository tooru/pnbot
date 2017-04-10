package main

import (
    "flag"
    "fmt"
    "log"
    "os"

//    "github.com/tooru/pnbot/prime"
    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
)

func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

    flags := flag.NewFlagSet("pnbot", flag.ExitOnError)
    consumerKey := flags.String("ck", "", "Twitter Consumer Key")
    consumerSecret := flags.String("cs", "", "Twitter Consumer Secret")
    accessToken := flags.String("at", "", "Twitter Access Token")
    accessSecret := flags.String("as", "", "Twitter Access Secret")
    flags.Parse(os.Args[1:])

    if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
        log.Fatal("Consumer key/secret and Access token/secret required")
    }

    config := oauth1.NewConfig(*consumerKey, *consumerSecret)
    token := oauth1.NewToken(*accessToken, *accessSecret)

    tweet(config, token, "hello twitter")
}

func tweet(config *oauth1.Config, token *oauth1.Token, status string) {
    httpClient := config.Client(oauth1.NoContext, token)
    client := twitter.NewClient(httpClient)
    tweet, _, err := client.Statuses.Update(status, nil)
    if err != nil {
        log.Fatalf("Tweet error: %v\n", err)
        return
    }
    fmt.Println(tweet)
}
