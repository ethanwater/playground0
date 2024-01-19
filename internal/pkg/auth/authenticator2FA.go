package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"
	"vivian.app/internal/utils"
)

const (
	charset     string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	authKeySize int    = 5
)

type T interface {
	GenerateAuthKey2FA(context.Context, string) (string, error)
	VerifyAuthKey2FA(context.Context, string, string) (bool, error)
}

type HashManager struct {
	atomicValue atomic.Value
	flag        uint16
}

var HashManagerAtomic HashManager

func GenerateAuthKey2FA(ctx context.Context, s *utils.VivianLogger) ([]byte, error) {
	source := rand.New(rand.NewSource(time.Now().Unix()))
	var authKey strings.Builder

	for i := 0; i < authKeySize; i++ {
		sample := source.Intn(len(charset))
		authKey.WriteString(string(charset[sample]))
	}

	HashManagerAtomic.flag = 0

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		authKeyHash, err := HashKeyphrase(ctx, authKey.String())
		if err != nil {
			s.LogError("failure hashing the authentication key", err)
			return
		}
		HashManagerAtomic.atomicValue.Store([]byte(authKeyHash))
	}()
	wg.Wait()

	hash := HashManagerAtomic.atomicValue.Load().([]byte)

	if hash == nil {
		s.LogError("failure hashing the authentication key", errors.New("empty hash"))
		return []byte{}, nil
	}

	s.LogSuccess(fmt.Sprintf("authentication key generated: %v", authKey.String()))
	//t.sender.Get().SendVerificationCodeEmail(ctx, email, authKey.String())
	return hash, nil
}

func VerifyAuthKey2FA(ctx context.Context, key string, s *utils.VivianLogger) (bool, error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if HashManagerAtomic.flag == 1 {
		return false, nil
	}

	authkey_hash := HashManagerAtomic.atomicValue.Load()
	if authkey_hash == nil {
		// Handle the case where the value is nil
		s.LogWarning("hashChannel is not initialized")
		return false, nil
	}

	if SanitizeCheck(key) {
		status := bcrypt.CompareHashAndPassword(authkey_hash.([]byte), []byte(key))
		if status != nil {
			s.LogWarning("invalid key")
			return false, status
		} else {
			s.LogSuccess("verified key")
			return true, status
		}
	}

	return false, nil
}
