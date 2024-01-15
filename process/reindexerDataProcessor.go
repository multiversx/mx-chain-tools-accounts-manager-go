package process

import "github.com/multiversx/mx-chain-core-go/core/check"

type reindexerDataProcessor struct {
	accountsProcessor AccountsProcessorHandler
	reindexer         Reindexer

	s3Proc *s3Balances
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

	s3Proc, err := NewS3Balances()
	if err != nil {
		return nil, err
	}

	return &reindexerDataProcessor{
		accountsProcessor: accountsProcessor,
		reindexer:         reindexer,
		s3Proc:            s3Proc,
	}, nil
}

// ProcessAccountsData will process accounts data
func (dp *reindexerDataProcessor) ProcessAccountsData() error {
	epoch, err := dp.accountsProcessor.GetCurrentEpoch()
	if err != nil {
		return err
	}

	accountsRest, err := dp.accountsProcessor.GetAllAccountsWithStake(epoch)
	if err != nil {
		return err
	}

	newIndex, err := dp.accountsProcessor.ComputeClonedAccountsIndex(epoch)
	if err != nil {
		return err
	}

	return dp.reindexer.ReindexAccounts(accountsIndex, newIndex, accountsRest)
}

func (dp *reindexerDataProcessor) IndexDataFromS3(epoch uint32) error {
	allAccounts, err := dp.s3Proc.GetBalancesForEpoch(epoch)
	if err != nil {
		return err
	}

	newIndex, err := dp.accountsProcessor.ComputeClonedAccountsIndex(epoch)
	if err != nil {
		return err
	}

	return dp.reindexer.IndexAllAccounts(newIndex, allAccounts)
}
