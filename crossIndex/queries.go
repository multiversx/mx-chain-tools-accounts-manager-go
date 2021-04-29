package crossIndex

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type object = map[string]interface{}

func EncodeQuery(query object) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(query); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding query: %s", err.Error())
	}

	return buff, nil
}

func GetAll() *bytes.Buffer {
	obj := object{
		"query": object{
			"match_all": object{},
		},
	}

	encoded, _ := EncodeQuery(obj)

	return &encoded
}
