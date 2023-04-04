package vault_helper

import (
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/hashicorp/vault/api"
)

type Vault struct {
	Address string
	Token   string
}

type KVSecret struct {
	Key   string
	Value string
}

type KVSecretData struct {
	Data map[string]interface{} `json:"data"`
}

func NewVault(address, token string) *Vault {
	return &Vault{
		Address: address,
		Token:   token,
	}
}

func (v *Vault) GetSecrets(path string) ([]KVSecret, error) {
	cfg := api.DefaultConfig()
	cfg.Address = v.Address
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, logs.Local().Errorf("vault: %v", err)
	}
	client.SetToken(v.Token)
	data, err := client.Logical().Read(path)
	if err != nil {
		return nil, logs.Local().Errorf("vault: %v", err)
	}
	if data == nil {
		return nil, logs.Local().Errorf("vault: %v", "no data returned")
	}
	if data.Data == nil {
		return nil, logs.Local().Errorf("vault: %v", "no data returned")
	}

	return parseSecrets(data.Data)
}

func parseSecrets(data map[string]interface{}) ([]KVSecret, error) {
	var secrets []KVSecret
	for k, v := range data {
		if v == nil {
			continue
		}
		if v.(string) == "data" {
			s, err := parseSecrets(v.(map[string]interface{}))
			if err != nil {
				return nil, logs.Local().Errorf("vault: %v", err)
			}
			secrets = append(secrets, s...)
		}
		secrets = append(secrets, KVSecret{
			Key:   k,
			Value: v.(string),
		})
	}
	return secrets, nil
}
