package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crosschainkeeper "github.com/cosmos/cosmos-sdk/x/crosschain/keeper"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/bnb-chain/greenfield/app"
	paymentmodulekeeper "github.com/bnb-chain/greenfield/x/payment/keeper"
	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissionmodulekeeper "github.com/bnb-chain/greenfield/x/permission/keeper"
	permissionmoduletypes "github.com/bnb-chain/greenfield/x/permission/types"
	spkeeper "github.com/bnb-chain/greenfield/x/sp/keeper"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	storageMaccPerms = map[string][]string{
		authtypes.Minter:               {authtypes.Minter},
		authtypes.FeeCollectorName:     {authtypes.Minter, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		sptypes.ModuleName:             {authtypes.Staking},
		types.ModuleName:               {authtypes.Staking},
		paymentmoduletypes.ModuleName:  {},
	}
)

type StorageDepKeepers struct {
	PaymentKeeper *paymentmodulekeeper.Keeper
	SpKeeper      *spkeeper.Keeper
	BankKeeper    *bankkeeper.BaseKeeper
	AccountKeeper *authkeeper.AccountKeeper
}

func StorageKeeper(t testing.TB) (*keeper.Keeper, StorageDepKeepers, sdk.Context) {
	storeKeys := sdk.NewKVStoreKeys(authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey, evidencetypes.StoreKey,
		ibctransfertypes.StoreKey, icahosttypes.StoreKey, capabilitytypes.StoreKey, group.StoreKey,
		icacontrollertypes.StoreKey,
		crosschaintypes.StoreKey,
		sptypes.StoreKey,
		paymentmoduletypes.StoreKey,
		permissionmoduletypes.StoreKey,
		oracletypes.StoreKey, types.StoreKey)

	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)

	storeKey := sdk.NewKVStoreKey(types.StoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)

	stateStore.MountStoreWithDB(storeKeys[paramstypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[authtypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[banktypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[paymentmoduletypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[sptypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[crosschaintypes.StoreKey], storetypes.StoreTypeIAVL, db)

	stateStore.MountStoreWithDB(tkeys[paramstypes.TStoreKey], storetypes.StoreTypeTransient, nil)

	require.NoError(t, stateStore.LoadLatestVersion())

	cdcConfig := app.MakeEncodingConfig()
	cdc := cdcConfig.Marshaler

	paramKeeper := paramskeeper.NewKeeper(cdc, types.Amino, storeKeys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	paramKeeper.Subspace(authtypes.ModuleName)
	paramKeeper.Subspace(banktypes.ModuleName)
	paramKeeper.Subspace(authz.ModuleName)
	paramKeeper.Subspace(sptypes.ModuleName)
	paramKeeper.Subspace(permissionmoduletypes.ModuleName)
	paramKeeper.Subspace(paymentmoduletypes.ModuleName)
	paramKeeper.Subspace(crosschaintypes.ModuleName)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"StorageParams",
	)

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		storeKeys[authtypes.StoreKey],
		GetSubspace(paramKeeper, authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		storageMaccPerms,
	)

	authzKeeper := authzkeeper.NewKeeper(
		storeKeys[authz.ModuleName],
		cdc,
		baseapp.NewMsgServiceRouter(),
		accountKeeper,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		storeKeys[banktypes.StoreKey],
		accountKeeper,
		GetSubspace(paramKeeper, banktypes.ModuleName),
		nil,
	)

	spKeeper := spkeeper.NewKeeper(
		cdc,
		storeKeys[sptypes.ModuleName],
		storeKeys[sptypes.MemStoreKey],
		GetSubspace(paramKeeper, sptypes.ModuleName),
		accountKeeper,
		bankKeeper,
		authzKeeper,
	)

	paymentKeeper := paymentmodulekeeper.NewKeeper(
		cdc,
		storeKeys[paymentmoduletypes.StoreKey],
		storeKeys[paymentmoduletypes.MemStoreKey],
		GetSubspace(paramKeeper, paymentmoduletypes.ModuleName),

		bankKeeper,
		accountKeeper,
		spKeeper,
	)
	crossChainKeeper := crosschainkeeper.NewKeeper(
		cdc,
		storeKeys[crosschaintypes.StoreKey],
		GetSubspace(paramKeeper, crosschaintypes.ModuleName),
	)

	permissionKeeper := permissionmodulekeeper.NewKeeper(
		cdc,
		storeKeys[permissionmoduletypes.ModuleName],
		storeKeys[permissionmoduletypes.MemStoreKey],
		GetSubspace(paramKeeper, permissionmoduletypes.ModuleName),
		accountKeeper,
	)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		accountKeeper,
		spKeeper,
		paymentKeeper,
		permissionKeeper,
		crossChainKeeper,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())
	accountKeeper.SetParams(ctx, authtypes.DefaultParams())
	spKeeper.SetParams(ctx, sptypes.DefaultParams())
	paymentKeeper.SetParams(ctx, paymentmoduletypes.DefaultParams())

	// Initialize module accounts
	paymentModulePool := accountKeeper.GetModuleAccount(ctx, paymentmoduletypes.ModuleName)
	accountKeeper.SetModuleAccount(ctx, paymentModulePool)

	amount := sdk.NewIntFromUint64(1e19)
	err := bankKeeper.MintCoins(ctx, authtypes.Minter, sdk.Coins{sdk.Coin{
		Denom:  "BNB",
		Amount: amount,
	}})
	if err != nil {
		panic("mint coins error")
	}

	return k, StorageDepKeepers{
		SpKeeper:      spKeeper,
		PaymentKeeper: paymentKeeper,
		BankKeeper:    &bankKeeper,
		AccountKeeper: &accountKeeper,
	}, ctx
}
