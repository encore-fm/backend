// +build !ci

package systest

import (
	"os"
	"testing"

	"github.com/encore-fm/backend/config"
	"github.com/encore-fm/backend/db"
)

func TestMain(m *testing.M) {
	// Create and open db connection
	config.Setup()
	dbConn, err := db.New() // todo maybe write create and open db myself to avoid db package dependency
	if err != nil {
		panic(err)
	}
	client = dbConn.Client

	dropDB()
	setupDB()
	status := m.Run()

	os.Exit(status)
}
