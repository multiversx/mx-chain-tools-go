package main

type generatedKey struct {
	Index     int
	SecretKey string `json:"secretKey"`
	PublicKey string `json:"publicKey"`
	Address   string `json:"address"`
}
