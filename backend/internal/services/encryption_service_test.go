package services

import (
	"testing"
)

func TestEncryptionService_Encrypt_Decrypt(t *testing.T) {
	// Create service with 32-byte key
	key := "12345678901234567890123456789012"
	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	// Test data
	plaintext := "my-secret-api-key-12345"

	// Encrypt
	encrypted, err := service.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Verify encrypted is different from plaintext
	if encrypted == plaintext {
		t.Error("Encrypted text should be different from plaintext")
	}

	// Decrypt
	decrypted, err := service.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify decrypted matches original
	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match. Got %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptionService_InvalidKeyLength(t *testing.T) {
	// Test with wrong key length
	_, err := NewEncryptionService("short-key")
	if err == nil {
		t.Error("Expected error for invalid key length")
	}

	_, err = NewEncryptionService("12345678901234567890123456789012345") // 35 bytes
	if err == nil {
		t.Error("Expected error for invalid key length")
	}
}

func TestEncryptionService_EmptyString(t *testing.T) {
	key := "12345678901234567890123456789012"
	service, _ := NewEncryptionService(key)

	// Test empty plaintext
	_, err := service.Encrypt("")
	if err == nil {
		t.Error("Expected error for empty plaintext")
	}

	// Test empty ciphertext
	_, err = service.Decrypt("")
	if err == nil {
		t.Error("Expected error for empty ciphertext")
	}
}

func TestEncryptionService_ConsistentEncryption(t *testing.T) {
	key := "12345678901234567890123456789012"
	service, _ := NewEncryptionService(key)

	plaintext := "test-data"

	// Encrypt twice
	encrypted1, _ := service.Encrypt(plaintext)
	encrypted2, _ := service.Encrypt(plaintext)

	// Should be different (due to random nonce)
	if encrypted1 == encrypted2 {
		t.Error("Two encryptions of same data should produce different ciphertexts")
	}

	// But both should decrypt to same plaintext
	decrypted1, _ := service.Decrypt(encrypted1)
	decrypted2, _ := service.Decrypt(encrypted2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}

func BenchmarkEncryption(b *testing.B) {
	key := "12345678901234567890123456789012"
	service, _ := NewEncryptionService(key)
	plaintext := "my-secret-api-key-12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Encrypt(plaintext)
	}
}

func BenchmarkDecryption(b *testing.B) {
	key := "12345678901234567890123456789012"
	service, _ := NewEncryptionService(key)
	plaintext := "my-secret-api-key-12345"
	encrypted, _ := service.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Decrypt(encrypted)
	}
}
