package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	"github.com/neilotoole/errgroup"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/tendermint/tendermint/state"
	tmstore "github.com/tendermint/tendermint/store"
	db "github.com/tendermint/tm-db"

	"github.com/binaryholdings/cosmos-pruner/internal/rootmulti"
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

			ctx := cmd.Context()
			errs, _ := errgroup.WithContext(ctx)
			var err error
			if tendermint {
				errs.Go(func() error {
					if err = pruneTMData(args[0]); err != nil {
						return err
					}
					return nil
				})
			}

			if cosmosSdk {
				err = pruneAppState(args[0])
				if err != nil {
					return err
				}
				return nil

			}

			return errs.Wait()
		},
	}
	return cmd
}

func pruneAppState(home string) error {

	// this has the potential to expand size, should just use state sync
	// dbType := db.BackendType(backend)

	dbDir := rootify(dataDir, home)

	o := opt.Options{
		DisableSeeksCompaction: true,
	}

	// Get BlockStore
	appDB, err := db.NewGoLevelDBWithOpts("application", dbDir, &o)
	if err != nil {
		return err
	}

	//TODO: need to get all versions in the store, setting randomly is too slow
	fmt.Println("pruning application state")

	// only mount keys from core sdk
	// todo allow for other keys to be mounted
	keys := types.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
	)

	if app == "osmosis" {
		osmoKeys := types.NewKVStoreKeys(
			"icahost",        //icahosttypes.StoreKey,
			"gamm",           // gammtypes.StoreKey,
			"lockup",         //lockuptypes.StoreKey,
			"incentives",     // incentivestypes.StoreKey,
			"epochs",         // epochstypes.StoreKey,
			"poolincentives", //poolincentivestypes.StoreKey,
			"authz",          //authzkeeper.StoreKey,
			"txfees",         // txfeestypes.StoreKey,
			"superfluid",     // superfluidtypes.StoreKey,
			"bech32ibc",      // bech32ibctypes.StoreKey,
			"wasm",           // wasm.StoreKey,
			"tokenfactory",   //tokenfactorytypes.StoreKey,
		)
		for key, value := range osmoKeys {
			keys[key] = value
		}
	} else if app == "cosmoshub" {
		cosmoshubKeys := types.NewKVStoreKeys(
			"liquidity",
			"feegrant",
			"authz",
			"icahost", // icahosttypes.StoreKey
		)
		for key, value := range cosmoshubKeys {
			keys[key] = value
		}
	} else if app == "terra" { // terra classic
		terraKeys := types.NewKVStoreKeys(
			"oracle",   // oracletypes.StoreKey,
			"market",   // markettypes.StoreKey,
			"treasury", //treasurytypes.StoreKey,
			"wasm",     // wasmtypes.StoreKey,
			"authz",    //authzkeeper.StoreKey,
			"feegrant", // feegrant.StoreKey
		)
		for key, value := range terraKeys {
			keys[key] = value
		}
	} else if app == "kava" {
		kavaKeys := types.NewKVStoreKeys(
			"feemarket", //feemarkettypes.StoreKey,
			"authz",     //authzkeeper.StoreKey,
			"kavadist",  //kavadisttypes.StoreKey,
			"auction",   //auctiontypes.StoreKey,
			"issuance",  //issuancetypes.StoreKey,
			"bep3",      //bep3types.StoreKey,
			//"pricefeed", //pricefeedtypes.StoreKey,
			//"swap",      //swaptypes.StoreKey,
			"cdp",       //cdptypes.StoreKey,
			"hard",      //hardtypes.StoreKey,
			"committee", //committeetypes.StoreKey,
			"incentive", //incentivetypes.StoreKey,
			"evmutil",   //evmutiltypes.StoreKey,
			"savings",   //savingstypes.StoreKey,
			"bridge",    //bridgetypes.StoreKey,
		)
		for key, value := range kavaKeys {
			keys[key] = value
		}

		delete(keys, "mint") // minttypes.StoreKey
	} else if app == "evmos" {
		evmosKeys := types.NewKVStoreKeys(
			"feegrant",   // feegrant.StoreKey,
			"authz",      // authzkeeper.StoreKey,
			"evm",        // evmtypes.StoreKey,
			"feemarket",  // feemarkettypes.StoreKey,
			"inflation",  // inflationtypes.StoreKey,
			"erc20",      // erc20types.StoreKey,
			"incentives", // incentivestypes.StoreKey,
			"epochs",     // epochstypes.StoreKey,
			"claims",     // claimstypes.StoreKey,
			"vesting",    // vestingtypes.StoreKey,
		)
		for key, value := range evmosKeys {
			keys[key] = value
		}
	} else if app == "gravitybridge" {
		gravitybridgeKeys := types.NewKVStoreKeys(
			"authz",     // authzkeeper.StoreKey,
			"gravity",   //  gravitytypes.StoreKey,
			"bech32ibc", // bech32ibctypes.StoreKey,
		)
		for key, value := range gravitybridgeKeys {
			keys[key] = value
		}
	} else if app == "sifchain" {
		sifchainKeys := types.NewKVStoreKeys(
			"feegrant",      // feegrant.StoreKey,
			"dispensation",  // disptypes.StoreKey,
			"ethbridge",     // ethbridgetypes.StoreKey,
			"clp",           // clptypes.StoreKey,
			"oracle",        // oracletypes.StoreKey,
			"tokenregistry", // tokenregistrytypes.StoreKey,
			"admin",         // admintypes.StoreKey,
		)
		for key, value := range sifchainKeys {
			keys[key] = value
		}
	} else if app == "starname" {
		starnameKeys := types.NewKVStoreKeys(
			"wasm",          // wasm.StoreKey,
			"configuration", // configuration.StoreKey,
			"starname",      // starname.DomainStoreKey,
		)
		for key, value := range starnameKeys {
			keys[key] = value
		}
	} else if app == "regen" {
		regenKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"authz",    // uthzkeeper.StoreKey,
		)
		for key, value := range regenKeys {
			keys[key] = value
		}
	} else if app == "akash" {
		akashKeys := types.NewKVStoreKeys(
			"authz",      // authzkeeper.StoreKey,
			"escrow",     // escrow.StoreKey,
			"deployment", // deployment.StoreKey,
			"market",     // market.StoreKey,
			"provider",   // provider.StoreKey,
			"audit",      // audit.StoreKey,
			"cert",       // cert.StoreKey,
			"inflation",  // inflation.StoreKey,
		)
		for key, value := range akashKeys {
			keys[key] = value
		}
	} else if app == "sentinel" {
		sentinelKeys := types.NewKVStoreKeys(
			"authz",        // authzkeeper.StoreKey,
			"distribution", // distributiontypes.StoreKey,
			"feegrant",     // feegrant.StoreKey,
			"custommint",   // customminttypes.StoreKey,
			"swap",         // swaptypes.StoreKey,
			"vpn",          // vpntypes.StoreKey,
		)
		for key, value := range sentinelKeys {
			keys[key] = value
		}
	} else if app == "emoney" {
		emoneyKeys := types.NewKVStoreKeys(
			"liquidityprovider", // lptypes.StoreKey,
			"issuer",            // issuer.StoreKey,
			"authority",         // authority.StoreKey,
			"market",            // market.StoreKey,
			//"market_indices",    // market.StoreKeyIdx,
			"buyback",   // buyback.StoreKey,
			"inflation", // inflation.StoreKey,
		)

		for key, value := range emoneyKeys {
			keys[key] = value
		}
	} else if app == "ixo" {
		ixoKeys := types.NewKVStoreKeys(
			"did",      // didtypes.StoreKey,
			"bonds",    // bondstypes.StoreKey,
			"payments", // paymentstypes.StoreKey,
			"project",  // projecttypes.StoreKey,
		)

		for key, value := range ixoKeys {
			keys[key] = value
		}
	} else if app == "juno" {
		junoKeys := types.NewKVStoreKeys(
			"authz",    // authzkeeper.StoreKey,
			"feegrant", // feegrant.StoreKey,
			"icahost",  // icahosttypes.StoreKey,
			"wasm",     // wasm.StoreKey,
		)

		for key, value := range junoKeys {
			keys[key] = value
		}
	} else if app == "likecoin" {
		likecoinKeys := types.NewKVStoreKeys(
			// common modules
			"feegrant", // feegrant.StoreKey,
			"authz",    // authzkeeper.StoreKey,
			// custom modules
			"iscn",     // iscntypes.StoreKey,
			"nft",      // nftkeeper.StoreKey,
		        "likenft",  // likenfttypes.StoreKey,
		)

		for key, value := range likecoinKeys {
			keys[key] = value
		}
	} else if app == "kichain" {
		kichainKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"authz",    // authzkeeper.StoreKey,
			"wasm",     // wasm.StoreKey,
		)

		for key, value := range kichainKeys {
			keys[key] = value
		}
	} else if app == "cyber" {
		cyberKeys := types.NewKVStoreKeys(
			"liquidity", // liquiditytypes.StoreKey,
			"feegrant",  // feegrant.StoreKey,
			"authz",     // authzkeeper.StoreKey,
			"bandwidth", // bandwidthtypes.StoreKey,
			"graph",     // graphtypes.StoreKey,
			"rank",      // ranktypes.StoreKey,
			"grid",      // gridtypes.StoreKey,
			"dmn",       // dmntypes.StoreKey,
			"wasm",      // wasm.StoreKey,
		)

		for key, value := range cyberKeys {
			keys[key] = value
		}
	} else if app == "cheqd" {
		cheqdKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"authz",    // authzkeeper.StoreKey,
			"cheqd",    // cheqdtypes.StoreKey,
		)

		for key, value := range cheqdKeys {
			keys[key] = value
		}
	} else if app == "stargaze" {
		stargazeKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"claim",    // claimmoduletypes.StoreKey,
			"alloc",    // allocmoduletypes.StoreKey,
			"authz",    // authzkeeper.StoreKey,
			"wasm",     // wasm.StoreKey,
		)

		for key, value := range stargazeKeys {
			keys[key] = value
		}
	} else if app == "bandchain" {
		bandchainKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"authz",    // authzkeeper.StoreKey,
			"oracle",   // oracletypes.StoreKey,
		)

		for key, value := range bandchainKeys {
			keys[key] = value
		}
	} else if app == "chihuahua" {
		chihuahuaKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"wasm",     // wasm.StoreKey,
		)

		for key, value := range chihuahuaKeys {
			keys[key] = value
		}
	} else if app == "bitcanna" {
		bitcannaKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"bcna",     // bcnamoduletypes.StoreKey,
		)

		for key, value := range bitcannaKeys {
			keys[key] = value
		}
	} else if app == "konstellation" {
		konstellationKeys := types.NewKVStoreKeys(
			"oracle", // racletypes.StoreKey,
			"wasm",   // wasm.StoreKey,
		)

		for key, value := range konstellationKeys {
			keys[key] = value
		}
	} else if app == "omniflixhub" {
		omniflixhubKeys := types.NewKVStoreKeys(
			"feegrant",    // feegrant.StoreKey,
			"authz",       // authzkeeper.StoreKey,
			"alloc",       // alloctypes.StoreKey,
			"onft",        // onfttypes.StoreKey,
			"marketplace", // marketplacetypes.StoreKey,
		)

		for key, value := range omniflixhubKeys {
			keys[key] = value
		}
	} else if app == "vidulum" {
		vidulumKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"vidulum",  // vidulummoduletypes.StoreKey,
		)

		for key, value := range vidulumKeys {
			keys[key] = value
		}
	} else if app == "provenance" {
		provenanceKeys := types.NewKVStoreKeys(
			"feegrant",  // feegrant.StoreKey,
			"authz",     // authzkeeper.StoreKey,
			"metadata",  // metadatatypes.StoreKey,
			"marker",    // markertypes.StoreKey,
			"attribute", // attributetypes.StoreKey,
			"name",      // nametypes.StoreKey,
			"msgfees",   // msgfeestypes.StoreKey,
			"wasm",      // wasm.StoreKey,
		)

		for key, value := range provenanceKeys {
			keys[key] = value
		}
	} else if app == "dig" {
		digKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"authz",    // authzkeeper.StoreKey,
			"wasm",     // wasm.StoreKey,
		)

		for key, value := range digKeys {
			keys[key] = value
		}
	} else if app == "comdex" {
		comdexKeys := types.NewKVStoreKeys(
			"wasm", // wasm.StoreKey,
		)

		for key, value := range comdexKeys {
			keys[key] = value
		}
	} else if app == "cerberus" {
		cerberusKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"authz",    // authzkeeper.StoreKey,
		)

		for key, value := range cerberusKeys {
			keys[key] = value
		}
	} else if app == "bitsong" {
		bitsongKeys := types.NewKVStoreKeys(
			"feegrant",               // feegrant.StoreKey,
			"authz",                  // authzkeeper.StoreKey,
			"packetfowardmiddleware", // routertypes.StoreKey,
			"fantoken",               // fantokentypes.StoreKey,
			"merkledrop",             // merkledroptypes.StoreKey,
		)

		for key, value := range bitsongKeys {
			keys[key] = value
		}
	} else if app == "assetmantle" {
		assetmantleKeys := types.NewKVStoreKeys(
			"feegrant",               // feegrant.StoreKey,
			"authz",                  // authzKeeper.StoreKey,
			"packetfowardmiddleware", // routerTypes.StoreKey,
			"icahost",                // icaHostTypes.StoreKey,
		)

		for key, value := range assetmantleKeys {
			keys[key] = value
		}
	} else if app == "fetchhub" {
		fetchhubKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"wasm",     // wasm.StoreKey,
			"authz",    // authzkeeper.StoreKey,
		)

		for key, value := range fetchhubKeys {
			keys[key] = value
		}
	} else if app == "persistent" {
		persistentKeys := types.NewKVStoreKeys(
			"halving",  // halving.StoreKey,
			"authz",    // sdkAuthzKeeper.StoreKey,
			"feegrant", // feegrant.StoreKey,
		)

		for key, value := range persistentKeys {
			keys[key] = value
		}
	} else if app == "cryptoorgchain" {
		cryptoorgchainKeys := types.NewKVStoreKeys(
			"feegrant",  // feegrant.StoreKey,
			"authz",     // authzkeeper.StoreKey,
			"chainmain", // chainmaintypes.StoreKey,
			"supply",    // supplytypes.StoreKey,
			"nft",       // nfttypes.StoreKey,
		)

		for key, value := range cryptoorgchainKeys {
			keys[key] = value
		}
	} else if app == "irisnet" {
		irisnetKeys := types.NewKVStoreKeys(
			"guardian", // guardiantypes.StoreKey,
			"token",    // tokentypes.StoreKey,
			"nft",      // nfttypes.StoreKey,
			"htlc",     // htlctypes.StoreKey,
			"record",   // recordtypes.StoreKey,
			"coinswap", // coinswaptypes.StoreKey,
			"service",  // servicetypes.StoreKey,
			"oracle",   // oracletypes.StoreKey,
			"random",   // randomtypes.StoreKey,
			"farm",     // farmtypes.StoreKey,
			"feegrant", // feegrant.StoreKey,
			"tibc",     // tibchost.StoreKey,
			"NFT",      // tibcnfttypes.StoreKey,
			"MT",       // tibcmttypes.StoreKey,
			"mt",       // mttypes.StoreKey,
		)

		for key, value := range irisnetKeys {
			keys[key] = value
		}
	} else if app == "axelar" {
		axelarKeys := types.NewKVStoreKeys(
			"feegrant",   // feegrant.StoreKey,
			"vote",       // voteTypes.StoreKey,
			"evm",        // evmTypes.StoreKey,
			"snapshot",   // snapTypes.StoreKey,
			"tss",        // tssTypes.StoreKey,
			"nexus",      // nexusTypes.StoreKey,
			"axelarnet",  // axelarnetTypes.StoreKey,
			"reward",     // rewardTypes.StoreKey,
			"permission", // permissionTypes.StoreKey,
		)

		for key, value := range axelarKeys {
			keys[key] = value
		}
	} else if app == "umee" {
		umeeKeys := types.NewKVStoreKeys(
			"feegrant", // feegrant.StoreKey,
			"authz",    // authzkeeper.StoreKey,
			"gravity",  // gravitytypes.StoreKey,
		)

		for key, value := range umeeKeys {
			keys[key] = value
		}
	} else if app == "desmos" {
		desmosKeys := types.NewKVStoreKeys(
			// common modules
			"feegrant",      // feegrant.StoreKey,
			"wasm",          // wasm.StoreKey,
			"authz",         // authzkeeper.StoreKey,
			// mainnet
			"profiles",      // profilestypes.StoreKey,
			// testnet
			"subspaces",     // subspacestypes.StoreKey,
			"posts",         // poststypes.StoreKey,
			"relationships", // relationshipstypes.StoreKey,
			"reports",       // reports.StoreKey,
			"reactions",     // reactions.StoreKey,
			"fees",          // fees.StoreKey,
		)

		for key, value := range desmosKeys {
			keys[key] = value
		}
	}

	// TODO: cleanup app state
	appStore := rootmulti.NewStore(appDB)

	for _, value := range keys {
		appStore.MountStoreWithDB(value, sdk.StoreTypeIAVL, nil)
	}

	err = appStore.LoadLatestVersion()
	if err != nil {
		return err
	}

	versions := appStore.GetAllVersions()

	v64 := make([]int64, len(versions))
	for i := 0; i < len(versions); i++ {
		v64[i] = int64(versions[i])
	}

	fmt.Println(len(v64))

	appStore.PruneHeights = v64[:len(v64)-10]

	appStore.PruneStores()

	fmt.Println("compacting application state")
	if err := appDB.ForceCompact(nil, nil); err != nil {
		return err
	}

	//create a new app store
	return nil
}

// pruneTMData prunes the tendermint blocks and state based on the amount of blocks to keep
func pruneTMData(home string) error {

	dbDir := rootify(dataDir, home)

	o := opt.Options{
		DisableSeeksCompaction: true,
	}

	// Get BlockStore
	blockStoreDB, err := db.NewGoLevelDBWithOpts("blockstore", dbDir, &o)
	if err != nil {
		return err
	}
	blockStore := tmstore.NewBlockStore(blockStoreDB)

	// Get StateStore
	stateDB, err := db.NewGoLevelDBWithOpts("state", dbDir, &o)
	if err != nil {
		return err
	}

	stateStore := state.NewStore(stateDB)

	base := blockStore.Base()

	pruneHeight := blockStore.Height() - int64(blocks)

	errs, _ := errgroup.WithContext(context.Background())
	errs.Go(func() error {
		fmt.Println("pruning block store")
		// prune block store
		blocks, err = blockStore.PruneBlocks(pruneHeight)
		if err != nil {
			return err
		}

		fmt.Println("compacting block store")
		if err := blockStoreDB.ForceCompact(nil, nil); err != nil {
			return err
		}

		return nil
	})

	fmt.Println("pruning state store")
	// prune state store
	err = stateStore.PruneStates(base, pruneHeight)
	if err != nil {
		return err
	}

	fmt.Println("compacting state store")
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
