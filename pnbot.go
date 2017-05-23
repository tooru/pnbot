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

    httpClient := config.Client(oauth1.NoContext, token)
    client := twitter.NewClient(httpClient)

    
    primes := prime.NewPrime()

    pnTweet(client, primes)
}

func pnTweet(client *twitter.Client, primes *prime.Prime) {
    maxPrime, err := getMaxPrime(client)
    if err != nil {
        log.Fatalf("getMaxPrime: %v\n", err)
        return
    }

    ch := make(chan *big.Int, 1000)

    go makePrimes(primes, maxPrime, ch)
    tweetPrimes(client, ch)
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
    }

    for {
        ch <- prime

        prime, err = primes.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            ch <- nil
            return
        }            
    }
}

func tweetPrimes(client *twitter.Client, ch chan *big.Int) {
    for {
        prime := <- ch
        if prime == nil {
            log.Fatalf("prime generator was died")
            return
        }

        text := prime.Text(10)
        _, _, err := client.Statuses.Update(text, nil)
        if err != nil {
            log.Fatalf("Tweet error: %v\n", err)
            return
        }
        log.Printf("tweet %s\n", text)
        time.Sleep(30 * time.Second)
    }
}
