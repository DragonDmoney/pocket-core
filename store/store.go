package store

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/pokt-network/pocket-core/store/rootmulti"
	"github.com/pokt-network/pocket-core/store/types"
)

func NewCommitMultiStore(db dbm.DB, cache bool) types.CommitMultiStore {
	return rootmulti.NewStore(db, cache)
}

func NewPruningOptionsFromString(strategy string) (opt PruningOptions) {
	switch strategy {
	case "nothing":
		opt = PruneNothing
	case "everything":
		opt = PruneEverything
	case "syncable":
		opt = PruneSyncable
	default:
		opt = PruneSyncable
	}
	return
}
