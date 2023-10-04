package main

type exportedRoot struct {
	RootHash    string            `json:"rootHash"`
	NumAccounts int               `json:"numAccounts"`
	Accounts    []exportedAccount `json:"accounts"`
}

type exportedAccount struct {
	Address string         `json:"address"`
	Pubkey  string         `json:"pubkey"`
	Tokens  []accountToken `json:"tokens"`
}

type accountToken struct {
	Identifier string `json:"id"`
	Name       string `json:"name"`
	Nonce      uint64 `json:"nonce"`
	Balance    string `json:"balance"`
	Attributes []byte `json:"attributes"`
	Creator    string `json:"creator"`
}
