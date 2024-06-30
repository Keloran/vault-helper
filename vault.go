package vault_helper

import (
  "encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/hashicorp/vault/api"
  "os"
  "strings"
  "time"
)

type VaultDetails struct {
  Address string
  Token string

  CredPath string
  DetailsPath string
  LocalSecretsPath string

  ExpireTime time.Time
}

type VaultHelper interface {
  GetSecrets(path string) error
	GetRemoteSecrets(path string) error
  GetLocalSecrets(path string) error
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
  if strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/") {
    return v.GetLocalSecrets(path)
  }

  return v.GetRemoteSecrets(path)
}

func (v *Vault) GetLocalSecrets(path string) error {
  if path == "" {
    return logs.Local().Errorf("path: %s, err: %s", path, "no path provided")
  }

  file, err := os.ReadFile(path)
  if err != nil {
    return logs.Local().Errorf("reading of local file: %s, err: %v", path, err)
  }

  if strings.HasSuffix(path, ".json") {
    jdata, err := ParseJSON(file)
    if err != nil {
      return logs.Local().Errorf("failed to parse local JSON file: %s, err: %v", string(file), err)
    }
    secrets, err := ParseData(jdata, "")
    if err != nil {
      return logs.Local().Errorf("failed to parse post json data: %+v, err: %v", jdata, err)
    }

    v.KVSecrets = secrets
  } else {
    fstrng := string(file)
    data, err := ParseDATA(fstrng)
    if err != nil {
      return logs.Local().Errorf("failed to parse local DATA file: %s, err: %v", fstrng, err)
    }
    secrets, err := ParseData(data, "")
    if err != nil {
      return logs.Local().Errorf("failed to parse post local data: %+v, err: %v", data, err)
    }
    v.KVSecrets = secrets
  }

  return nil
}

func ParseJSON(data []byte) (map[string]interface{}, error) {
	var parsedData map[string]interface{}
	err := json.Unmarshal(data, &parsedData)
	if err != nil {
    return nil, logs.Local().Errorf("error unmarshalling JSON: %v", err)
	}
	return parsedData, nil
}

func (v *Vault) GetRemoteSecrets(path string) error {
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

func ParseDATA(data string) (map[string]interface{}, error) {
	parsedData := make(map[string]interface{})
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove the "map[" and "]" parts from the value string
		if strings.HasPrefix(value, "map[") {
			value = strings.TrimPrefix(value, "map[")
			value = strings.TrimSuffix(value, "]")

			innerMap := make(map[string]interface{})
			innerParts := strings.Split(value, " ")
			for _, innerPart := range innerParts {
				innerKV := strings.SplitN(innerPart, ":", 2)
				if len(innerKV) != 2 {
					continue
				}
				innerMap[innerKV[0]] = innerKV[1]
			}
			parsedData[key] = innerMap
		} else {
			parsedData[key] = value
		}
	}

	return parsedData, nil
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
