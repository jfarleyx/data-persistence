package main

import (
	"fmt"
	"testing"
)

// TO EXECUTE TESTS:
// Run all tests: 	go test -run ''
// Run test group: 	go test -run TestPartitionManager
// Run sub test:  	go test -run TestPartitionManager/TestBuildPartitionMap
func TestPartitionManager(t *testing.T) {
	fmt.Printf("Running test group: %s\n", t.Name())

	// TEST SET UP //
	var pm PartitionManager

	// TESTS //
	t.Run("TestBuildPartitionMap", func(t *testing.T) {
		fmt.Printf("Running test: %s\n", t.Name())
		dbs, err := getDatabases()
		if err != nil {
			t.Fatalf("Expected to receive databases, received error: %v\n", err)
		}

		pm = NewPartitionManager(dbs)
		/*
			for k, v := range pm.partitionMap {
				fmt.Printf("key: %v; db name: %s\n", k, v.Name)
			}
		*/

		if len(pm.partitionMap) == 0 {
			t.Fatal("Expected > 0 partitions, received 0")
		}
	})

	t.Run("TestGetPartitionKeyFromString", func(t *testing.T) {
		fmt.Printf("Running test: %s\n", t.Name())
		res := pm.GetPartitionKeyFromString("apple")
		if res != 65 {
			t.Errorf("Expected rune value of 65, received: %v", res)
		}
	})

	t.Run("TestGetPartitionKeyFromString", func(t *testing.T) {
		fmt.Printf("Running test: %s\n", t.Name())
		res := pm.GetDatabaseByPartitionKey(65)
		if res == nil {
			t.Error("Expected to receive Database struct, received nil")
		}
		if err := res.db.Ping(); err != nil {
			t.Error("Expected database connection to be open, ping failed")
		}
	})

	t.Run("TestGetDatabaseByPartitionString", func(t *testing.T) {
		fmt.Printf("Running test: %s\n", t.Name())
		res := pm.GetDatabaseByPartitionString("apple")
		if res == nil {
			t.Error("Expected to receive Database struct, received nil")
		}
		if err := res.db.Ping(); err != nil {
			t.Error("Expected database connection to be open, ping failed")
		}
	})

	t.Run("TestGetDatabaseByName", func(t *testing.T) {
		fmt.Printf("Running test: %s\n", t.Name())
		res := pm.GetDatabaseByName("enrollment1.db")
		if res == nil {
			t.Error("Expected to receive Database struct, received nil")
		}
		if err := res.db.Ping(); err != nil {
			t.Error("Expected database connection to be open, ping failed")
		}
	})

	// TEST TEAR DOWN //
	pm.CloseConnections()
}

// HELPER FUNCTIONS //
func getDatabases() ([]*Database, error) {
	db1, err := NewDatabase("enrollment1.db", "./enrollment1.db", 65, 77)
	if err != nil {
		return nil, err
	}

	db2, err := NewDatabase("enrollment2.db", "./enrollment2.db", 78, 90)
	if err != nil {
		return nil, err
	}

	dbs := make([]*Database, 2)
	dbs[0] = db1
	dbs[1] = db2

	return dbs, nil
}
