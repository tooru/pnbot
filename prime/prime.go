package prime

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/boltdb/bolt"
)

var big0 = big.NewInt(0)
var big2 = big.NewInt(2)
var big3 = big.NewInt(3)

var primes = []*big.Int{big2, big3}
var pMutex = new(sync.Mutex)

// db format
//  "nSuccession": bytes of big.Int
//  "succession": Bucket
//    big.Int: 1
//  "noSuccession": Bucket
//    big.Int: 0,1
var db bolt.DB

type Primes struct {
    i int // TODO more scalable
}

// error will be IO or timeout
func IsPrime(n *big.Int) (bool, error) {
    if n.Cmp(big2) < 0 {
        return false, nil
    }

    r := new(big.Int)
    r.Sqrt(n)

    fmt.Printf("IsPrime: %v/%v/%v\n", n, r, primes)

    ps := new(Primes)

    for {
        p, err := ps.Next()
        if err != nil {
            return false, err
        }
        fmt.Printf("IsPrime: %v/%v/%v/%v\n", n, r, p, primes)

        if p.Cmp(r) > 0 {
            break
        }
        m := new(big.Int)
        if m.Mod(n, p).Cmp(big0) == 0 {
            return false, nil
        }
    }
    return true, nil
}

func (ps *Primes) Next() (*big.Int, error) {
    //pMutex.Lock() // TODO optimize
    //defer pMutex.Unlock()

    fmt.Printf("Next:%v/%v\n", ps.i, primes)

    if ps.i < len(primes) {
        i := ps.i
        ps.i++
        return newInt(primes[i]), nil
    }

    p := newInt(primes[len(primes)-1])

    for {
        p.Add(p, big2)

        b, err := IsPrime(p)
        if err != nil {
            return nil, err
        }
        if b {
            primes = append(primes, newInt(p))

            ps.i++
            return newInt(p), nil
        }
    }
}

func newInt(i *big.Int) *big.Int {
    bi := new(big.Int)
    bi.Set(i)

    return bi
}
