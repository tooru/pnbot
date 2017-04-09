package pnbot

import (
    "math/big"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestIsPrime(t *testing.T) {
    assert.Equal(t, false, IsPrime(big.NewInt(1)))
    assert.Equal(t, true, IsPrime(big.NewInt(2)))
    assert.Equal(t, true, IsPrime(big.NewInt(3)))
    assert.Equal(t, false, IsPrime(big.NewInt(4)))
    assert.Equal(t, true, IsPrime(big.NewInt(5)))
    assert.Equal(t, false, IsPrime(big.NewInt(6)))
    assert.Equal(t, true, IsPrime(big.NewInt(7)))
    assert.Equal(t, false, IsPrime(big.NewInt(8)))
    assert.Equal(t, false, IsPrime(big.NewInt(9)))
    assert.Equal(t, false, IsPrime(big.NewInt(10)))
    assert.Equal(t, true, IsPrime(big.NewInt(11)))
}
