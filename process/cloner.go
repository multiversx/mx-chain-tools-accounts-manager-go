package process

import (
	"errors"
	"fmt"
	"strings"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	log = logger.GetOrCreate("process")

	errNewIndexAlreadyExits = errors.New("new index already exists")
)

const (
	backOffTime = time.Second * 10
	maxBackOff  = time.Minute * 5
)

type cloner struct {
	backOffTime   time.Duration
	elasticClient ElasticClientHandler
}

// NewCloner will create a new instance of a cloner
func NewCloner(ec ElasticClientHandler) (*cloner, error) {
	return &cloner{
		elasticClient: ec,
	}, nil
}

// CloneIndex will clone a given index
func (c *cloner) CloneIndex(index, newIndex string) error {
TRY:
	cloned, err := c.elasticClient.CloneIndex(index, newIndex)

	switch {
	case cloned && err != nil:
		for errUnset := c.elasticClient.UnsetReadOnly(index); errUnset != nil; {
			c.backOffAndWarn(err, "c.elasticClient.UnsetReadOnly: cannot unset readonly option")
		}
	case !cloned && err != nil:
		if checkIfErrorIsAlreadyExits(err, newIndex) {
			return fmt.Errorf("cannot do clone operation of the index: %w", errNewIndexAlreadyExits)
		}
		c.backOffAndWarn(err, "c.elasticClient.CloneIndex: cannot clone index")
		goto TRY
	default:
		c.backOffTime = 0
		return nil
	}

	return nil
}

func (c *cloner) backOffAndWarn(err error, reason string) {
	log.Warn(reason, "received back off:", err.Error())

	c.increaseBackOffTime()
	time.Sleep(c.backOffTime)
}

func (c *cloner) increaseBackOffTime() {
	if c.backOffTime == 0 {
		c.backOffTime = backOffTime
		return
	}
	if c.backOffTime >= maxBackOff {
		return
	}

	c.backOffTime += c.backOffTime / 5
}

func checkIfErrorIsAlreadyExits(err error, newIndex string) bool {
	return strings.Contains(err.Error(), "resource_already_exists_exception") &&
		strings.Contains(err.Error(), newIndex)
}
