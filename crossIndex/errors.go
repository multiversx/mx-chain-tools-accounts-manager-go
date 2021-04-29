package crossIndex

import "errors"

// ErrNilElasticClient signals that a nil elastic client has been provided
var ErrNilElasticClient = errors.New("nil elastic search client")
