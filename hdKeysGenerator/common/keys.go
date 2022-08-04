package common

// GeneratedKey holds the secret key and public key, along with the bech32 address and other metadata.
type GeneratedKey struct {
	AccountIndex int
	AddressIndex int
	SecretKey    []byte
	PublicKey    []byte
	Address      string
}
