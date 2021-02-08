package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

func httpRequest(method string, url string, query map[string]string, body interface{}, header map[string]string) (data []byte, err error) {
	client := &http.Client{}
	var requestReader io.Reader
	if body != nil {
		requestByte, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestReader = bytes.NewReader(requestByte)
	}
	req, err := http.NewRequest(method, url, requestReader)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}

	if query != nil {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if res.StatusCode >= 300 {
		return nil, errors.New(string(b))
	}
	return b, nil
}

func httpRequestByPage(method string, url string, query map[string]string, body *struct{}, header map[string]string, page int) (data []byte, err error) {
	query["page"] = strconv.Itoa(page)
	return httpRequest(method, url, query, body, header)
}

func getVars() (l *Varlist, err error) {
	l = &Varlist{}
	i := 1
	for {
		data, err := httpRequestByPage(
			"GET",
			fmt.Sprintf(gitlabURL, projectID),
			map[string]string{},
			nil,
			map[string]string{"PRIVATE-TOKEN": gitlabToken, "Content-Type": "application/json"},
			i,
		)
		if err != nil {
			return nil, err
		}
		tmp := &Varlist{}
		if err = json.Unmarshal(data, &tmp); err != nil {
			return nil, err
		}
		if len(*tmp) != 0 {
			*l = append(*l, *tmp...)
			i++
		} else {
			return l, nil
		}
	}
}

func createVars(v *GitlabVar) error {
	_, err := httpRequest(
		"POST",
		fmt.Sprintf(gitlabURL, projectID),
		map[string]string{"filter[environment_scope]": v.EnvironmentScope},
		v,
		map[string]string{"PRIVATE-TOKEN": gitlabToken, "Content-Type": "application/json"},
	)
	if err != nil {
		return err
	}
	return nil
}

func updateVars(v *GitlabVar) error {
	_, err := httpRequest(
		"PUT",
		fmt.Sprintf(gitlabURL+"/"+v.Key, projectID),
		map[string]string{"filter[environment_scope]": v.EnvironmentScope},
		v,
		map[string]string{"PRIVATE-TOKEN": gitlabToken, "Content-Type": "application/json"},
	)
	if err != nil {
		return err
	}
	return nil
}

func deleteVars(v *GitlabVar) error {
	_, err := httpRequest(
		"DELETE",
		fmt.Sprintf(gitlabURL+"/"+v.Key, projectID),
		map[string]string{"filter[environment_scope]": v.EnvironmentScope},
		v,
		map[string]string{"PRIVATE-TOKEN": gitlabToken, "Content-Type": "application/json"},
	)
	if err != nil {
		return err
	}
	return nil
}
