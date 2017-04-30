package prime

import (
    "math/big"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestIsPrime(t *testing.T) {
    isPrime(t, false, big.NewInt(1))
    isPrime(t, true,  big.NewInt(2))
    isPrime(t, true,  big.NewInt(3))
    isPrime(t, false, big.NewInt(4))
    isPrime(t, true,  big.NewInt(5))
    isPrime(t, false, big.NewInt(6))
    isPrime(t, true,  big.NewInt(7))
    isPrime(t, false, big.NewInt(8))
    isPrime(t, false, big.NewInt(9))
    isPrime(t, false, big.NewInt(10))
    isPrime(t, true,  big.NewInt(11))
    isPrime(t, false, big.NewInt(12))
    isPrime(t, true,  big.NewInt(13))
    isPrime(t, false, big.NewInt(14))
    isPrime(t, false, big.NewInt(15))
    isPrime(t, false, big.NewInt(16))
    isPrime(t, true,  big.NewInt(17))
    isPrime(t, false, big.NewInt(18))
    isPrime(t, true,  big.NewInt(19))
    isPrime(t, false, big.NewInt(20))
    isPrime(t, false, big.NewInt(21))
    isPrime(t, false, big.NewInt(22))
    isPrime(t, true,  big.NewInt(23))
    isPrime(t, false, big.NewInt(24))
    isPrime(t, false, big.NewInt(25))
    isPrime(t, false, big.NewInt(26))
    isPrime(t, false, big.NewInt(27))
    isPrime(t, false, big.NewInt(28))
    isPrime(t, true,  big.NewInt(29))
    isPrime(t, false, big.NewInt(30))
    isPrime(t, true,  big.NewInt(31))
    isPrime(t, false, big.NewInt(32))
    isPrime(t, false, big.NewInt(33))
    isPrime(t, false, big.NewInt(34))
    isPrime(t, false, big.NewInt(35))
    isPrime(t, false, big.NewInt(36))
    isPrime(t, true,  big.NewInt(37))
    isPrime(t, false, big.NewInt(38))
    isPrime(t, false, big.NewInt(39))
    isPrime(t, false, big.NewInt(40))
    isPrime(t, true,  big.NewInt(41))
    isPrime(t, false, big.NewInt(42))
    isPrime(t, true,  big.NewInt(43))
    isPrime(t, false, big.NewInt(44))
    isPrime(t, false, big.NewInt(45))
    isPrime(t, false, big.NewInt(46))
    isPrime(t, true,  big.NewInt(47))
    isPrime(t, false, big.NewInt(48))
    isPrime(t, false, big.NewInt(49))
    isPrime(t, false, big.NewInt(50))
    isPrime(t, false, big.NewInt(51))
    isPrime(t, false, big.NewInt(52))
    isPrime(t, true,  big.NewInt(53))
    isPrime(t, false, big.NewInt(54))
    isPrime(t, false, big.NewInt(55))
    isPrime(t, false, big.NewInt(56))
    isPrime(t, false, big.NewInt(57))
    isPrime(t, false, big.NewInt(58))
    isPrime(t, true,  big.NewInt(59))
    isPrime(t, false, big.NewInt(60))
    isPrime(t, true,  big.NewInt(61))
    isPrime(t, false, big.NewInt(62))
    isPrime(t, false, big.NewInt(63))
    isPrime(t, false, big.NewInt(64))
    isPrime(t, false, big.NewInt(65))
    isPrime(t, false, big.NewInt(66))
    isPrime(t, true,  big.NewInt(67))
    isPrime(t, false, big.NewInt(68))
    isPrime(t, false, big.NewInt(69))
    isPrime(t, false, big.NewInt(70))
    isPrime(t, true,  big.NewInt(71))
    isPrime(t, false, big.NewInt(72))
    isPrime(t, true,  big.NewInt(73))
    isPrime(t, false, big.NewInt(74))
    isPrime(t, false, big.NewInt(75))
    isPrime(t, false, big.NewInt(76))
    isPrime(t, false, big.NewInt(77))
    isPrime(t, false, big.NewInt(78))
    isPrime(t, true,  big.NewInt(79))
    isPrime(t, false, big.NewInt(80))
    isPrime(t, false, big.NewInt(81))
    isPrime(t, false, big.NewInt(82))
    isPrime(t, true,  big.NewInt(83))
    isPrime(t, false, big.NewInt(84))
    isPrime(t, false, big.NewInt(85))
    isPrime(t, false, big.NewInt(86))
    isPrime(t, false, big.NewInt(87))
    isPrime(t, false, big.NewInt(88))
    isPrime(t, true,  big.NewInt(89))
    isPrime(t, false, big.NewInt(90))
    isPrime(t, false, big.NewInt(91))
    isPrime(t, false, big.NewInt(92))
    isPrime(t, false, big.NewInt(93))
    isPrime(t, false, big.NewInt(94))
    isPrime(t, false, big.NewInt(95))
    isPrime(t, false, big.NewInt(96))
    isPrime(t, true,  big.NewInt(97))
    isPrime(t, false, big.NewInt(98))
    isPrime(t, false, big.NewInt(99))
    isPrime(t, false, big.NewInt(101*101))

    isPrime(t, true, big.NewInt(1098481))
    isPrime(t, false, big.NewInt(1134211291487)) // 1098481 * 1032527
}

func isPrime(t *testing.T, expect bool, n *big.Int) {
    b, err := IsPrime(n)

    if _, ok := err.(error); ok {
        assert.Fail(t, "error: isPrime: %v", err)
    }

    assert.Equal(t, expect, b)
}
