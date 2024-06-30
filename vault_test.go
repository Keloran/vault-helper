package vault_helper

import (
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestLocalSecrets(t *testing.T) {
  mockLogical := &MockLogical{
		MockRead: func(path string) (*api.Secret, error) {
			// Return a mock Secret for testing purposes
			return &api.Secret{
				Data: map[string]interface{}{
					"keycloak-realm": "test-realm",
				},
			}, nil
		},
	}

	mockClient := &MockVaultClient{
		MockLogical: func() LogicalClient {
			return mockLogical
		},
		MockSetToken: func(token string) {
			// Do nothing or validate the token
		},
	}

	v := &Vault{
		Client:  mockClient,
		Address: "mockaddress",
		Token:   "mocktoken",
	}

  // Test path secret
  err := v.GetSecrets("./test_data.json")
  assert.Nil(t, err)

  localSecret, err := v.GetSecret("keycloak-realm")
  assert.Nil(t, err)
  assert.Equal(t, "test_realm", localSecret)

  // test remote
  err = v.GetSecrets("mockpath")
  assert.Nil(t, err)

  remoteSecret, err := v.GetSecret("keycloak-realm")
  assert.Nil(t, err)
  assert.Equal(t, "test-realm", remoteSecret)
}

func TestParseJSON(t *testing.T) {
  v := &Vault{}

  err := v.GetLocalSecrets("./test_data.json")
  assert.Nil(t, err)


  secret, err := v.GetSecret("keycloak-realm")
  assert.Nil(t, err)

  assert.Equal(t, "test_realm", secret)
}

func TestParseDATA(t *testing.T) {
  v := &Vault{}

  err := v.GetLocalSecrets("./test_data")
  assert.Nil(t, err)


  secret, err := v.GetSecret("keycloak-secret")
  assert.Nil(t, err)

  assert.Equal(t, "test_secret", secret)
}

func TestGetRemoteSecrets(t *testing.T) {
	mockLogical := &MockLogical{
		MockRead: func(path string) (*api.Secret, error) {
			// Return a mock Secret for testing purposes
			return &api.Secret{
				Data: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			}, nil
		},
	}

	mockClient := &MockVaultClient{
		MockLogical: func() LogicalClient {
			return mockLogical
		},
		MockSetToken: func(token string) {
			// Do nothing or validate the token
		},
	}

	v := &Vault{
		Client:  mockClient,
		Address: "mockaddress",
		Token:   "mocktoken",
	}

	err := v.GetRemoteSecrets("mockpath")
	assert.Nil(t, err)
	// Add more assertions based on the expected behavior
}

func TestNewVault(t *testing.T) {
	v := NewVault("mockaddress", "mocktoken")
	assert.NotNil(t, v)
}

func TestGetSecret(t *testing.T) {
	mockLogical := &MockLogical{
		MockRead: func(path string) (*api.Secret, error) {
			// Return a mock Secret for testing purposes
			return &api.Secret{
				Data: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			}, nil
		},
	}

	mockClient := &MockVaultClient{
		MockLogical: func() LogicalClient {
			return mockLogical
		},
		MockSetToken: func(token string) {
			// Do nothing or validate the token
		},
	}

	v := &Vault{
		Client:  mockClient,
		Address: "mockaddress",
		Token:   "mocktoken",
	}

	err := v.GetRemoteSecrets("mockpath")
	assert.Nil(t, err)

	secret, err := v.GetSecret("key1")
	assert.Nil(t, err)
	assert.Equal(t, "value1", secret)
}
