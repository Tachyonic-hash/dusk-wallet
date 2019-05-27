package transactions

import (
	"bytes"
	"errors"
	"io"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/encoding"
)

// Output defines an output in a transaction.
type Output struct {
	// Commitment is the pedersen commitment to the underlying amount
	// In a bidding transaction, it is the amount in cleartext
	// For this reason, the size is varied. Once bidding transactions use Commitments,
	// The size will be changed to a fixed 32 bytes
	Commitment []byte // Variable size
	// DestKey is the one-time public key of the address that
	// the funds should be sent to.
	DestKey []byte // 32 bytes
	// RangeProof is the bulletproof rangeproof that proves that the hidden amount
	// is between 0 and 2^64
	RangeProof []byte // Variable size
}

// NewOutput constructs a new Output from the passed parameters.
// This function is placed here for consistency with the rest of the API.
func NewOutput(comm []byte, dest []byte, proof []byte) (*Output, error) {

	if len(dest) != 32 {
		return nil, errors.New("destination key is not 32 bytes")
	}

	return &Output{
		Commitment: comm,
		DestKey:    dest,
		RangeProof: proof,
	}, nil
}

// Encode an Output struct and write to w.
func (o *Output) Encode(w io.Writer) error {
	if err := encoding.Write256(w, o.Commitment); err != nil {
		return err
	}

	if err := encoding.Write256(w, o.DestKey); err != nil {
		return err
	}

	if err := encoding.WriteVarBytes(w, o.RangeProof); err != nil {
		return err
	}

	return nil
}

// Decode an Output object from r into an output struct.
func (o *Output) Decode(r io.Reader) error {
	if err := encoding.Read256(r, &o.Commitment); err != nil {
		return err
	}

	if err := encoding.Read256(r, &o.DestKey); err != nil {
		return err
	}

	if err := encoding.ReadVarBytes(r, &o.RangeProof); err != nil {
		return err
	}

	return nil
}

// Equals returns true if two outputs are the same
func (o *Output) Equals(out *Output) bool {
	if o == nil || out == nil {
		return false
	}

	if !bytes.Equal(o.Commitment, out.Commitment) {
		return false
	}

	if !bytes.Equal(o.DestKey, out.DestKey) {
		return false
	}

	if !bytes.Equal(o.RangeProof, out.RangeProof) {
		return false
	}

	return true
}
