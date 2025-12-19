package AccountManager

import "github.com/ethereum/go-ethereum/crypto"

func (m *AccountManager) AddAccount(name string, hexKey string) error {
	pk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return err
	}

	addr := crypto.PubkeyToAddress(pk.PublicKey)

	m.accounts[name] = &Account{
		Name:       name,
		PrivateKey: pk,
		Address:    addr,
	}
	return nil
}
