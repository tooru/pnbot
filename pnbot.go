package main

import (
    "flag"
    "fmt"
    "log"
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
    maxRetry = 100
    retryInterval = 10 * time.Minute
    interval = 1 * time.Second

    errorUserIsOverDailyStatusUpdateLimit = 185
    errorStatusIsADuplicate = 187
)

func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

    flags := flag.NewFlagSet("pnbot", flag.ExitOnError)

    mode := flags.String("mode", "normal", "Tweet Mode: 'normal'(default), 'twin', 'primep', 'primeptest'")
    target := flags.String("target", "", "tweet target")

    consumerKey := flags.String("ck", "", "Twitter Consumer Key")
    consumerSecret := flags.String("cs", "", "Twitter Consumer Secret")
    accessToken := flags.String("at", "", "Twitter Access Token")
    accessSecret := flags.String("as", "", "Twitter Access Secret")

    debug := flags.Bool("debug", false, "debug mode")

    flags.Parse(os.Args[1:])

    if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
        log.Fatal("Consumer key/secret and Access token/secret required")
    }

    pnbot := NewPNBot(mode, target, *debug,
        consumerKey, consumerSecret, accessToken, accessSecret)
    pnbot.Start()
}

func NewPNBot(mode *string, target *string, debug bool,
              consumerKey *string, consumerSecret *string,
              accessToken *string, accessSecret *string) *PNBot {
    pnbot := &PNBot{
        mode: mode,
        target: target,
        debug: debug,
        consumerKey: consumerKey,
        consumerSecret: consumerSecret,
        accessToken: accessToken,
        accessSecret: accessSecret,
        prime: prime.NewPrime(),
    }

    return pnbot
}

type PNBot struct {
    mode *string
    target *string
    debug bool

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

    switch *pnbot.mode {
    case "normal":
        return pnbot.startNormal()
    case "twin":
        return pnbot.startTwin()
    case "primep":
        return pnbot.startPrimeP()
    case "primeptest":
        return nil
        //return pnbot.startTestPrimeP()
    default:
        return fmt.Errorf("Start: unknown mode: %v", *pnbot.mode)
    }
}

func (pnbot *PNBot) newClient() *twitter.Client {
    config := oauth1.NewConfig(*pnbot.consumerKey, *pnbot.consumerSecret)
    token := oauth1.NewToken(*pnbot.accessToken, *pnbot.accessSecret)
    httpClient := config.Client(oauth1.NoContext, token)

    return twitter.NewClient(httpClient)
}

func (pnbot *PNBot) tweet(tweets chan *PNTweet, quit chan interface{}) error {
    totalN := 0
    totalStart := time.Now()

    contN := 0
    contStart := totalStart

    for {
        var pnTweet *PNTweet

        select {
        case pnTweet = <- tweets:
        case <- quit:
            log.Fatalf("error occurred.")
        }

        retry := 0
        for {
            var err error
            if pnbot.debug {
                log.Printf("Tweet: %s", pnTweet.text)
            } else {
                _, _, err = pnbot.client.Statuses.Update(pnTweet.text, pnTweet.params)
            }
            if err == nil {
                break
            }
            v, ok := err.(twitter.APIError)
            if ok {
                if len(v.Errors) == 0 {
                    return fmt.Errorf("Unknown Error:len(Errors)=0: %v\n", err)
                }
                errDetail := v.Errors[0]
                if errDetail.Code == errorStatusIsADuplicate {
                    log.Printf(err.Error())
                    break
                } else if errDetail.Code != errorUserIsOverDailyStatusUpdateLimit {
                    return fmt.Errorf("Unknown Error:len(Errors)=0: %v\n", err)
                }
            }

            if retry >= maxRetry {
                return fmt.Errorf("Too many tweet error: %v\n", err)
            }
            log.Printf("Tweet error[%d/%d]:sleep=%s: %v\n", retry+1, maxRetry, retryInterval, err)

            time.Sleep(retryInterval)
            pnbot.client = pnbot.newClient()
            retry++
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
        time.Sleep(interval)
    }
}

func (pnbot *PNBot) startNormal() error {
    tweet, err := pnbot.lastTweet()
    if err != nil {
        return err
    }
    log.Printf("last tweet: %v", tweet)

    maxPrime := new(big.Int)
    if tweet != "" {
        maxPrime.SetString(tweet, 10)
    } else {
        maxPrime.SetInt64(0)
    }


    primes := make(chan *PNTweet, 10)
    quit := make(chan interface{})

    go pnbot.makePrimes(maxPrime, primes, quit)

    return pnbot.tweet(primes, quit)
}

func (pnbot *PNBot) startTwin() error {
    tweet, err := pnbot.lastTweet()
    if err != nil {
        return err
    }
    log.Printf("last tweet: %v", tweet)

    maxPrime := new(big.Int)
    if tweet != "" {
        twin := strings.Split(tweet, ",")
        if len(twin) != 2 {
            return fmt.Errorf("invalid tweet len: %v", tweet)
        }
        _, ok := maxPrime.SetString(twin[1], 10)
        if !ok {
            return fmt.Errorf("invalid tweet: %v", tweet)
        }
    } else {
        maxPrime.SetInt64(3)
    }


    primes := make(chan *PNTweet, 10)
    quit := make(chan interface{})

    go pnbot.makeTwinPrimes(maxPrime, primes, quit)

    return pnbot.tweet(primes, quit)
}

func (pnbot *PNBot) lastTweet() (lastTweet string, err error) {
    params := twitter.UserTimelineParams{
        SinceID: 0,
    }

    tweets, _, err := pnbot.client.Timelines.UserTimeline(&params)
    if err != nil {
        log.Printf("lastTweet: %v\n", err)
        return "", err
    }

    if len(tweets) == 0 {
        log.Printf("lastTweet:len=0\n")
        return "", nil
    }

    return tweets[0].Text, nil
}

func (pnbot *PNBot) makePrimes(maxPrime *big.Int, tweets chan *PNTweet, quit chan interface{}) {
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
        tweets <- &PNTweet{
            text: prime.Text(10),
        }

        prime, err = pnbot.prime.Next()
        if err != nil {
            log.Fatalf("makePrimes: %v\n", err)
            quit <- nil
            return
        }
    }
}

func (pnbot *PNBot) makeTwinPrimes(maxPrime *big.Int, tweets chan *PNTweet, quit chan interface{}) {
    var prevPrime *big.Int
    var prime *big.Int
    var err error

    big2 := big.NewInt(2)

    log.Printf("makeTwinPrimes: %v\n", maxPrime)

    for {
        prevPrime, err = pnbot.prime.Next()
        if err != nil {
            log.Fatalf("makeTwinPrimes: %v\n", err)
            quit <- nil
            return
        }

        if prevPrime.Cmp(maxPrime) >= 0 {
            break;
        }
    }

    for {
        for {
            prime, err = pnbot.prime.Next()
            if err != nil {
                log.Fatalf("makeTwinPrimes: %v\n", err)
                quit <- nil
                return
            }

            n := new(big.Int)
            if n.Sub(prime, prevPrime).Cmp(big2) == 0 {
                log.Printf("makeTwinPrimes: found %v,%v\n", prevPrime, prime)

                break
            }
            prevPrime = prime
        }

        tweets <- &PNTweet{
            text: fmt.Sprintf("%s,%s", prevPrime.Text(10), prime.Text(10)),
        }
        prevPrime = prime
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

func (pnbot *PNBot) startPrimeP() error {

    lastReplyID, err := pnbot.lastReplyID()
    if err != nil {
        return err
    }

    tweets := make(chan *PNTweet, 10)
    quit := make(chan interface{})

    go pnbot.reply(tweets, quit, lastReplyID)

    return pnbot.tweet(tweets, quit)
}

func (pnbot *PNBot) lastReplyID() (lastReplyID int64, err error) {
    params := twitter.UserTimelineParams{
        Count: 200,
        SinceID: 0,
    }

    n := 0

    for {
        tweets, _, err := pnbot.client.Timelines.UserTimeline(&params)
        if err != nil {
            log.Printf("lastReplyiD: %v\n", err)
            time.Sleep(time.Minute)
            continue
        }

        if len(tweets) == 0 {
            return -1, nil
        }

        for _, tweet := range tweets {
            params.MaxID = tweet.ID - 1

            if tweet.InReplyToStatusID != 0 {
                return tweet.InReplyToStatusID, nil
            }
        }
        n++
        log.Printf("get next tweets:%d", n)
        time.Sleep(10 * interval)
    }
}

func (pnbot *PNBot) reply(tweets chan *PNTweet, quit chan interface{}, lastReplyID int64) error {
    var prevID int64 = -1
    var id int64 = lastReplyID + 1

    var mentionInterval time.Duration = 10 * time.Second
    var totalCount int = 0;
    var count int = 0;

    var retry int = 0

    for {
        if prevID != id {
            log.Printf("lastReplyID: %d", id)
        }
        params := twitter.MentionTimelineParams{
            SinceID: id,
        }
        mentions := []twitter.Tweet{}
        retry = 0
        for {
            ms, _, err := pnbot.client.Timelines.MentionTimeline(&params)

            if err != nil {
                if retry >= maxRetry {
                    quit <- nil
                    return fmt.Errorf("Too many tweet error: %v\n", err)
                }
                log.Printf("MentionTimeline error[%d/%d]:sleep=%s: %v\n", retry+1, maxRetry, retryInterval, err)

                time.Sleep(retryInterval)
                pnbot.client = pnbot.newClient()
                count = 0
                retry++
                continue
            }
            if len(ms) == 0 {
                break
            }

            mentions = append(mentions, ms...)
            params.MaxID = ms[len(ms)-1].ID - 1
        }

        reverse(mentions)

        for _, mention := range mentions {
            n, ok := parseTweet(mention)

            if !ok {
                continue
            }

            b, err := pnbot.prime.IsPrime(n)
            if err != nil {
                log.Printf("%v", err)

                err := replyPrimeImpl(tweets, n, "timeout", mention)

                if err != nil {
                    quit <- nil
                    return err
                }
                continue
            }

            if b {
                err = replyPrimeImpl(tweets, n, "prime number", mention)
            } else {
                err = replyPrimeImpl(tweets, n, "composite number", mention)
            }
            if err != nil {
                quit <- nil
                return err
            }
        }
        prevID = id
        if len(mentions) > 0 {
            id = mentions[len(mentions)-1].ID + 1
        }
        count++
        totalCount++

        log.Printf("Sleep %v:count=%d:totalCount=%d", mentionInterval, count, totalCount)
        time.Sleep(mentionInterval)
    }
}

func replyPrimeImpl(replies chan *PNTweet, n *big.Int, result string, tweet twitter.Tweet) error {
    var text string

    pnText := n.Text(10)

    for {
        text = fmt.Sprintf("@%s %v: %s", tweet.User.ScreenName, pnText, result)
        log.Printf("%s %d", text, len(text))
        if len(text) <= maxTweetCharactors {
            break
        }
        pnText = pnText[0:(len(pnText)+maxTweetCharactors-len(text)-len(hellipsis))]
        pnText = pnText + hellipsis
    }


    replies <- &PNTweet{
        text: text,
        params: &twitter.StatusUpdateParams{
            InReplyToStatusID: tweet.ID,
        },
    }
    time.Sleep(1 * time.Second)
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
