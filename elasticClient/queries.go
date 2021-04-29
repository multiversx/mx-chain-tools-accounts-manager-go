package elasticClient

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type objectsMap = map[string]interface{}

func encode(obj objectsMap) (*bytes.Buffer, error) {
	buff := &bytes.Buffer{}
	if err := json.NewEncoder(buff).Encode(obj); err != nil {
		return nil, fmt.Errorf("error encoding : %w", err)
	}

	return buff, nil
}

func settingsWriteEncoded(state bool) *bytes.Buffer {
	obj := objectsMap{
		"settings": objectsMap{
			"index.blocks.write": state,
		},
	}

	encodedObj, _ := encode(obj)

	return encodedObj
}

func getDocumentsByIDsQueryEncoded(ids []string) *bytes.Buffer {
	interfaceSlice := make([]interface{}, len(ids))
	for idx := range ids {
		interfaceSlice[idx] = objectsMap{
			"_id":     ids[idx],
			"_source": true,
		}
	}

	obj := objectsMap{
		"docs": interfaceSlice,
	}
	encodedObj, _ := encode(obj)

	return encodedObj
}
