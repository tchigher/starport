package starportsecretconf

import (
	"io"
	"os"
	"path/filepath"

	"github.com/cosmos/go-bip39"
	"gopkg.in/yaml.v2"
)

const (
	SecretFile             = ".secret.yml"
	SelfRelayerAccountName = "relayer"
)

var (
	selfRelayerAccountDefaultCoins = []string{"800token"}
)

type Config struct {
	Accounts []Account `yaml:"accounts"`
	Relayer  Relayer   `yaml:"relayer"`
}

func (c *Config) SelfRelayerAccount() (account Account, found bool) {
	for _, a := range c.Accounts {
		if a.Name == SelfRelayerAccountName {
			return a, true
		}
	}
	return Account{}, false
}

func (c *Config) SetSelfRelayerAccount() error {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return err
	}
	c.Accounts = append(c.Accounts, Account{
		Name:     SelfRelayerAccountName,
		Coins:    selfRelayerAccountDefaultCoins,
		Mnemonic: mnemonic,
	})
	return nil
}

func (c *Config) UpsertRelayerAccount(acc RelayerAccount) {
	var found bool
	for i, account := range c.Relayer.Accounts {
		if account.ID == acc.ID {
			found = true
			c.Relayer.Accounts[i] = acc
			break
		}
	}
	if !found {
		c.Relayer.Accounts = append(c.Relayer.Accounts, acc)
	}
}

// Account holds the options related to setting up Cosmos wallets.
type Account struct {
	Name     string   `yaml:"name"`
	Coins    []string `yaml:"coins"`
	Mnemonic string   `yaml:"mnemonic"`
}

type RelayerAccount struct {
	ID         string `yaml:"id"`
	Mnemonic   string `yaml:"mnemonic"`
	RPCAddress string `yaml:"rpc_address"`
}

type Relayer struct {
	Accounts []RelayerAccount `yaml:"accounts"`
}

// Parse parses config.yml into Config.
func Parse(r io.Reader) (*Config, error) {
	var conf Config
	return &conf, yaml.NewDecoder(r).Decode(&conf)
}

func Open(path string) (*Config, error) {
	file, err := os.Open(filepath.Join(path, SecretFile))
	if err != nil {
		return &Config{}, nil
	}
	defer file.Close()
	return Parse(file)
}

func Save(path string, conf *Config) error {
	file, err := os.OpenFile(filepath.Join(path, SecretFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	return yaml.NewEncoder(file).Encode(conf)
}
