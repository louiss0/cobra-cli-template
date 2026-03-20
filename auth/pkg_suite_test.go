package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	tAssert "github.com/stretchr/testify/assert"
)

var assert *tAssert.Assertions

func TestAuth(t *testing.T) {
	assert = tAssert.New(GinkgoT())
	RunSpecs(t, "Auth Suite")
}
