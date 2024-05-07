package vault_helper

import "testing"

func TestMock(t *testing.T) {
	// Create a new MockVaultHelper
	mvh := &MockVaultHelper{
		KVSecrets: []KVSecret{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
	}

	// Test GetSecrets
	t.Run("Get Empty Secrets", func(t *testing.T) {
		if err := mvh.GetSecrets(""); err == nil {
			t.Error("Expected error, got nil")
		}
	})

	// Test GetSecret
	t.Run("Get Non-Existent Secret", func(t *testing.T) {
		if _, err := mvh.GetSecret("key2"); err == nil {
			t.Error("Expected error, got nil")
		}
	})

	// Test GetSecret
	t.Run("Get Existing Secret", func(t *testing.T) {
		if _, err := mvh.GetSecret("key1"); err != nil {
			t.Error("Expected nil, got error")
		}
	})
}
