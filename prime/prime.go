package prime

// TODO create db viewer

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/boltdb/bolt"
)

var big0 = big.NewInt(0)
var big1 = big.NewInt(1)
var big2 = big.NewInt(2)
var big3 = big.NewInt(3)

//var primes = []*big.Int{big2, big3}
var pMutex = new(sync.Mutex)

// db format
//  "prop"
//    "nSuccession": bytes of big.Int
//  "succession": Bucket
//    big.Int: 1
//  "noSuccession": Bucket
//    big.Int: 0,1
var db *bolt.DB

type Primes struct {
    i *big.Int
    cur *big.Int
}

func NewPrimes () *Primes {
    ps := new(Primes)
    ps.i = big.NewInt(0)
    ps.cur = nil
    return ps
}

// error will be IO or timeout
func IsPrime(n *big.Int) (bool, error) {
    if n.Cmp(big2) < 0 {
        return false, nil
    }

    r := new(big.Int)
    r.Sqrt(n)

    fmt.Printf("IsPrime: n=%v/r=%v\n", n, r)

    ps := NewPrimes()

    for {
        p, err := ps.Next()
        if err != nil {
            panic("") //return false, err
        }

        if p.Cmp(r) > 0 {
            fmt.Printf("IsPrime: p> r: n=%v/r=%v/p=%v\n", n, r, p)
            break
        }
        m := new(big.Int)
        if m.Mod(n, p).Cmp(big0) == 0 {
            fmt.Printf("IsPrime: n %% p == 0: n=%v/r=%v/p=%v\n", n, r, p)
            return false, nil
        }
        fmt.Printf("IsPrime: n %% p != 0: n=%v/r=%v/p=%v\n", n, r, p)
    }
    // TODO update DB
    return true, nil
}

func (ps *Primes) Next() (next *big.Int, err error) {
    //pMutex.Lock() // TODO optimize
    //defer pMutex.Unlock()

    fmt.Printf("Next:%v\n", ps.i)

    if ps.cur == nil {
        ps.cur = big.NewInt(2)
        ps.i.Add(ps.i, big1)
        return big.NewInt(2), nil
    } else if ps.cur.Cmp(big2) == 0 {
        ps.cur = big.NewInt(3)
        ps.i.Add(ps.i, big1)
        return big.NewInt(3), nil
    }

    if db == nil {
        db, err = bolt.Open("prime.db", 0600, nil)
        defer func() {
            db.Close()
            db = nil
        }()
    }
    if err != nil {
        panic("") //return nil, err
    }

    nSuccession := big.NewInt(2)
    err = db.Update(func(tx *bolt.Tx) error {
        prop := tx.Bucket([]byte("prop"))
        if prop == nil {
            prop, err = tx.CreateBucket([]byte("prop"))
            if err != nil {
                panic("") //return err
            }
            prop.Put([]byte("nSuccession"), big2.Bytes())
        } else {
            nSuccession = parseInt(prop.Get([]byte("nSuccession")))
        }

        if ps.i.Cmp(nSuccession) < 0 {
            primes := tx.Bucket([]byte("primes"))

            next = parseInt(primes.Get(ps.cur.Bytes()))
        }
        return nil
    })
    if err != nil {
        panic("") //return nil, err
    }
    if next != nil {
        ps.i.Add(ps.i, big1)
        return next, nil
    }

    p := newInt(ps.cur)
    for {
        p.Add(p, big2)

        b, err := IsPrime(p)
        if err != nil {
            panic("") //return nil, err
        }
        if b {
            break
        }
    }
    err = db.Update(func(tx *bolt.Tx) error {
        primes := tx.Bucket([]byte("primes"))
        if primes == nil {
            primes, err = tx.CreateBucket([]byte("primes"))
            if err != nil {
                panic("") //return err
            }
        }

        prop := tx.Bucket([]byte("prop"))
        nSuccession.Add(nSuccession, big1)
        err := prop.Put([]byte("nSuccession"), nSuccession.Bytes())
        if err != nil {
            panic("") //return err
        }

        fmt.Printf("Put:%v -> %v\n", ps.cur, p)
        return primes.Put(ps.cur.Bytes(), p.Bytes())
    })
    if err != nil {
        panic("") //return nil, err
    }
    ps.cur = p
    ps.i.Add(ps.i, big1)

    return p, nil
}

func newInt(i *big.Int) *big.Int {
    bi := new(big.Int)
    bi.Set(i)

    return bi
}

func parseInt(bytes []byte) *big.Int {
    bi := new(big.Int)
    if bytes != nil {
        bi.SetBytes(bytes)
    }

    return bi
}
