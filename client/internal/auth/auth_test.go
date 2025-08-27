package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := getConfigPath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".data-vault", "auth.json")

	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestSaveAndLoadJWT(t *testing.T) {
	testJWT := "test.jwt.token"
	testUsername := "testuser"

	err := SaveJWT(testJWT, testUsername)
	if err != nil {
		t.Fatalf("Expected no error saving JWT, got %v", err)
	}

	loadedJWT, err := LoadJWT()
	if err != nil {
		t.Fatalf("Expected no error loading JWT, got %v", err)
	}

	if loadedJWT != testJWT {
		t.Errorf("Expected JWT %s, got %s", testJWT, loadedJWT)
	}

	err = ClearJWT()
	if err != nil {
		t.Fatalf("Expected no error clearing JWT, got %v", err)
	}
}

func TestLoadJWTNotExists(t *testing.T) {
	_ = ClearJWT()

	jwt, err := LoadJWT()
	if err != nil {
		t.Fatalf("Expected no error when JWT doesn't exist, got %v", err)
	}

	if jwt != "" {
		t.Errorf("Expected empty JWT when file doesn't exist, got %s", jwt)
	}
}

func TestClearJWT(t *testing.T) {
	testJWT := "test.jwt.token"
	testUsername := "testuser"
	err := SaveJWT(testJWT, testUsername)
	if err != nil {
		t.Fatalf("Expected no error saving JWT, got %v", err)
	}

	err = ClearJWT()
	if err != nil {
		t.Fatalf("Expected no error clearing JWT, got %v", err)
	}

	jwt, err := LoadJWT()
	if err != nil {
		t.Fatalf("Expected no error loading JWT after clear, got %v", err)
	}

	if jwt != "" {
		t.Errorf("Expected empty JWT after clearing, got %s", jwt)
	}
}

func TestClearJWTNotExists(t *testing.T) {
	_ = ClearJWT()

	err := ClearJWT()
	if err != nil {
		t.Fatalf("Expected no error clearing non-existent JWT, got %v", err)
	}
}
