package key

import (
	"fmt"
	"testing"

	ristretto "github.com/bwesterb/go-ristretto"
	"github.com/stretchr/testify/assert"
)

func TestDidReceive(t *testing.T) {

	k := NewKeyPair([]byte("this is the seed"))

	var r ristretto.Scalar
	r.Rand()

	var R ristretto.Point
	R.ScalarMultBase(&r)

	pubKey0 := k.PublicKey().StealthAddress(r, 0)
	pubKey1 := k.PublicKey().StealthAddress(r, 2)

	privKey0, ok := k.DidReceiveTx(R, *pubKey0, 0)
	assert.True(t, ok)
	privKey1, ok := k.DidReceiveTx(R, *pubKey1, 2)
	assert.True(t, ok)

	var expectedPubKey0 ristretto.Point
	expectedPubKey0.ScalarMultBase(privKey0)
	var expectedPubKey1 ristretto.Point
	expectedPubKey1.ScalarMultBase(privKey1)

	assert.True(t, expectedPubKey0.Equals(&pubKey0.P))
	assert.True(t, expectedPubKey1.Equals(&pubKey1.P))

	fmt.Println(expectedPubKey0, pubKey0.P)
	fmt.Println(expectedPubKey1, pubKey1.P)
}
