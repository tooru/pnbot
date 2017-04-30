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

var bTrue  = []byte{byte(1)}
var bFalse = []byte{byte(0)}

var pdb = new(primeDB)

var pMutex = new(sync.Mutex) // TODO remove?

// error will be IO or timeout
func IsPrime(n *big.Int) (bool, error) {
    if n.Cmp(big2) < 0 {
        return false, nil
    }

    isPrime, err := pdb.isPrime(n)
    if err != nil {
        return false, err
    }

    switch isPrime {
    case "prime":
        return true, nil
    case "notPrime":
        return false, nil
    }


    r := new(big.Int)
    r.Sqrt(n)

    //fmt.Printf("IsPrime: n=%v/r=%v\n", n, r)

    ps := NewPrime()

    for {
        p, err := ps.Next()
        if err != nil {
            panic("") //return false, err
        }

        if p.Cmp(r) > 0 {
            //fmt.Printf("IsPrime: p> r: n=%v/r=%v/p=%v\n", n, r, p)
            break
        }
        m := new(big.Int)
        if m.Mod(n, p).Cmp(big0) == 0 {
            //fmt.Printf("IsPrime: n %% p == 0: n=%v/r=%v/p=%v\n", n, r, p)
            pdb.addNotPrime(n)
            return false, nil
        }
        //fmt.Printf("IsPrime: n %% p != 0: n=%v/r=%v/p=%v\n", n, r, p)
    }
    pdb.addPrime(n)
    return true, nil
}

type Prime struct {
    i *big.Int
    cur *big.Int
}

func NewPrime () *Prime {
    ps := new(Prime)
    ps.i = big.NewInt(0)
    ps.cur = nil
    return ps
}

func (ps *Prime) Next() (next *big.Int, err error) {
    //pMutex.Lock() // TODO optimize
    //defer pMutex.Unlock()

    //fmt.Printf("Next:%v\n", ps.i)

    if ps.cur == nil {
        ps.cur = big.NewInt(2)
        ps.i.Add(ps.i, big1)
        return big.NewInt(2), nil
    } else if ps.cur.Cmp(big2) == 0 {
        ps.cur = big.NewInt(3)
        ps.i.Add(ps.i, big1)
        return big.NewInt(3), nil
    }

    if err != nil {
        panic("") //return nil, err
    }

    nSuccession, err := pdb.nSuccession()
    if err != nil {
        panic("") //return nil, err
    }
    
    if ps.i.Cmp(nSuccession) < 0 {
        next, err = pdb.nextPrime(ps.cur)
        if err != nil {
            return nil, err
        }
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

    //fmt.Printf("Put:%v -> %v\n", ps.cur, p)
    
    if err := pdb.addPrime(p); err != nil {
        panic("") //return nil, err
    }
    ps.cur = p
    ps.i.Add(ps.i, big1)
  
    return p, nil
}

type primeDB struct {
}

func (pdb *primeDB) nSuccession() (nSucc *big.Int, err error) {
    err = updatePDB(func(tx *bolt.Tx) error {
        prop := tx.Bucket([]byte("prop"))

        nSucc = new(big.Int)
        nSucc.SetBytes(prop.Get([]byte("nSuccession")))

        return nil
    })
    if err != nil {
        return nil, err
    }
    return nSucc, nil
}

func (pdb *primeDB) isPrime(n *big.Int) (result string, err error) {
    err = updatePDB(func(tx *bolt.Tx) error {
        primes := tx.Bucket([]byte("primes"))
        next := primes.Get(n.Bytes())
        if next != nil{
            result = "prime"
            return nil
        }
        
        noSuccession := tx.Bucket([]byte("noSuccession"))
        num := noSuccession.Get(n.Bytes())

        if num != nil {
            if num[0] == 0 {
                result = "noPrime"
            } else {
                result = "prime"
            }
        } else {
            result = "unknown"
        }
        return nil
    })
    if err != nil {
        return "", err
    }
    return result, nil
}

func (pdb *primeDB) addPrime(prime *big.Int) error {
    err := updatePDB(func(tx *bolt.Tx) error {
        prop := tx.Bucket([]byte("prop"))

        maxSuccPrime := new(big.Int)
        maxSuccPrime.SetBytes(prop.Get([]byte("maxSuccPrime")))

        if prime.Cmp(maxSuccPrime) <= 0 {
            return nil
        }
        
        fmt.Printf("addPrime: %v\n", prime)

        maxSuccNoPrime := new(big.Int)
        maxSuccNoPrime.SetBytes(prop.Get([]byte("maxSuccNoPrime")))

        primeBytes := prime.Bytes()

        prevPrime := new(big.Int)
        prevPrime.Sub(prime, big1)

        noSuccession := tx.Bucket([]byte("noSuccession"))
        if noSuccession == nil {
            panic("noSuccession is nil")
        }

        if prevPrime.Cmp(maxSuccNoPrime) > 0 {
            noSuccession.Put(primeBytes, bTrue)
            return nil
        }

        nSuccession := new(big.Int)
        nSuccession.SetBytes(prop.Get([]byte("nSuccession")))
        nSuccession.Add(nSuccession, big1)

        primes := tx.Bucket([]byte("primes"))
        primes.Put(maxSuccPrime.Bytes(), primeBytes)

        maxSuccPrime.Set(prime)
        maxSuccNoPrime.Add(prime, big1)

        maxSucc := new(big.Int)
        maxSucc.Add(maxSuccNoPrime, big1)

        for {
            isPrime := noSuccession.Get(maxSucc.Bytes())
            if (isPrime == nil) {
                break;
            }

            err := noSuccession.Delete(maxSucc.Bytes())
            if err != nil {
                return err
            }

            if (isBytesTrue(isPrime)) {
                primes.Put(maxSuccPrime.Bytes(), maxSucc.Bytes())

                maxSuccPrime.Set(maxSucc)
                maxSuccNoPrime.Add(maxSucc, big1)
                nSuccession.Add(nSuccession, big1)
            } else {
                maxSuccNoPrime.Set(maxSucc)
            }
        }

        prop.Put([]byte("nSuccession"), nSuccession.Bytes())
        prop.Put([]byte("maxSuccPrime"), primeBytes)
        prop.Put([]byte("maxSuccNoPrime"), maxSuccNoPrime.Bytes())
        
        return nil
    })
    if err != nil {
        return err
    }
    return nil
}

func (pdb *primeDB) addNotPrime(noPrime *big.Int) error {
    err := updatePDB(func(tx *bolt.Tx) error {
        prop := tx.Bucket([]byte("prop"))

        maxSuccNoPrime := new(big.Int)
        maxSuccNoPrime.SetBytes(prop.Get([]byte("maxSuccNoPrime")))

        if noPrime.Cmp(maxSuccNoPrime) <= 0 {
            return nil
        }
        fmt.Printf("addNotPrime: %v %v\n", noPrime, maxSuccNoPrime)
        
        noPrimeBytes := noPrime.Bytes()

        maxSuccNoPrime.Add(maxSuccNoPrime, big1)

        noSuccession := tx.Bucket([]byte("noSuccession"))
        if noPrime.Cmp(maxSuccNoPrime) != 0 {
            noSuccession.Put(noPrimeBytes, bFalse)
            return nil
        }

        maxSuccPrime := new(big.Int)
        maxSuccPrime.SetBytes(prop.Get([]byte("maxSuccPrime")))
        
        primes := tx.Bucket([]byte("primes"))

        next := new(big.Int)
        next.Add(noPrime, big1)

        for {
            nextBytes := next.Bytes()
            isPrime := noSuccession.Get(nextBytes)
            if len(isPrime) == 0 {
                if noPrime.Bit(0) == 0 {
                    prop.Put([]byte("maxSuccNoPrime"), noPrimeBytes)
                } else {
                    prop.Put([]byte("maxSuccNoPrime"), nextBytes)
                }
                return nil
            }

            err := noSuccession.Delete(nextBytes)
            if err != nil {
                return err
            }
            if isPrime[0] == 0 {
                next.Add(next, big1)
                prop.Put([]byte("maxSuccNoPrime"), nextBytes)
                continue
            }

            primes.Put(maxSuccPrime.Bytes(), nextBytes)
            maxSuccPrime.Set(next)
            prop.Put([]byte("maxSuccPrime"), nextBytes)
        }
    })
    if err != nil {
        return err
    }
    return nil
}

func (pdb *primeDB) nextPrime(prime *big.Int) (next *big.Int, err error) {
    next = new(big.Int)
    err = updatePDB(func(tx *bolt.Tx) error {
        primes := tx.Bucket([]byte("primes"))

        next.SetBytes(primes.Get(prime.Bytes()))

        return nil
    })
    if err != nil {
        return nil, err
    }
    return next, nil
}

// db format
//  "prop"
//    "nSuccession": big.Int
//    "maxSuccPrime": big.Int
//    "maxSuccNoPrime": big.Int
//  "prime": Bucket
//    big.Int: big.Int (next prime)
//  "noSuccession": Bucket
//    big.Int: 0,1

func updatePDB(fn func(*bolt.Tx) error) error {
    db, err := bolt.Open("prime.db", 0600, nil)
    if err != nil {
        return err
    }
    defer func() {
        db.Close()
        db = nil
    }()

    return db.Update(func(tx *bolt.Tx) error {
        prop := tx.Bucket([]byte("prop"))
        if prop == nil {
            prop, err := tx.CreateBucket([]byte("prop"))
            if err != nil {
                panic("") //return err
            }
            prop.Put([]byte("nSuccession"), big2.Bytes())
            prop.Put([]byte("maxSuccPrime"), big3.Bytes())
            prop.Put([]byte("maxSuccNoPrime"), big.NewInt(4).Bytes())

            primes, err := tx.CreateBucket([]byte("primes"))
            if err != nil {
                panic("") //return err
            }
            err = primes.Put(big2.Bytes(), big3.Bytes());
            if err != nil {
                panic("") //return err
            }

            _, err = tx.CreateBucket([]byte("noSuccession"))
            if err != nil {
                panic("") //return err
            }
        }
        return fn(tx)
    })
}

func initDB(tx *bolt.Tx, prop *bolt.Bucket) error {
    prop.Put([]byte("nSuccession"), big2.Bytes())
    primes, err := tx.CreateBucket([]byte("primes"))
    
    if err != nil {
        panic("") //return err
    }
    return primes.Put(big2.Bytes(), big3.Bytes());
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

func isBytesTrue(bytes []byte) bool {
    if len(bytes) != len(bTrue) {
        return false
    }

    return bytes[0] == bTrue[0]
}
