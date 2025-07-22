package util

import (
	cryptoRand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"time"
)

// RandomBuildImageName generates a random PR number and commit SHA for image naming
func RandomBuildImageName() (pr int, sha string, err error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Random PR number between 1 and 999
	pr = r.Intn(999) + 1

	// Random commit hash (8 characters)
	var buf [4]byte
	if _, err := cryptoRand.Read(buf[:]); err != nil {
		return 0, "", err
	}

	sha = hex.EncodeToString(buf[:])
	return
}
