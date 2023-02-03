package idgen

import (
	"math/big"
)

// IID Integer ID
type IID uint64

func (z IID) IsZero() bool {
	return z == 0
}

// Bytes ...
func (z IID) Bytes() []byte {
	var bInt big.Int
	return bInt.SetUint64(uint64(z)).Bytes()
}

// String ...
func (z IID) String() string {
	var bInt big.Int
	return bInt.SetUint64(uint64(z)).Text(36)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (z IID) MarshalBinary() ([]byte, error) {
	return z.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (z *IID) UnmarshalBinary(data []byte) (err error) {
	var bI big.Int
	id := bI.SetBytes(data).Uint64()
	*z = IID(id)
	return
}

// MarshalText implements the encoding.TextMarshaler interface.
func (z IID) MarshalText() ([]byte, error) {
	b := []byte(z.String())
	return b, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (z *IID) UnmarshalText(data []byte) (err error) {
	if id, ok := ParseID(string(data)); ok {
		*z = id
	}

	return
}

// ParseID ...
func ParseID(s string) (IID, bool) {
	var bI big.Int
	if i, ok := bI.SetString(s, 36); ok {
		return IID(i.Uint64()), true
	}
	return 0, false
}

// nolint
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
