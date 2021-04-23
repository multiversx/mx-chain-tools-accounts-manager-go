package process

import "github.com/ElrondNetwork/elrond-go-logger/check"

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
	accountsRest, _, err := dp.accountsProcessor.GetAllAccountsWithStake()
	if err != nil {
		return err
	}

	newIndex, err := dp.accountsProcessor.ComputeClonedAccountsIndex()
	if err != nil {
		return err
	}

	return dp.reindexer.ReindexAccounts("accounts", newIndex, accountsRest)
}
