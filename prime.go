package pnbot

import (
    "math/big"
)

var big0 = big.NewInt(0)
var big2 = big.NewInt(2)
var primes = []*big.Int{big2}

func IsPrime(n *big.Int) bool {
    if (n.Cmp(big2) < 0) {
        return false
    }

    r := new(big.Int)
    r.Sqrt(n)

    for p := primes[len(primes)-1]; p.Cmp(r) <= 0; p.Add(p, big2) {
        if IsPrime(p) {
            primes = append(primes, p)
        }
    }

    for _, p := range primes {
        if p.Cmp(r) > 0 {
            break
        }
        m := new(big.Int)
        if m.Mod(n, p).Cmp(big0) == 0 {
            return false
        }
    }
    return true
}
