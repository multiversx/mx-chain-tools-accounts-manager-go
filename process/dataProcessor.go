package process

import (
	"time"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/mappings"
)

const (
	numAddressesInBulk = 2000
)

type dataProcessor struct {
	accountsIndexer   AccountsIndexerHandler
	accountsProcessor AccountsProcessorHandler
	cloner            Cloner
}

// NewDataProcessor will create a new instance of dataProcessor
func NewDataProcessor(
	accountsIndexer AccountsIndexerHandler,
	accountsProcessor AccountsProcessorHandler,
	cloner Cloner,
) (*dataProcessor, error) {
	return &dataProcessor{
		accountsIndexer:   accountsIndexer,
		accountsProcessor: accountsProcessor,
		cloner:            cloner,
	}, nil
}

// ProcessAccountsData will process accounts data
func (dp *dataProcessor) ProcessAccountsData() error {
	accountsRest, addresses, err := dp.accountsProcessor.GetAllAccountsWithStake()
	if err != nil {
		return err
	}

	accountsES := dp.getAccountsESDatabase(addresses)
	preparedAccounts := dp.accountsProcessor.PrepareAccountsForReindexing(accountsES, accountsRest)

	newIndex, err := dp.cloneAccountsIndex()
	if err != nil {
		return err
	}

	defer logExecutionTime(time.Now(), "Indexed modified accounts")

	log.Info("Accounts to index", "total", len(preparedAccounts))

	return dp.accountsIndexer.IndexAccounts(preparedAccounts, newIndex)
}

func (dp *dataProcessor) cloneAccountsIndex() (string, error) {
	defer logExecutionTime(time.Now(), "Cloned accounts index")

	newIndex, err := dp.accountsProcessor.ComputeClonedAccountsIndex()
	if err != nil {
		return "", err
	}

	err = dp.cloner.CloneIndex(accountsIndex, newIndex, mappings.AccountsCloned.ToBuffer())
	if err != nil {
		return "", err
	}

	return newIndex, nil
}

func (dp *dataProcessor) getAccountsESDatabase(addresses []string) map[string]*data.AccountInfoWithStakeValues {
	defer logExecutionTime(time.Now(), "Fetched accounts from elasticseach database")

	accountsES := make(map[string]*data.AccountInfoWithStakeValues)
	for idx := 0; idx < len(addresses); idx += numAddressesInBulk {
		from := idx
		to := idx + numAddressesInBulk

		if to > len(addresses) {
			to = len(addresses)
		}

		newSliceOfAddresses := make([]string, numAddressesInBulk)
		copy(newSliceOfAddresses, addresses[from:to])
		accounts, errGet := dp.accountsIndexer.GetAccounts(newSliceOfAddresses, accountsIndex)
		if errGet != nil {
			log.Warn("dataProcessor.getAccountsESDatabase: cannot get accounts", "error", errGet)
			continue
		}
		mergeAccountsMaps(accountsES, accounts)
	}

	return accountsES
}

func mergeAccountsMaps(dst, src map[string]*data.AccountInfoWithStakeValues) {
	for key, value := range src {
		dst[key] = value
	}
}

func logExecutionTime(start time.Time, message string) {
	log.Info(message, "duration in seconds", time.Since(start).Seconds())
}
