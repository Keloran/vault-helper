package vault_helper

import (
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/hashicorp/vault/api"
)

type LogicalClient interface {
	Read(string) (*api.Secret, error)
}

type RealVaultClient struct {
	Client *api.Client
}

func (r *RealVaultClient) Logical() LogicalClient {
	return &RealLogicalClient{Logical: r.Client.Logical()}
}

func (r *RealVaultClient) SetToken(token string) {
	r.Client.SetToken(token)
}

type RealLogicalClient struct {
	Logical *api.Logical
}

func (r *RealLogicalClient) Read(path string) (*api.Secret, error) {
	return r.Logical.Read(path)
}

type VaultClient interface {
	Logical() LogicalClient
	SetToken(string)
}

type Vault struct {
	Client        VaultClient
	Address       string
	Token         string
	LeaseDuration int
	KVSecrets     []KVSecret
}

type KVSecret struct {
	Key   string
	Value string
}

type KVSecretData struct {
	Data map[string]interface{} `json:"data"`
}

func NewVault(address, token string) *Vault {
	cfg := api.DefaultConfig()
	cfg.Address = address
	client, _ := api.NewClient(cfg) // Handle error appropriately

	return &Vault{
		Client:  &RealVaultClient{Client: client},
		Address: address,
		Token:   token,
	}
}

func (v *Vault) GetSecrets(path string) error {
	v.Client.SetToken(v.Token)
	data, err := v.Client.Logical().Read(path)
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

	v.KVSecrets = secrets
	return nil
}

func (v *Vault) GetSecret(key string) (string, error) {
	for _, s := range v.KVSecrets {
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

func (v *Vault) Secrets() []KVSecret {
	return v.KVSecrets
}
