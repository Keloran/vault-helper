package vault_helper

import (
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/hashicorp/vault/api"
  "time"
)

type VaultDetails struct {
  Address string
  Token string

  CredPath string
  DetailsPath string

  ExpireTime time.Time
}

type VaultHelper interface {
	GetSecrets(path string) error
	GetSecret(key string) (string, error)
	Secrets() []KVSecret
	LeaseDuration() int
}

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
	Client    VaultClient
	Address   string
	Token     string
	Lease     int
	KVSecrets []KVSecret
}

type Details struct {
  CredPath string `env:"VAULT_CRED_PATH" envDefault:"secret/data/chewedfeed/creds"`
  DetailsPath string `env:"VAULT_DETAILS_PATH" envDefault:"secret/data/chewedfeed/details"`

  ExpireTime time.Time
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
	if path == "" {
		return logs.Local().Errorf("path: %s, err: %s", path, "no path provided")
	}

	v.Client.SetToken(v.Token)
	data, err := v.Client.Logical().Read(path)
	if err != nil {
		return logs.Local().Errorf("path: %s, err: %v", path, err)
	}
	if data == nil {
		return logs.Local().Errorf("path: %s, err: %s", path, "no data returned")
	}
	if data.Data == nil {
		return logs.Local().Errorf("path: %s, err: %s", path, "no data returned")
	}

	if data.LeaseDuration != 0 {
		v.Lease = data.LeaseDuration
	}

	secrets, err := ParseData(data.Data, "data")
	if err != nil {
		return logs.Local().Errorf("path: %s, err: %v", path, err)
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
	return "", logs.Local().Errorf("key: '%s' not found", key)
}

func ParseData(data map[string]interface{}, filterName string) ([]KVSecret, error) {
	var secrets []KVSecret
	for k, v := range data {
		if v == nil {
			continue
		}

		switch value := v.(type) {
		case string:
			if value == filterName {
				continue
			}
			secrets = append(secrets, KVSecret{
				Key:   k,
				Value: value,
			})
		case map[string]interface{}:
			s, err := ParseData(value, filterName)
			if err != nil {
				return nil, logs.Local().Errorf("data: %+v, filter: %s, err: %v", data, filterName, err)
			}
			secrets = append(secrets, s...)
		}
	}
	return secrets, nil
}

func (v *Vault) Secrets() []KVSecret {
	return v.KVSecrets
}

func (v *Vault) LeaseDuration() int {
	return v.Lease
}
