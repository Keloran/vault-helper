package vault_helper

import (
	"fmt"
	"github.com/hashicorp/vault/api"
)

type MockVaultClient struct {
	MockLogical  func() LogicalClient
	MockSetToken func(string)
}

func (m *MockVaultClient) Logical() LogicalClient {
	if m.MockLogical != nil {
		return m.MockLogical()
	}
	return nil
}

func (m *MockVaultClient) SetToken(token string) {
	if m.MockSetToken != nil {
		m.MockSetToken(token)
	}
}

type MockLogical struct {
	MockRead func(string) (*api.Secret, error)
}

func (m *MockLogical) Read(path string) (*api.Secret, error) {
	if m.MockRead != nil {
		return m.MockRead(path)
	}
	return nil, nil
}

type MockVaultHelper struct {
	KVSecrets []KVSecret
	Lease     int
}

func (m *MockVaultHelper) GetRemoteSecrets(path string) error {
	if path == "" {
		return fmt.Errorf("path not found: %s", path)
	}

	return nil
}

func (m *MockVaultHelper) GetLocalSecrets(path string) error {
	if path == "" {
		return fmt.Errorf("path not found: %s", path)
	}

	return nil
}

func (m *MockVaultHelper) GetSecret(key string) (string, error) {
	for _, s := range m.Secrets() {
		for s.Key == key {
			return s.Value, nil
		}
	}

	return "", fmt.Errorf("key: '%s' not found", key)
}

func (m *MockVaultHelper) Secrets() []KVSecret {
	return m.KVSecrets
}
func (m *MockVaultHelper) LeaseDuration() int {
	return m.Lease
}
