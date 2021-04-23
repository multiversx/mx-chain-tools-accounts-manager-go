package process

import "errors"

// ErrNilAccountsProcessor signals that a nil accounts processor has been provided
var ErrNilAccountsProcessor = errors.New("nil accounts processor")

// ErrNilAccountsIndexer signals that a nil accounts indexer handler has been provided
var ErrNilAccountsIndexer = errors.New("nil accounts indexer")

// ErrNilReindexer signals that a nil reindexer has been provided
var ErrNilReindexer = errors.New("nil reindexer")

// ErrNilCloner signals that a nil cloner has been provided
var ErrNilCloner = errors.New("nil cloner")
