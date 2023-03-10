package utils_test

import (
	"ngantri/utils"
	"testing"
)

func TestConnectDB(t *testing.T) {
	db, err := utils.ConnectDB()
	if err != nil {
		t.Fatalf("error connecting to database: %v", err)
	}

	if db == nil {
		t.Fatalf("database is nil")
	}

	t.Logf("database connection is successful")
}
