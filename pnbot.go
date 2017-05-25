package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "math/big"
    "time"

    "github.com/tooru/pnbot/prime"
    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
)

const (
    queueSize = 10
    maxRetry = 10
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

    pnbot := NewPNBot(consumerKey, consumerSecret, accessToken, accessSecret)
    pnbot.Start()
}

func NewPNBot(consumerKey *string, consumerSecret *string,
              accessToken *string, accessSecret *string) *PNBot {
    pnbot := &PNBot{
        consumerKey: consumerKey,
        consumerSecret: consumerSecret,
        accessToken: accessToken,
        accessSecret: accessSecret,
        prime: prime.NewPrime(),
        ch: make(chan *big.Int, queueSize),
    }

    return pnbot
}

type PNBot struct {
    consumerKey *string
    consumerSecret *string
    accessToken *string
    accessSecret *string

    prime *prime.Prime
    ch chan *big.Int

    client *twitter.Client

}

func (pnbot *PNBot) Start() error {
    pnbot.client = pnbot.newClient()

    maxPrime, err := pnbot.getMaxPrime()
    if err != nil {
        log.Printf("getMaxPrime: %v\n", err)
        return err
    }

    go pnbot.makePrimes(maxPrime)

    return pnbot.tweetPrimes()
}

func (pnbot *PNBot) newClient() *twitter.Client {
    config := oauth1.NewConfig(*pnbot.consumerKey, *pnbot.consumerSecret)
    token := oauth1.NewToken(*pnbot.accessToken, *pnbot.accessSecret)
    httpClient := config.Client(oauth1.NoContext, token)

    return twitter.NewClient(httpClient)
}

func (pnbot *PNBot) getMaxPrime() (prime *big.Int, err error) {
    tweets, _, err := pnbot.client.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
        SinceID: 0,
    })
    if err != nil {
        return nil, err
    }

    var maxPrime big.Int

    for _, tweet := range tweets {
        if _, ok := maxPrime.SetString(tweet.Text, 10); ok {
            log.Printf("continue from '%s'", tweet.Text)
            return &maxPrime, nil
        }
        log.Printf("skip '%v'", tweet.Text)
    }
    maxPrime.SetInt64(0)
    return &maxPrime, nil
}

func (pnbot *PNBot) makePrimes(maxPrime *big.Int) {
    var prime *big.Int
    var err error

    for {
        prime, err = pnbot.prime.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            pnbot.ch <- nil
            return
        }            

        if prime.Cmp(maxPrime) > 0 {
            break;
        }
        log.Printf("makePrimes: skip %v\n", prime)
    }

    for {
        log.Printf("makePrimes: found %v\n", prime)
        pnbot.ch <- prime

        prime, err = pnbot.prime.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            pnbot.ch <- nil
            return
        }            
    }
}

func (pnbot *PNBot) tweetPrimes() error {
    for {
        prime := <- pnbot.ch
        if prime == nil {
            log.Fatalf("prime generator was died")
            return fmt.Errorf("prime generator was died")
        }

        text := prime.Text(10)
        retry := 0
        for {
            _, _, err := pnbot.client.Statuses.Update(text, nil)
            if err == nil {
                break
            }
            if retry >= maxRetry {
                return fmt.Errorf("Too many tweet error: %v\n", err)
            }
            log.Printf("Tweet error[%d/%d]: %v\n", retry+1, maxRetry, err)
            time.Sleep(5 * (time.Duration(retry) + 1) * time.Minute)
            pnbot.client = pnbot.newClient()
            retry++
            continue
        }
        log.Printf("tweet %s\n", text)
        time.Sleep(15 * time.Second)
    }
}
