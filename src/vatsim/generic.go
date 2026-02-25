package vatsim

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"vatusa-cobalt/config"
)

func API2Request(uri string, method string, data []byte) ([]byte, error) {
	url := fmt.Sprintf("%s%s", config.VatsimApi2URL(), uri)

	client := &http.Client{}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "vatusa/cobalt")
	req.Header.Set("X-API-Key", config.VatsimApi2Key())
	req.Header.Set("x-identifier", config.VatsimApi2Identifier())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
