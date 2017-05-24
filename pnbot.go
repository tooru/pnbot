package main

import (
    "flag"
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

    config := oauth1.NewConfig(*consumerKey, *consumerSecret)
    token := oauth1.NewToken(*accessToken, *accessSecret)
    
    primes := prime.NewPrime()

    pnTweet(config, token, primes)
}

func newClient(config *oauth1.Config, token *oauth1.Token) *twitter.Client {
    httpClient := config.Client(oauth1.NoContext, token)

    return twitter.NewClient(httpClient)
}

func pnTweet(config *oauth1.Config, token *oauth1.Token, primes *prime.Prime) {
    client := newClient(config, token)

    maxPrime, err := getMaxPrime(client)
    if err != nil {
        log.Fatalf("getMaxPrime: %v\n", err)
        return
    }

    ch := make(chan *big.Int, queueSize)

    go makePrimes(primes, maxPrime, ch)

    tweetPrimes(client, config, token, ch)
}

func getMaxPrime(client *twitter.Client) (prime *big.Int, err error) {
    tweets, _, err := client.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
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

func makePrimes(primes *prime.Prime, maxPrime *big.Int, ch chan *big.Int) {
    var prime *big.Int
    var err error

    for {
        prime, err = primes.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            ch <- nil
            return
        }            

        if prime.Cmp(maxPrime) > 0 {
            break;
        }
        log.Printf("makePrimes: skip %v\n", prime)
    }

    for {
        ch <- prime

        prime, err = primes.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            ch <- nil
            return
        }            
        log.Printf("makePrimes: found %v\n", prime)
    }
}

func tweetPrimes(client *twitter.Client, config *oauth1.Config, token *oauth1.Token, ch chan *big.Int) {
    for {
        prime := <- ch
        if prime == nil {
            log.Fatalf("prime generator was died")
            return
        }

        text := prime.Text(10)
        retry := 0
        for {
            _, _, err := client.Statuses.Update(text, nil)
            if err == nil {
                break
            }
            if retry >= maxRetry {
                log.Fatalf("Too many tweet error: %v\n", err)
                return
            }
            log.Printf("Tweet error: %v\n", err)
            time.Sleep(5 * (time.Duration(retry) + 1) * time.Minute)
            client = newClient(config, token)
            retry++
            continue
        }
        log.Printf("tweet %s\n", text)
        time.Sleep(15 * time.Second)
    }
}
