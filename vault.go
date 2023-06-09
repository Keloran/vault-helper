package vault_helper

import (
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/hashicorp/vault/api"
)

type Vault struct {
	Address       string
	Token         string
	Secrets       []KVSecret
	LeaseDuration int
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

func (v *Vault) GetSecrets(path string) error {
	cfg := api.DefaultConfig()
	cfg.Address = v.Address
	client, err := api.NewClient(cfg)
	if err != nil {
		return logs.Local().Errorf("vault: %v", err)
	}
	client.SetToken(v.Token)
	data, err := client.Logical().Read(path)
	if err != nil {
		return logs.Local().Errorf("vault: %v", err)
	}
	if data == nil {
		return logs.Local().Errorf("vault: %v", "no data returned")
	}
	if data.Data == nil {
		return logs.Local().Errorf("vault: %v", "no data returned")
	}

	if data.LeaseDuration != 0 {
		v.LeaseDuration = data.LeaseDuration
	}

	secrets, err := ParseData(data.Data, "data")
	if err != nil {
		return logs.Local().Errorf("vault: %v", err)
	}

	v.Secrets = secrets
	return nil
}

func (v *Vault) GetSecret(key string) (string, error) {
	for _, s := range v.Secrets {
		if s.Key == key {
			return s.Value, nil
		}
	}
	return "", logs.Local().Errorf("vault: %v", "key not found")
}

func ParseData(data map[string]interface{}, filterName string) ([]KVSecret, error) {
	var secrets []KVSecret
	for k, v := range data {
		if v == nil {
			continue
		}
		if v.(string) == filterName {
			s, err := ParseData(v.(map[string]interface{}), filterName)
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
