package AccountManager

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func (m *AccountManager) Auth(name string) (*bind.TransactOpts, error) {
	acc, ok := m.accounts[name]
	if !ok {
		return nil, fmt.Errorf("account not found: %s", name)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(acc.PrivateKey, m.chainID)
	if err != nil {
		return nil, err
	}

	auth.GasLimit = m.gasLimit
	return auth, nil
}
