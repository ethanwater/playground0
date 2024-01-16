package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"vivian.app/utils"
)

const (
	charset     string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	authKeySize int    = 5
)

//var (
//	generatedAuthKeys = metrics.NewCounter(
//		"generatedAuthKeys",
//		"total number of generated authentication keys",
//	)
//
//	validAuthKeys = metrics.NewCounter(
//		"validAuthKeys",
//		"totoal number of successfully matching keys",
//	)
//)

type T interface {
	GenerateAuthKey2FA(context.Context, string) (string, error)
	VerifyAuthKey2FA(context.Context, string, string) (bool, error)
}

func GenerateAuthKey2FA(ctx context.Context, s *utils.VivianLogger) (string, error) {
	source := rand.New(rand.NewSource(time.Now().Unix()))
	var authKey strings.Builder

	for i := 0; i < authKeySize; i++ {
		sample := source.Intn(len(charset))
		authKey.WriteString(string(charset[sample]))
	}

	hashChannel := make(chan string, 1)
	go func() {
		authKeyHash, err := HashKeyphrase(ctx, authKey.String())
		if err != nil {
			s.LogError("failure hashing the authentication key", err)
			hashChannel <- ""
			return
		}
		hashChannel <- authKeyHash
	}()
	hash := <-hashChannel

	if hash == "" {
		s.LogError("failure hashing the authentication key", errors.New("empty hash"))
		return "", nil
	}

	s.LogSuccess(fmt.Sprintf("authentication key generated: %v", authKey.String()))
	//t.sender.Get().SendVerificationCodeEmail(ctx, email, authKey.String())
	return hash, nil
}

func VerifyAuthKey2FA(ctx context.Context, authkey_hash, input string, s *utils.VivianLogger) (bool, error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if SanitizeCheck(input) {
		status := bcrypt.CompareHashAndPassword([]byte(authkey_hash), []byte(input))
		if status != nil {
			s.LogWarning("invalid key")
			//t.Logger(ctx).Debug("vivian: [warning]", "key invalid", http.StatusNotAcceptable)
			return status == nil, status
		} else {
			s.LogSuccess("verified key")
			//t.Logger(ctx).Debug("vivian: [ok]", "key verified", status == nil, "status", http.StatusOK)
			return status == nil, status
		}
	}

	return false, nil
}
