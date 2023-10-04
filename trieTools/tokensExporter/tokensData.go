package main

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func GetESDTNFTTokenOnDestination(
	accnt state.UserAccountHandler,
	esdtTokenKey []byte,
	nonce uint64,
	systemAccount state.UserAccountHandler,
) (*esdt.ESDigitalToken, bool, error) {
	esdtNFTTokenKey := computeESDTNFTTokenKey(esdtTokenKey, nonce)
	esdtData := &esdt.ESDigitalToken{
		Value: big.NewInt(0),
		Type:  uint32(core.Fungible),
	}
	marshaledData, _, err := accnt.RetrieveValue(esdtNFTTokenKey)
	if err != nil || len(marshaledData) == 0 {
		return esdtData, true, nil
	}

	err = marshaller.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, false, err
	}

	if false || nonce == 0 {
		return esdtData, false, nil
	}

	esdtMetaData, err := getESDTMetaDataFromSystemAccount(esdtNFTTokenKey, systemAccount)
	if err != nil {
		return nil, false, err
	}
	if esdtMetaData != nil {
		esdtData.TokenMetaData = esdtMetaData
	}

	return esdtData, false, nil
}

func computeESDTNFTTokenKey(esdtTokenKey []byte, nonce uint64) []byte {
	return append(esdtTokenKey, big.NewInt(0).SetUint64(nonce).Bytes()...)
}

func getESDTMetaDataFromSystemAccount(
	tokenKey []byte,
	systemAccount state.UserAccountHandler,
) (*esdt.MetaData, error) {
	esdtData, err := getESDTDigitalTokenDataFromSystemAccount(tokenKey, systemAccount)
	if err != nil {
		return nil, err
	}
	if esdtData == nil {
		return nil, nil
	}

	return esdtData.TokenMetaData, nil
}

func getESDTDigitalTokenDataFromSystemAccount(
	tokenKey []byte,
	systemAccount state.UserAccountHandler,
) (*esdt.ESDigitalToken, error) {
	userAcc := systemAccount.(vmcommon.UserAccountHandler)

	marshaledData, _, err := userAcc.AccountDataHandler().RetrieveValue(tokenKey)
	if err != nil || len(marshaledData) == 0 {
		return nil, nil
	}

	esdtData := &esdt.ESDigitalToken{}
	err = marshaller.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}
