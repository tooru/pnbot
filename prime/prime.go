package prime

import (
    "errors"
    "log"
    "math/big"
    "sort"
    "sync"
    "time"
)

const (
    timeout = 5 * time.Second
)

var big0 = big.NewInt(0)
var big2 = big.NewInt(2)

var maxCacheNumber = big.NewInt(1000 * 1000)

type Prime struct {
    primes []*big.Int
    index int
    maxPrime *big.Int
    mutex sync.Mutex
}

func NewPrime() *Prime {
    prime := &Prime{
        primes: []*big.Int{
            big.NewInt(2),
            big.NewInt(3),
        },
        maxPrime: big.NewInt(2),
    }
    go prime.start()

    return prime
}

func (prime *Prime) start() {
    for p := big.NewInt(3); p.Cmp(maxCacheNumber) < 0; {
        p.Add(p, big2)

        b, err := prime.isPrime(p, time.Now().Add(timeout))

        if err != nil {
            log.Printf("%v", err)
            return
        }

        if !b {
            continue
        }
        prime.mutex.Lock()
        prime.primes = append(prime.primes, newInt(p))
        prime.mutex.Unlock()
    }
}

func (prime *Prime) Next() (*big.Int, error) {
    prime.mutex.Lock()
    primes := prime.primes[:]
    prime.mutex.Unlock()

    if prime.index < len(primes) {
        p := primes[prime.index]
        prime.index++
        return p, nil
    }

    expire := time.Now().Add(timeout)

    lastPrime := last(primes)

    if lastPrime.Cmp(prime.maxPrime) > 0 {
        prime.maxPrime = newInt(lastPrime)
    }
    q := newInt(prime.maxPrime)
    q.Add(q, big2)
    for {
        b, err := prime.isPrime(q, expire)
        if err != nil {
            return nil, err
        }
        if b {
            prime.maxPrime = q
            prime.index++
            return newInt(q), nil
        }
        q.Add(q, big2)
    }
}

func (prime *Prime) IsPrime(n *big.Int) (bool, error) {
    return prime.isPrime(n, time.Now().Add(timeout))
}

func (prime *Prime) isPrime(n *big.Int, expire time.Time) (bool, error) {
    prime.mutex.Lock()
    primes := prime.primes[:]
    prime.mutex.Unlock()
    if n.Cmp(last(primes)) > 0 {
        return isPrime(n, primes, expire)
    }

    prime.mutex.Lock()
    defer prime.mutex.Unlock()

    i := sort.Search(len(prime.primes), func(i int) bool {
        return prime.primes[i].Cmp(n) >= 0
    })

    if i < len(prime.primes) && prime.primes[i].Cmp(n) == 0 {
        return true, nil
    } else {
        return false, nil
    }
}

func isPrime(n *big.Int, primes []*big.Int, expire time.Time) (bool, error) {
    m := new(big.Int)
    m.Sqrt(n)

    r := new(big.Int)

    for _, p := range primes {
        if p.Cmp(m) > 0 {
            return true, nil
        }
        if r.Mod(n, p).Cmp(big0) == 0 {
            return false, nil
        }
    }
    q := new(big.Int)
    q.Add(last(primes), big2)

    for ; q.Cmp(m) <= 0; q.Add(q, big2) {
        if time.Now().After(expire) {
            return false, errors.New("timeout")
        }

        if r.Mod(n, q).Cmp(big0) == 0 {
            return false, nil
        }
    }
    return true, nil
}

func last(s []*big.Int) *big.Int {
    if len(s) == 0 {
        return nil
    }
    return s[len(s)-1]
}

func newInt(i *big.Int) *big.Int {
    bi := new(big.Int)
    bi.Set(i)

    return bi
}
