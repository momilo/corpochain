// Public Domain (-) 2011 The Golly Authors.
// See the Golly UNLICENSE file for details.

// Package crypto implements utility functions for handling passwords and
// strings securely.
package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"
)

// PBKDF2 implements the Password-Based Key Derivation Function 2 as specified
// in PKCS #5 v2.0 from RSA Laboratories <http://www.ietf.org/rfc/rfc2898.txt>.
func PBKDF2(hashfunc func() hash.Hash, password, salt []byte, iterations, keylen int) (key []byte) {

	var (
		digest          []byte
		i, j, k, length int
	)

	key = make([]byte, keylen)
	slice := key

	hash := hmac.New(hashfunc, password)
	hashlen := hash.Size()
	scratch := make([]byte, 4)

	for keylen > 0 {

		if hashlen > keylen {
			length = keylen
		} else {
			length = hashlen
		}

		i += 1

		scratch[0] = byte(i >> 24)
		scratch[1] = byte(i >> 16)
		scratch[2] = byte(i >> 8)
		scratch[3] = byte(i)

		hash.Write(salt)
		hash.Write(scratch)

		digest = hash.Sum(nil)
		hash.Reset()

		for j = 0; j < length; j++ {
			slice[j] = digest[j]
		}

		for k = 1; k < iterations; k++ {
			hash.Write(digest)
			digest = hash.Sum(nil)
			for j = 0; j < length; j++ {
				slice[j] ^= digest[j]
			}
			hash.Reset()
		}

		keylen -= length
		slice = slice[length:]

	}

	return

}

var (
	passwordHashFunc   = sha256.New
	passwordIterations = 20000
	passwordKeyLength  = 80
	passwordSaltLength = 40
)

func SetPasswordHashFunc(hashfunc func() hash.Hash) {
	passwordHashFunc = hashfunc
}

func SetPasswordIterations(n int) {
	passwordIterations = n
}

func SetPasswordKeyLength(n int) {
	passwordKeyLength = n
}

func SetPasswordSaltLength(n int) {
	passwordSaltLength = n
}

type Password struct {
	Iterations int
	Key        []byte
	KeyLength  int
	Salt       []byte
}

// NewPassword abstracts away the complexity of generating PBKDF2-based
// passwords.
func NewPassword(secret string) (password *Password, err error) {
	salt := make([]byte, passwordSaltLength)
	_, err = rand.Read(salt)
	if err != nil {
		return
	}
	return &Password{
		Iterations: passwordIterations,
		Key: PBKDF2(
			passwordHashFunc, []byte(secret), salt, passwordIterations, passwordKeyLength),
		KeyLength: passwordKeyLength,
		Salt:      salt,
	}, nil
}

func (password *Password) Validate(secret string) (valid bool) {
	return password.ValidateWithHashFunc(secret, passwordHashFunc)
}

func (password *Password) ValidateWithHashFunc(secret string, hashfunc func() hash.Hash) (valid bool) {
	expected := PBKDF2(hashfunc, []byte(secret), password.Salt, password.Iterations, password.KeyLength)
	if len(expected) != len(password.Key) {
		return
	}
	return subtle.ConstantTimeCompare(password.Key, expected) == 1
}

var ironHMAC = sha256.New

func SetIronHMAC(hmac func() hash.Hash) {
	ironHMAC = hmac
}

// IronString returns "tamper-resistant" strings.
func IronString(name, value string, key []byte, duration int64) string {
	if duration > 0 {
		value = fmt.Sprintf("%d:%s", time.Now().UnixNano()+duration, value)
	}
	message := fmt.Sprintf("%s|%s", strings.Replace(name, "|", `\|`, -1), value)
	hash := hmac.New(ironHMAC, key)
	hash.Write([]byte(message))
	mac := base64.URLEncoding.EncodeToString(hash.Sum(nil))
	return fmt.Sprintf("%s:%s", mac, value)
}

func GetIronValue(name, value string, key []byte, timestamped bool) (val string, ok bool) {
	split := strings.SplitN(value, ":", 2)
	if len(split) != 2 {
		return
	}
	expected, value := []byte(split[0]), split[1]
	message := fmt.Sprintf("%s|%s", strings.Replace(name, "|", `\|`, -1), value)
	hash := hmac.New(ironHMAC, key)
	hash.Write([]byte(message))
	digest := hash.Sum(nil)
	mac := make([]byte, base64.URLEncoding.EncodedLen(len(digest)))
	base64.URLEncoding.Encode(mac, digest)
	if subtle.ConstantTimeCompare(mac, expected) != 1 {
		return
	}
	if timestamped {
		split = strings.SplitN(value, ":", 2)
		if len(split) != 2 {
			return
		}
		timestring, value := split[0], split[1]
		timestamp, err := strconv.ParseInt(timestring, 10, 64)
		if err != nil {
			return
		}
		if time.Now().UnixNano() > timestamp {
			return
		}
		return value, true
	}
	return value, true
}
