package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HashDirectories computes the SHA256 hash for a set of directories.
func HashDirectories(dirs ...string) (string, error) {
	var allHashes []string
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				fileHash, err := hashFile(path)
				if err != nil {
					return err
				}
				allHashes = append(allHashes, fileHash)
			}

			return nil
		})

		if err != nil {
			return "", err
		}
	}

	// Sort all hashes and concatenate them.
	sort.Strings(allHashes)
	combinedHashData := strings.Join(allHashes, "")

	// Compute the final hash.
	hasher := sha256.New()
	hasher.Write([]byte(combinedHashData))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// hashFile returns the SHA256 hash of the file content.
func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
