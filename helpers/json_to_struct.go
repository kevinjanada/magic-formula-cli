package helpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// JSONToStruct -- Parse http response type JSON to struct
func JSONToStruct(resp *http.Response, myStruct interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(body), myStruct)
	if err != nil {
		return err
	}
	return nil
}
