package auth_test

import (
	"path/filepath"

	"github.com/louiss0/cobra-cli-template/auth"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Service", func() {
	It("registers users and creates their task files", func() {
		dataDir := GinkgoT().TempDir()
		service := auth.NewService(dataDir)

		err := service.Register("alice")

		assert.NoError(err)
		assert.FileExists(filepath.Join(dataDir, "tasks", "alice.json"))
	})

	It("signs in and reports the active user", func() {
		dataDir := GinkgoT().TempDir()
		service := auth.NewService(dataDir)

		err := service.Register("alice")
		assert.NoError(err)

		err = service.SignIn("alice")
		assert.NoError(err)

		username, err := service.CurrentUser()

		assert.NoError(err)
		assert.Equal("alice", username)
	})

	It("signs out the active user", func() {
		dataDir := GinkgoT().TempDir()
		service := auth.NewService(dataDir)

		err := service.Register("alice")
		assert.NoError(err)

		err = service.SignIn("alice")
		assert.NoError(err)

		err = service.SignOut()
		assert.NoError(err)

		_, err = service.CurrentUser()

		assert.ErrorIs(err, auth.ErrNotSignedIn)
	})
})
