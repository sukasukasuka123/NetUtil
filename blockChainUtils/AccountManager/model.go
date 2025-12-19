package AccountManager

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Name       string
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

type AccountManager struct {
	chainID  *big.Int
	gasLimit uint64
	accounts map[string]*Account
}

func NewAccountManager(chainID *big.Int) *AccountManager {
	return &AccountManager{
		chainID:  chainID,
		gasLimit: 3_000_000,
		accounts: make(map[string]*Account),
	}
}
