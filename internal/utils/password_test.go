package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "password123"
	
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	
	// Hash should be different each time
	hash2, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password123"
	wrongPassword := "wrongpassword"
	
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	
	// Correct password should match
	assert.True(t, CheckPasswordHash(password, hash))
	
	// Wrong password should not match
	assert.False(t, CheckPasswordHash(wrongPassword, hash))
	
	// Empty password should not match
	assert.False(t, CheckPasswordHash("", hash))
	
	// Empty hash should not match
	assert.False(t, CheckPasswordHash(password, ""))
}