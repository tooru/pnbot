package prime

// TODO create db viewer

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
var big1 = big.NewInt(1)
var big2 = big.NewInt(2)
var big3 = big.NewInt(3)

var bTrue  = []byte{byte(1)}
var bFalse = []byte{byte(0)}

var maxCacheNumber = big.NewInt(1000 * 1000)

type Prime struct {
    i *big.Int
    cur *big.Int
    pdb *primeDB

    mutex sync.Mutex
}

func NewPrime() *Prime {
    return newPrime(newPrimeDB())
}

func newPrime(primeDB *primeDB) *Prime {
    return &Prime {
        i: big.NewInt(0),
        cur: nil,
        pdb: primeDB,
    }
}

func (prime *Prime) IsPrime(n *big.Int) (bool, error) {
    prime.mutex.Lock()
    defer prime.mutex.Unlock()

    return prime.isPrime(n, time.Now().Add(timeout))
}

func (prime *Prime) isPrime(n *big.Int, expire time.Time) (bool, error) {
    if n.Cmp(big2) < 0 {
        return false, nil
    }
    if time.Now().After(expire) {
        return false, errors.New("timeout")
    }

    isPrime, err := prime.pdb.isPrime(n)
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

    pdb := prime.pdb
    subPrime := prime.NewPrime()

    for {
        p, err := subPrime.next(expire)
        if err != nil {
            return false, err
        }

        if p.Cmp(r) > 0 {
            //fmt.Printf("IsPrime: p> r: n=%v/r=%v/p=%v\n", n, r, p)
            break
        }
        m := new(big.Int)
        if m.Mod(n, p).Cmp(big0) == 0 {
            return false, nil
        }
    }
    pdb.addPrime(n)
    return true, nil
}


func (prime *Prime) NewPrime() *Prime {
    return newPrime(prime.pdb)
}

func (prime *Prime) Next() (next *big.Int, err error) {
    prime.mutex.Lock()
    defer prime.mutex.Unlock()

    
    return prime.next(time.Now().Add(timeout))
}

func (prime *Prime) next(expire time.Time) (next *big.Int, err error) {
    //fmt.Printf("Next:%v\n", ps.i)

    if time.Now().After(expire) {
        return nil, errors.New("timeout")
    }
    if prime.cur == nil {
        prime.cur = big.NewInt(2)
        prime.i.Add(prime.i, big1)
        return big.NewInt(2), nil
    } else if prime.cur.Cmp(big2) == 0 {
        prime.cur = big.NewInt(3)
        prime.i.Add(prime.i, big1)
        return big.NewInt(3), nil
    }

    if err != nil {
        panic("") //return nil, err
    }

    pdb := prime.pdb
    p, err := pdb.nPrime(prime.i)

    if p != nil {
        prime.cur.Set(p)
        prime.i.Add(prime.i, big1)
        return newInt(p), nil
    }

    p = newInt(prime.cur)
    for {
        p.Add(p, big2)
        //log.Printf("next?: %v\n", p)

        b, err := prime.isPrime(p, expire)
        if err != nil {
            return nil, err
        }
        if b {
            break
        }
    }

    //log.Printf("nextPrime:%v", p)
    
    if p.Cmp(maxCacheNumber) < 0 {
        if err := pdb.addNextPrime(p); err != nil {
            panic("") //return nil, err
        }
    }
    prime.cur = p
    prime.i.Add(prime.i, big1)
  
    return newInt(p), nil
}

type primeDB struct {
    mutex sync.Mutex

    primes []*big.Int
}

func newPrimeDB() *primeDB {
    return &primeDB{
        primes: []*big.Int{
            big.NewInt(2),
            big.NewInt(3),
        },
    }
}

func (pdb *primeDB) isPrime(n *big.Int) (result string, err error) {
    pdb.mutex.Lock()
    defer pdb.mutex.Unlock()

    i := sort.Search(len(pdb.primes), func(i int) bool {
        return pdb.primes[i].Cmp(n) >= 0
    })

    if i == len(pdb.primes) {
        return "unknown", nil
    } else if pdb.primes[i].Cmp(n) == 0 {
        return "prime", nil
    } else {
        return "noPrime", nil
    }
}

func (pdb *primeDB) addNextPrime(prime *big.Int) error {
    pdb.mutex.Lock()
    defer pdb.mutex.Unlock()

    assertNextPrime(pdb, prime)

    pdb.primes = append(pdb.primes, prime)
    printSlice(pdb.primes)

    return nil
}

func assertNextPrime(pdb *primeDB, next *big.Int) {
    prev := last(pdb.primes)

    if prev.Cmp(next) >= 0 {
        log.Printf("prev: %v next: %v", prev, next)
        panic("")
        //log.Fatalf("prev: %v next: %v", prev, next)
    }

    j := new(big.Int)
    j.Add(prev, big1)
    for ; j.Cmp(next) < 0; j.Add(j, big1) {
        //log.Printf(" skip: %v\n", j)
    }

}

func (pdb *primeDB) addPrime(prime *big.Int) error {
    pdb.mutex.Lock()
    defer pdb.mutex.Unlock()

    if last(pdb.primes).Cmp(prime) >= 0 {
        return nil
    }

    return nil
}

func (pdb *primeDB) nPrime(i *big.Int) (prime *big.Int, err error) {
    pdb.mutex.Lock()
    defer pdb.mutex.Unlock()

    //defer log.Printf("nPrime %v %v %v\n", i, prime, len(pdb.primes))

    length := new(big.Int)
    length.SetUint64(uint64(len(pdb.primes)))
    if length.Cmp(i) <= 0 {
        return nil, nil
    }

    return newInt(pdb.primes[i.Int64()]), nil
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

func printSlice(bs []*big.Int) {
//    sep := ""
//    log.Printf("slice: [")
//    for _, b := range bs {
//        log.Printf("%v", b)
//        log.Printf("%s", sep)
//        sep = ","
//    }
//    log.Printf("]\n")
}
