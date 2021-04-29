package mappings

import (
	"bytes"
	"encoding/json"
)

type Object map[string]interface{}

// ToBuffer will convert an Object to a *bytes.Buffer
func (o *Object) ToBuffer() *bytes.Buffer {
	objectBytes, _ := json.Marshal(o)

	buff := &bytes.Buffer{}
	_, _ = buff.Write(objectBytes)

	return buff
}
