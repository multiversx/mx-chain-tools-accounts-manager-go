package process

import (
	"time"

	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/mappings"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
)

var log = logger.GetOrCreate("process")

type clonerDataProcessor struct {
	accountsIndexer   AccountsIndexerHandler
	accountsProcessor AccountsProcessorHandler
	cloner            Cloner
}

// NewClonerDataProcessor will create a new instance of clonerDataProcessor
func NewClonerDataProcessor(
	accountsIndexer AccountsIndexerHandler,
	accountsProcessor AccountsProcessorHandler,
	cloner Cloner,
) (*clonerDataProcessor, error) {
	if check.IfNil(accountsIndexer) {
		return nil, ErrNilAccountsIndexer
	}
	if check.IfNil(accountsProcessor) {
		return nil, ErrNilAccountsProcessor
	}
	if check.IfNil(cloner) {
		return nil, ErrNilCloner
	}

	return &clonerDataProcessor{
		accountsIndexer:   accountsIndexer,
		accountsProcessor: accountsProcessor,
		cloner:            cloner,
	}, nil
}

// ProcessAccountsData will process accounts data
func (dp *clonerDataProcessor) ProcessAccountsData() error {
	accountsRest, addresses, err := dp.accountsProcessor.GetAllAccountsWithStake()
	if err != nil {
		return err
	}

	accountsES, err := dp.getAccountsESDatabase(addresses)
	if err != nil {
		return err
	}

	preparedAccounts := core.MergeElasticAndRestAccounts(accountsES, accountsRest)

	newIndex, err := dp.cloneAccountsIndex()
	if err != nil {
		return err
	}

	defer logExecutionTime(time.Now(), "Indexed modified accounts")

	log.Info("accounts to index", "total", len(preparedAccounts))

	return dp.accountsIndexer.IndexAccounts(preparedAccounts, newIndex)
}

func (dp *clonerDataProcessor) cloneAccountsIndex() (string, error) {
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

func (dp *clonerDataProcessor) getAccountsESDatabase(addresses []string) (map[string]*data.AccountInfoWithStakeValues, error) {
	defer logExecutionTime(time.Now(), "Fetched accounts from elasticseach database")

	return dp.accountsIndexer.GetAccounts(addresses, accountsIndex)
}

func logExecutionTime(start time.Time, message string) {
	log.Info(message, "duration in seconds", time.Since(start).Seconds())
}
