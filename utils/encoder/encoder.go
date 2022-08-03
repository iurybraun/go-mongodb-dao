package encoder

import (
	"math/rand"
	cr "crypto/rand"
	"encoding/hex"
	"crypto/md5"
	"math/big"
	"io"
	"github.com/iurybraun/go-mongodb-dao/utils/logger"
	"fmt"
)

func GetRandomString(length int) string {
	bytes := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQSTUVWXYZ")
	var result []byte

	seed, _ := cr.Int(cr.Reader, big.NewInt(2<<32))
	r := rand.New(rand.NewSource(seed.Int64()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func MD5(v interface{}) string {
	hash := md5.New()
	switch v.(type) {
	case string:
		hash.Write([]byte(v.(string)))
	case io.Reader:
		io.Copy(hash, v.(io.Reader))
	default:
		logger.Failed(fmt.Sprintf("unexpectedly type: %T", v))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
