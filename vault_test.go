package vault_helper

import (
  "context"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
  "github.com/testcontainers/testcontainers-go"
  "github.com/testcontainers/testcontainers-go/modules/vault"
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
  ctx := context.Background()
  vaultContainer, err := vault.RunContainer(ctx,
    testcontainers.WithImage("hashicorp/vault:1.13.0"),
    vault.WithToken("root-token"),
    vault.WithInitCommand("secrets enable transit", "write -f transit/keys/my-key"),
    vault.WithInitCommand("kv put secret/test foo1=bar"))
  assert.Nil(t, err)
  defer func() {
    err := vaultContainer.Terminate(ctx)
    assert.Nil(t, err)
  }()

  address, err := vaultContainer.HttpHostAddress(ctx)
  assert.Nil(t, err)

  v := NewVault(address, "root-token")
	err = v.GetRemoteSecrets("secret/test")
	assert.Nil(t, err)

	secret, err := v.GetSecret("foo1")
	assert.Nil(t, err)
	assert.Equal(t, "bar", secret)
}
