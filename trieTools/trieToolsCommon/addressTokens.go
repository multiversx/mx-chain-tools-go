package trieToolsCommon

type addressTokensMap struct {
	internalMap map[string]map[string]struct{}
}

func NewAddressTokensMap() AddressTokensMap {
	return &addressTokensMap{
		internalMap: make(map[string]map[string]struct{}),
	}
}

func (atm *addressTokensMap) Add(addr string, tokens map[string]struct{}) {
	_, addressExists := atm.internalMap[addr]
	if !addressExists {
		atm.internalMap[addr] = tokens
	} else {
		atm.addTokens(addr, tokens)
	}
}

func (atm *addressTokensMap) addTokens(addr string, tokens map[string]struct{}) {
	for token := range tokens {
		atm.internalMap[addr][token] = struct{}{}
	}
}

func (atm *addressTokensMap) HasAddress(addr string) bool {
	_, found := atm.internalMap[addr]
	return found
}

func (atm *addressTokensMap) HasToken(addr string, token string) bool {
	_, found := atm.internalMap[addr][token]
	return found
}

func (atm *addressTokensMap) NumAddresses() uint64 {
	return uint64(len(atm.internalMap))
}

func (atm *addressTokensMap) NumTokens() uint64 {
	numTokens := uint64(0)
	for _, tokens := range atm.internalMap {
		for range tokens {
			numTokens++
		}
	}

	return numTokens
}

func (atm *addressTokensMap) GetMapCopy() map[string]map[string]struct{} {
	addressTokensMapCopy := make(map[string]map[string]struct{})

	for address, tokens := range atm.internalMap {
		addressTokensMapCopy[address] = make(map[string]struct{})
		for token := range tokens {
			addressTokensMapCopy[address][token] = struct{}{}
		}
	}

	return addressTokensMapCopy
}

func (atm *addressTokensMap) GetAddresses() map[string]struct{} {
	ret := make(map[string]struct{})
	for addr := range atm.internalMap {
		ret[addr] = struct{}{}
	}

	return ret
}

func (atm *addressTokensMap) GetTokens(address string) map[string]struct{} {
	return atm.internalMap[address]
}

func (atm *addressTokensMap) Delete(address string) {
	delete(atm.internalMap, address)
}

func (atm *addressTokensMap) GetAllTokens() map[string]struct{} {
	allTokens := make(map[string]struct{})
	for _, tokens := range atm.internalMap {
		for token := range tokens {
			allTokens[token] = struct{}{}
		}
	}

	return allTokens
}

func (atm *addressTokensMap) ShallowClone() AddressTokensMap {
	mapCopy := atm.GetMapCopy()
	return &addressTokensMap{
		internalMap: mapCopy,
	}
}
