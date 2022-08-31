package trieToolsCommon

type AddressTokensMap interface {
	Add(addr string, tokens map[string]struct{})
	Delete(address string)
	GetAddresses() map[string]struct{}
	GetAllTokens() map[string]struct{}
	GetTokens(address string) map[string]struct{}
	GetMapCopy() map[string]map[string]struct{}
	ShallowClone() AddressTokensMap
	HasAddress(addr string) bool
	HasToken(addr string, token string) bool
	NumAddresses() uint64
	NumTokens() uint64
}
