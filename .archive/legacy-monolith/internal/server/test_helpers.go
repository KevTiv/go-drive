package server

import (
	"bytes"
	"encoding/json"
	"io"
)

func toJSONReader(v interface{}) io.Reader {
	body, _ := json.Marshal(v)
	return bytes.NewBuffer(body)
}
