package utils

import (
	"testing"

	. "github.com/onsi/gomega"
)

const testdataSha = "bcdac44253ef5ea900b61357c0aeb9dc0254b0b5b0471986b713c99cc3040bf2"

func TestHashDirectories(t *testing.T) {
	g := NewWithT(t)
	hash, err := HashDirectories("testdata")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(hash).To(Equal(testdataSha))
}
