package process

import "github.com/multiversx/mx-chain-core-go/core/check"

type reindexerDataProcessor struct {
	accountsProcessor AccountsProcessorHandler
	reindexer         Reindexer
}

// NewReindexerDataProcessor will create a new instance of reindexerDataProcessor
func NewReindexerDataProcessor(
	accountsProcessor AccountsProcessorHandler,
	reindexer Reindexer,
) (*reindexerDataProcessor, error) {
	if check.IfNil(accountsProcessor) {
		return nil, ErrNilAccountsProcessor
	}
	if check.IfNil(reindexer) {
		return nil, ErrNilReindexer
	}

	return &reindexerDataProcessor{
		accountsProcessor: accountsProcessor,
		reindexer:         reindexer,
	}, nil
}

// ProcessAccountsData will process accounts data
func (dp *reindexerDataProcessor) ProcessAccountsData() error {
	epoch, err := dp.accountsProcessor.GetCurrentEpoch()
	if err != nil {
		return err
	}

	log.Info("Processing accounts data", "epoch", epoch)

	accountsRest, err := dp.accountsProcessor.GetAllAccountsWithStake(epoch)
	if err != nil {
		return err
	}

	newIndex, err := dp.accountsProcessor.ComputeClonedAccountsIndex(epoch)
	if err != nil {
		return err
	}

	log.Info("Trying  to create new index", "newIndex", newIndex, "epoch", epoch)

	return dp.reindexer.ReindexAccounts(accountsIndex, newIndex, accountsRest)
}
