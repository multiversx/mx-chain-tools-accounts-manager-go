package process

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	data2 "github.com/multiversx/mx-chain-es-indexer-go/data"
	core2 "github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/tidwall/gjson"
)

const (
	numOfShards    = uint32(3)
	envS3URL       = "S3_URL"
	envBucketName  = "S3_BUCKET_NAME"
	envS3AccessKey = "S3_ACCESS_LEY"
	envS3SecretKey = "S3_SECRET_KEY"

	egldBalanceFilePrefix           = "egld_balance"
	legacyDelegationStateFilePrefix = "legacy_delegation_state"
	directStakeFilePrefix           = "direct_stake"
	delegatedInfoPrefix             = "delegated_info"
)

type s3Balances struct {
	sClient         *s3Client
	pubKeyConverter core.PubkeyConverter
}

func NewS3Balances() (*s3Balances, error) {
	url := os.Getenv(envS3URL)
	if url == "" {
		return nil, errors.New("S3_URL env empty")
	}

	bucketName := os.Getenv(envBucketName)
	if bucketName == "" {
		return nil, errors.New("S3_BUCKET_NAME env empty")
	}

	accessKey := os.Getenv(envS3AccessKey)
	if accessKey == "" {
		return nil, errors.New("S3_ACCESS_LEY env empty")
	}

	secretKey := os.Getenv(envS3SecretKey)
	if secretKey == "" {
		return nil, errors.New("S3_SECRET_KEY env empty")
	}

	sClient, err := newS3Client(bucketName, url, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	converter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	if err != nil {
		return nil, err
	}

	return &s3Balances{
		sClient:         sClient,
		pubKeyConverter: converter,
	}, nil
}

func (b *s3Balances) GetBalancesForEpoch(epoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
	accountsWithEgld, err := b.getAddressesWithEgld(epoch)
	if err != nil {
		return nil, err
	}

	// get direct stake
	accountsWithDirectStake, err := b.getAccountsWithDirectStake(epoch)
	if err != nil {
		return nil, err
	}

	// get delegation (staking providers)
	accountsWithDelegation, err := b.getDelegatorsAccounts(epoch)
	if err != nil && !(strings.Contains(err.Error(), "NoSuchKey") && epoch < 239) {
		return nil, err
	}

	// get legacy delegation
	accountsWithLegacyDelegation, err := b.getFullActiveAndWaitingListAccounts(epoch)
	if err != nil {
		return nil, err
	}

	allAccountsWithStake, _ := mergeAccounts(accountsWithLegacyDelegation, accountsWithDirectStake, accountsWithDelegation, nil, nil)
	calculateTotalStakeForAccountsAndTotalUnDelegated(allAccountsWithStake)

	for address, accountWithEgld := range accountsWithEgld {
		addressBytes, errC := b.pubKeyConverter.Decode(address)
		if errC != nil {
			return nil, errC
		}

		_, found := allAccountsWithStake[address]
		if !found {
			allAccountsWithStake[address] = &data.AccountInfoWithStakeValues{
				AccountInfo: data2.AccountInfo{
					Address:                  address,
					Nonce:                    accountWithEgld.Nonce,
					Balance:                  accountWithEgld.Balance,
					BalanceNum:               core2.ComputeBalanceAsFloat(accountWithEgld.Balance),
					TotalBalanceWithStake:    accountWithEgld.Balance,
					TotalBalanceWithStakeNum: core2.ComputeBalanceAsFloat(accountWithEgld.Balance),
					ShardID:                  sharding.ComputeShardID(addressBytes, 3),
				},
			}

			continue
		}

		totalBalanceWithStake, totalBalanceWithStakeNum := computeTotalBalance(accountWithEgld.Balance, allAccountsWithStake[address].TotalStake)

		allAccountsWithStake[address].Address = address
		allAccountsWithStake[address].Nonce = accountWithEgld.Nonce
		allAccountsWithStake[address].Balance = accountWithEgld.Balance
		allAccountsWithStake[address].BalanceNum = core2.ComputeBalanceAsFloat(accountWithEgld.Balance)
		allAccountsWithStake[address].TotalBalanceWithStake = totalBalanceWithStake
		allAccountsWithStake[address].TotalBalanceWithStakeNum = totalBalanceWithStakeNum
		allAccountsWithStake[address].ShardID = sharding.ComputeShardID(addressBytes, 3)
	}

	return allAccountsWithStake, nil
}

func (b *s3Balances) getAddressesWithEgld(epoch uint32) (map[string]*alteredAccount.AlteredAccount, error) {
	egldBalancesG := make(map[string]*alteredAccount.AlteredAccount)
	for shardID := uint32(0); shardID < numOfShards; shardID++ {
		egldBalances, err := b.getEgldBalancesForShard(shardID, epoch)
		if err != nil {
			return nil, err
		}

		egldBalancesG = mergeMaps(egldBalancesG, egldBalances)
	}

	egldBalances, err := b.getEgldBalancesForShard(core.MetachainShardId, epoch)
	if err != nil {
		return nil, err
	}

	egldBalancesG = mergeMaps(egldBalancesG, egldBalances)

	return egldBalancesG, nil
}

func (b *s3Balances) getEgldBalancesForShard(shardID, epoch uint32) (map[string]*alteredAccount.AlteredAccount, error) {
	egldBalanceBytes, err := b.sClient.GetFile(prepareFileName(egldBalanceFilePrefix, shardID, epoch))
	if err != nil {
		return nil, err
	}

	egldBalances := make(map[string]*alteredAccount.AlteredAccount)
	err = json.Unmarshal(egldBalanceBytes, &egldBalances)

	return egldBalances, err
}

func mergeMaps(m1, m2 map[string]*alteredAccount.AlteredAccount) map[string]*alteredAccount.AlteredAccount {
	for key, value := range m2 {
		m1[key] = value
	}

	return m1
}

func prepareFileName(prefix string, shardID, epoch uint32) string {
	return fmt.Sprintf("%s_%d_%d", prefix, shardID, epoch)
}

func (b *s3Balances) getAccountsWithDirectStake(epoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
	addressesWithStateBytes, err := b.sClient.GetFile(prepareFileName(directStakeFilePrefix, core.MetachainShardId, epoch))
	if err != nil {
		return nil, err
	}

	list := gjson.Get(string(addressesWithStateBytes), "list")
	accountsInfo := make([]data.StakedInfo, 0)
	err = json.Unmarshal([]byte(list.String()), &accountsInfo)
	if err != nil {
		return nil, err
	}

	accountsStake := make(map[string]*data.AccountInfoWithStakeValues)
	for _, acct := range accountsInfo {
		accountsStake[acct.Address] = &data.AccountInfoWithStakeValues{
			StakeInfo: data.StakeInfo{
				ValidatorsActive:    acct.Staked,
				ValidatorsActiveNum: core2.ComputeBalanceAsFloat(acct.Staked),
				ValidatorTopUp:      acct.TopUp,
				ValidatorTopUpNum:   core2.ComputeBalanceAsFloat(acct.TopUp),
			},
		}
	}

	return accountsStake, nil
}

func (b *s3Balances) getDelegatorsAccounts(epoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
	addressesWithStateBytes, err := b.sClient.GetFile(prepareFileName(delegatedInfoPrefix, core.MetachainShardId, epoch))
	if err != nil {
		return nil, err
	}

	accountsInfo := make([]data.DelegatorStake, 0)

	list := gjson.Get(string(addressesWithStateBytes), "list")
	err = json.Unmarshal([]byte(list.String()), &accountsInfo)
	if err != nil {
		log.Warn("cannot unmarshal accounts info", "error", err.Error())
		return nil, err
	}

	accountsStake := make(map[string]*data.AccountInfoWithStakeValues)
	for _, acct := range accountsInfo {
		accountsStake[acct.DelegatorAddress] = &data.AccountInfoWithStakeValues{
			StakeInfo: data.StakeInfo{
				Delegation:    acct.Total,
				DelegationNum: core2.ComputeBalanceAsFloat(acct.Total),
			},
		}
	}

	log.Info("delegators accounts", "num", len(accountsStake))

	return accountsStake, nil
}
