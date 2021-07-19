package main

import (
	"strings"
)

type PartitionManager struct {
	partitionMap map[rune]*Database
	DBs          []*Database
}

func NewPartitionManager(dbs []*Database) PartitionManager {
	pm := PartitionManager{
		partitionMap: make(map[rune]*Database),
		DBs:          dbs,
	}

	for _, v := range dbs {
		pm.buildPartitionMap(v)
	}

	return pm
}

func (pm *PartitionManager) buildPartitionMap(db *Database) {
	for x := db.PartitionStart; x <= db.PartitionEnd; x++ {
		pm.partitionMap[x] = db
	}
}

func (pm *PartitionManager) GetPartitionKeyFromString(input string) rune {
	x := []rune(strings.ToUpper(strings.TrimSpace(input)[0:1]))
	return x[0]
}

func (pm *PartitionManager) GetDatabaseByPartitionKey(key rune) *Database {
	return pm.partitionMap[key]
}

func (pm *PartitionManager) GetDatabaseByPartitionString(input string) *Database {
	return pm.GetDatabaseByPartitionKey(pm.GetPartitionKeyFromString(input))
}

func (pm *PartitionManager) GetDatabaseByName(name string) *Database {
	for i := range pm.DBs {
		if pm.DBs[i].Name == name {
			return pm.DBs[i]
		}
	}
	return nil
}

func (pm *PartitionManager) CloseConnections() {
	for i := range pm.DBs {
		pm.DBs[i].Close()
	}
}
