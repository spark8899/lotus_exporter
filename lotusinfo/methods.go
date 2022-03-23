package lotusinfo

var (
	MethodAccount  = []string{"Constructor", "Constructor"}
	MethodInit     = []string{"Constructor", "Exec"}
	MethodCron     = []string{"Constructor", "EpochTick"}
	MethodReward   = []string{"Constructor", "AwardBlockReward", "ThisEpochReward", "UpdateNetworkKPI"}
	MethodMultisig = []string{"Constructor", "Propose", "Approve", "Cancel", "AddSigner", "RemoveSigner", "SwapSigner",
		"ChangeNumApprovalsThreshold", "LockBalance"}
	MethodPaymentChannel = []string{"Constructor", "UpdateChannelState", "Settle", "Collect"}
	MethodStorageMarket  = []string{"Constructor", "AddBalance", "WithdrawBalance", "PublishStorageDeals",
		"VerifyDealsForActivation", "ActivateDeals", "OnMinerSectorsTerminate", "ComputeDataCommitment", "CronTick"}
	MethodStoragePower = []string{"Constructor", "CreateMiner", "UpdateClaimedPower", "EnrollCronEvent",
		"OnEpochTickEnd", "UpdatePledgeTotal", "Deprecated1", "SubmitPoRepForBulkVerify", "CurrentTotalPower"}
	MethodStorageMiner = []string{"Constructor", "ControlAddresses", "ChangeWorkerAddress", "ChangePeerID",
		"SubmitWindowedPoSt", "PreCommitSector", "ProveCommitSector", "ExtendSectorExpiration", "TerminateSectors",
		"DeclareFaults", "DeclareFaultsRecovered", "OnDeferredCronEvent", "CheckSectorProven", "ApplyRewards",
		"ReportConsensusFault", "WithdrawBalance", "ConfirmSectorProofsValid", "ChangeMultiaddrs", "CompactPartitions",
		"CompactSectorNumbers", "ConfirmUpdateWorkerKey", "RepayDebt", "ChangeOwnerAddress", "DisputeWindowedPoSt"}
	MethodVerifiedRegistry = []string{"Constructor", "AddVerifier", "RemoveVerifier", "AddVerifiedClient", "UseBytes",
		"RestoreBytes"}
	MethodMessageType = map[string][]string{"account": MethodAccount, "init": MethodInit, "cron": MethodCron,
		"reward": MethodReward, "multisig": MethodMultisig, "paymentchannel": MethodPaymentChannel,
		"storagemarket": MethodStorageMarket, "storagepower": MethodStoragePower, "storageminer": MethodStorageMiner,
		"verifiedregistry": MethodVerifiedRegistry}
)
