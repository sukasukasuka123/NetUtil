package hashutil

import (
	"crypto/sha1"
	"crypto/sha256"
	"hash/fnv"
)

func Sha256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func Sha1(data []byte) []byte {
	h := sha1.Sum(data)
	return h[:]
}

func FNVHash32(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

func FNVHashIndex(data []byte, n int) int {
	return int(FNVHash32(data) % uint32(n))
}
