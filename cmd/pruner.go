package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/state"
	tmstore "github.com/tendermint/tendermint/store"
	db "github.com/tendermint/tm-db"
)

// load db
// load app store and prune
// if immutable tree is not deletable we should import and export current state

func pruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune [path_to_home]",
		Short: "prune data from the application store and block store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pruneTMData(args[0]); err != nil {
				return err
			}

			if err := pruneAppState(args[0]); err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}

func pruneAppState(home string) error {
	dbType := db.BackendType(backend)

	dbDir := rootify(dataDir, home)
	// Get BlockStore
	appDB, err := db.NewDB("application", dbType, dbDir)
	if err != nil {
		return err
	}

	// TODO: cleanup app state
	// appStore := rootmulti.NewStore(appDB)
	//  get the latest version to prune latest - X versions
	// latest := rootmulti.GetLatestVersion(appDB)
	// pruneFrom := latest - int64(versions)
	// allVersions := appStore.GetAllVersions()
	// appStore.PruneStores()

	if err := appDB.ForceCompact(nil, nil); err != nil {
		return err
	}

	//create a new app store
	return nil
}

// pruneTMData prunes the tendermint blocks and state based on the amount of blocks to keep
func pruneTMData(home string) error {
	dbType := db.BackendType(backend)

	dbDir := rootify(dataDir, home)

	// Get BlockStore
	blockStoreDB, err := db.NewDB("blockstore", dbType, dbDir)
	if err != nil {
		return err
	}
	blockStore := tmstore.NewBlockStore(blockStoreDB)

	// Get StateStore
	stateDB, err := db.NewDB("state", dbType, dbDir)
	if err != nil {
		return err
	}

	stateStore := state.NewStore(stateDB)

	base := blockStore.Base()

	pruneHeight := blockStore.Height() - int64(blocks)

	// prune block store
	blocks, err = blockStore.PruneBlocks(pruneHeight)
	if err != nil {
		return err
	}

	// prune state store
	err = stateStore.PruneStates(base, pruneHeight)
	if err != nil {
		return err
	}

	if err := blockStoreDB.ForceCompact(nil, nil); err != nil {
		return err
	}

	if err := stateDB.ForceCompact(nil, nil); err != nil {
		return err
	}

	return nil
}

// Utils

func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
