package settlement

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestCalculateFactAmt(t *testing.T) {
	fmt.Println(big.NewFloat(0).Add(big.NewFloat(12), big.NewFloat(32)).SetMode(big.ToNearestAway))
	fmt.Println(big.NewFloat(0).Mul(big.NewFloat(12), big.NewFloat(0.3123662)).SetPrec(5).SetMode(big.ToNearestAway))
	fmt.Println(big.NewFloat(0).Quo(big.NewFloat(12), big.NewFloat(2)).SetMode(big.ToNearestAway))

	factAmt, fareAmt := calculateFactAmt(1200, 5)
	fmt.Println("factAmt: ", factAmt)
	fmt.Println("fareAmt: ", fareAmt)
	assert.Equal(t, fareAmt, uint32(60))
	assert.Equal(t, factAmt, uint32(1140))

	factAmt1, fareAmt1 := calculateFactAmt(1, 5)
	fmt.Println("factAmt1: ", factAmt1)
	fmt.Println("fareAmt1: ", fareAmt1)
	assert.Equal(t, fareAmt1, uint32(0))
	assert.Equal(t, factAmt1, uint32(1))

	factAmt10, fareAmt10 := calculateFactAmt(10, 0.5)
	fmt.Println("fareAmt10: ", fareAmt10)
	fmt.Println("factAmt10: ", factAmt10)
	assert.Equal(t, fareAmt10, uint32(0))
	assert.Equal(t, factAmt10, uint32(10))

	factAmt1231531234, fareAmt1231531234 := calculateFactAmt(1231531234, 5)
	fmt.Println("factAmt1231531234: ", factAmt1231531234)
	fmt.Println("fareAmt1231531234: ", fareAmt1231531234)
	assert.Equal(t, fareAmt1231531234, uint32(61576562))
	assert.Equal(t, factAmt1231531234, uint32(1169954672))

	factAmt123153123405, fareAmt123153123405 := calculateFactAmt(1231531234, 0.5)
	fmt.Println("factAmt123153123405: ", factAmt123153123405)
	fmt.Println("fareAmt123153123405: ", fareAmt123153123405)
	assert.Equal(t, fareAmt123153123405, uint32(6157656))
	assert.Equal(t, factAmt123153123405, uint32(1225373578))
}

func TestBigMath(t *testing.T) {
	f, _ := big.NewFloat(0).Mul(big.NewFloat(1234.123), big.NewFloat(109850912859.123)).Float64()
	fmt.Printf("%f", f)
}
