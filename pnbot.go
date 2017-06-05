package main

import (
    "flag"
    "fmt"
    "log"
    "encoding/json"
    "os"
    "math/big"
    "sort"
    "strings"
    "time"
    "unicode"

    "github.com/tooru/pnbot/prime"
    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
//    "github.com/davecgh/go-spew/spew"
)

const (
    maxTweetCharactors = 140
    hellipsis = "\u2026"

    queueSize = 10
    maxRetry = 30
    interval = 1 * time.Second
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
    }

    return pnbot
}

type PNBot struct {
    consumerKey *string
    consumerSecret *string
    accessToken *string
    accessSecret *string

    prime *prime.Prime

    client *twitter.Client

}

type PNTweet struct {
    text string
    params *twitter.StatusUpdateParams
}

func (pnbot *PNBot) Start() error {
    pnbot.client = pnbot.newClient()

    maxPrime, lastReplyID, err := pnbot.findLastUpdated()
    if err != nil {
        return err
    }
    log.Printf("%v %v", maxPrime, lastReplyID)

    primes := make(chan *PNTweet, 10)
    replies := make(chan *PNTweet, 10)
    quit := make(chan interface{})

    go pnbot.makePrimes(maxPrime, primes, quit)
    go pnbot.replyPrime(lastReplyID, replies, quit)

    return pnbot.tweet(primes, replies, quit)
}

func (pnbot *PNBot) newClient() *twitter.Client {
    config := oauth1.NewConfig(*pnbot.consumerKey, *pnbot.consumerSecret)
    token := oauth1.NewToken(*pnbot.accessToken, *pnbot.accessSecret)
    httpClient := config.Client(oauth1.NoContext, token)

    return twitter.NewClient(httpClient)
}

func (pnbot *PNBot) makePrimes(maxPrime *big.Int, primes chan *PNTweet, quit chan interface{}) {
    var prime *big.Int
    var err error

    log.Printf("makePrime: %v\n", maxPrime)

    for {
        prime, err = pnbot.prime.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            quit <- nil
            return
        }

        if prime.Cmp(maxPrime) > 0 {
            break;
        }
        //log.Printf("makePrimes: skip %v\n", prime)
    }

    for {
        log.Printf("makePrimes: found %v\n", prime)
        //primes <- &PNTweet{
        //text: prime.Text(10),
        //}

        prime, err = pnbot.prime.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            quit <- nil
            return
        }
        time.Sleep(600 * time.Second)
    }
}

func (pnbot *PNBot) tweet(primes chan *PNTweet, replies chan *PNTweet, quit chan interface{}) error {
    totalN := 0
    totalStart := time.Now()

    contN := 0
    contStart := totalStart

    for {
        var pnTweet *PNTweet

        select {
        case pnTweet = <- primes:
        case pnTweet = <- replies:
        case <- quit:
            log.Fatalf("error occurred.")
        }

        retry := 0
        retryInterval := 10 * time.Second
        for {
            //err := error(nil);
            log.Printf("TWEET: %s, %v", pnTweet.text, pnTweet.params)
            _, _, err := pnbot.client.Statuses.Update(pnTweet.text, pnTweet.params)
            if err == nil {
                break
            }
            if retry >= maxRetry {
                return fmt.Errorf("Too many tweet error: %v\n", err)
            }
            log.Printf("Tweet error[%d/%d]:sleep=%s: %v\n", retry+1, maxRetry, retryInterval, err)

            time.Sleep(retryInterval)
            pnbot.client = pnbot.newClient()
            retry++
            retryInterval = max(retryInterval * 2, 30 * time.Minute)
            continue
        }
        now := time.Now()

        if retry > 0 {
            contN = 0
            contStart = now
        }
        contN++
        totalN++

        log.Printf("tweet:%d:%d:%s:sleep=%s:%.2f tw/h: %.2f tw/h\n",
            totalN, contN,
            pnTweet.text,
            interval,
            float64(totalN)/(float64(now.Sub(totalStart))/float64(time.Hour)),
            float64(contN)/(float64(now.Sub(contStart))/float64(time.Hour)))
    }
}

func (pnbot *PNBot) findLastUpdated() (maxPrime *big.Int, lastReplyID int64, err error) {
    params := twitter.HomeTimelineParams{
        Count: 200,
        SinceID: 0,
    }

    n := 0

    for {
        tweets, _, err := pnbot.client.Timelines.HomeTimeline(&params)
        if err != nil {
            log.Printf("findLastUpdated: %v\n", err)
            time.Sleep(time.Minute)
            continue
        }

        var prime big.Int

        if len(tweets) == 0 {
            if maxPrime == nil {
                maxPrime = big.NewInt(0)
            }
            return maxPrime, lastReplyID, nil
        }

        for _, tweet := range tweets {
            //log.Printf("tweet:%d", tweet.ID)
            params.MaxID = tweet.ID - 1

            if tweet.InReplyToStatusID == 0 {
                if maxPrime != nil {
                    continue
                }
                if _, ok := prime.SetString(tweet.Text, 10); ok {
                    log.Printf("continue from '%s'", tweet.Text)
                    maxPrime = &prime
                    if lastReplyID != 0 {
                        return maxPrime, lastReplyID, nil
                    }
                }
                log.Printf("skip '%v'", tweet.Text)
            } else {
                if lastReplyID != 0 {
                    continue
                }
                lastReplyID = tweet.InReplyToStatusID
                if maxPrime != nil {
                    return maxPrime, lastReplyID, nil
                }
            }
        }
        n++
        log.Printf("get next tweets:%d", n)
        time.Sleep(10 * interval)
    }
}

func max(a time.Duration, b time.Duration) time.Duration {
    if a > b {
        return b
    } else {
        return a
    }
}

type Response struct {
    Method string `json:"method"`
    Params interface{} `json:"params"`
    Result interface{} `json:"result"`
}

func (pnbot *PNBot) replyPrime(lastReplyID int64, replies chan *PNTweet, quit chan interface{}) error {
    for {
        params := twitter.MentionTimelineParams{
            SinceID: lastReplyID + 1,
        }
        mentions := []twitter.Tweet{}
        for {
            tweets, _, err := pnbot.client.Timelines.MentionTimeline(&params)

            if err != nil {
                return err
            }
            if len(tweets) == 0 {
                break
            }

            mentions = append(mentions, tweets...)
            params.MaxID = tweets[len(tweets)-1].ID - 1
        }

        reverse(mentions)

        for _, tweet := range mentions {
            n, ok := parseTweet(tweet)

            if !ok {
                continue
            }

            b, err := pnbot.prime.IsPrime(n)
            if err != nil {
                log.Printf("%v", err)

                err := replyPrimeImpl(replies, n, "unknown", tweet)

                if err != nil {
                    quit <- nil
                    return err
                }
                continue
            }

            if b {
                err = replyPrimeImpl(replies, n, "primeNumber", tweet)
            } else {
                err = replyPrimeImpl(replies, n, "notPrimeNumber", tweet)
            }
            if err != nil {
                quit <- nil
                return err
            }
        }
        if len(mentions) > 0 {
            lastReplyID = mentions[len(mentions)-1].ID + 1
        }
        time.Sleep(10 * time.Second)
    }
}

func replyPrimeImpl(replies chan *PNTweet, n *big.Int, result string, tweet twitter.Tweet) error {
    var text string

    pnText := n.Text(10)

    for {
        bytes, err := json.Marshal(Response{
            Method: "isPrime",
            Params: pnText,
            Result: result,
        })

        if err != nil {
            return err
        }

        text = fmt.Sprintf("@%s %s", tweet.User.ScreenName, string(bytes))
        log.Printf("%v %v", text, len(text))
        if len(text) <= maxTweetCharactors {
            break
        }
        pnText = pnText[0:(len(pnText)+maxTweetCharactors-len(text)-len(hellipsis))]
        pnText = pnText + hellipsis
    }


    //replies <- &PNTweet{
    //    text: text,
    //    params: &twitter.StatusUpdateParams{
    //        InReplyToStatusID: tweet.ID,
    //    },
    //}
    time.Sleep(10 * time.Second)
    return nil
}

func parseTweet(tweet twitter.Tweet) (*big.Int, bool) {
    text := tweet.Text
    i := 0

    log.Printf("'%s' ", text)

    for _, indices := range getEntityIndices(&tweet) {
        if i == indices.Start() {
            i = indices.End()
            continue
        }
        str := text[i:indices.Start()]
        log.Printf("'%s'\n", str)
        if n, ok := parseNumber(str); ok {
            return n, true
        }
        i = indices.End()
    }

    if i < len(text) {
        str := text[i:]
        log.Printf("'%s'\n", str)
        if n, ok := parseNumber(str); ok {
            return n, true
        }
    }

    log.Printf("ignored\n")
    return nil, false
}

func parseNumber(str string) (*big.Int, bool) {
    n := new(big.Int)
    if _, ok := n.SetString(strings.TrimFunc(str, unicode.IsSpace), 10); ok {
        return n, true
    }
    return nil, false

}

type indices []twitter.Indices

func (is indices) Len() int {
    return len(is)
}

func (is indices) Swap(i, j int) {
    is[i], is[j] = is[j], is[i]
}

func (is indices) Less(i, j int) bool {
    return is[i].Start() < is[j].Start()
}

func getEntityIndices(tweet *twitter.Tweet) []twitter.Indices {
    var res indices = make([]twitter.Indices, 0)

    entities := tweet.Entities

    for _, entity := range entities.Hashtags {
        res = append(res, entity.Indices)
    }
    for _, entity := range entities.Media {
        res = append(res, entity.Indices)
    }
    for _, entity := range entities.Urls {
        res = append(res, entity.Indices)
    }
    for _, entity := range entities.UserMentions {
        res = append(res, entity.Indices)
    }

    sort.Sort(res)

    return res
}

func reverse(tweets []twitter.Tweet) {
    for i, j := 0, len(tweets)-1; i < j; i, j = i + 1, j - 1 {
        tweets[i], tweets[j] = tweets[j], tweets[i]
    }
}
