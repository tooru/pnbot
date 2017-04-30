package main

import (
    "fmt"
    "math/big"

    "github.com/boltdb/bolt"
)

func main() {
    db, err := bolt.Open("prime/prime.db", 0600, nil)
    if err != nil {
        fmt.Printf("err=%v\n", err)
        return
    }
    defer db.Close()

    db.View(func(tx *bolt.Tx) error {
        fmt.Printf("path: %v\n", db.Path())

        prop := tx.Bucket([]byte("prop"))

        if prop == nil {
            panic("prop is nil")
        }

        fmt.Printf("prop\n")

        maxSuccPrime := new(big.Int)
        maxSuccPrime.SetBytes(prop.Get([]byte("maxSuccPrime")))
        fmt.Printf(" maxSuccPrime: %v\n", maxSuccPrime)

        maxSuccNoPrime := new(big.Int)
        maxSuccNoPrime.SetBytes(prop.Get([]byte("maxSuccNoPrime")))
        fmt.Printf(" maxSuccNoPrime: %v\n", maxSuccNoPrime)

        nSuccession := new(big.Int)
        nSuccession.SetBytes(prop.Get([]byte("nSuccession")))
        fmt.Printf(" nSuccession: %v\n", nSuccession)

        fmt.Printf("primes\n")
        primes := tx.Bucket([]byte("primes"))

        if primes == nil {
            panic("primes is nil")
        }

        prime := big.NewInt(2)
        for {
            fmt.Printf(" %v\n", prime)

            next := primes.Get(prime.Bytes())

            if len(next) == 0 {
                break
            }
            prime.SetBytes(next)
        }

        fmt.Printf("noSuccession\n")
        noSuccs := tx.Bucket([]byte("noSuccession"))

        noSucc := new(big.Int)
        cursor := noSuccs.Cursor()

        for k, v := cursor.First(); k != nil && v != nil; k, v = cursor.Next() {
            noSucc.SetBytes(k)
            isPrime := len(v) == 1 && v[0] == 1
            
            fmt.Printf(" %v: %v\n", noSucc, isPrime)
        }

        return nil
    })
}
