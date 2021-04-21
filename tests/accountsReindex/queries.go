package accountsReindex

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type object = map[string]interface{}

func encodeQuery(query object) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(query); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding query: %s", err.Error())
	}

	return buff, nil
}

func getAll() *bytes.Buffer {
	obj := object{
		"query": object{
			"match_all": object{},
		},
	}

	encoded, _ := encodeQuery(obj)

	return &encoded
}
