package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"net/http"
	"reflect"
)

type Args map[string]interface{}

func DoQuery(url, query string, args Args) (r Result, err error) {
	requestBytes := bytes.Buffer{}
	request := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: args,
	}
	if err = json.NewEncoder(&requestBytes).Encode(request); err != nil {
		return
	}
	rqst, err := http.NewRequest(http.MethodPost, url, &requestBytes)
	if err != nil {
		return
	}
	rqst.Header.Set("Content-Type", "application/json; charset=utf-8")
	rqst.Header.Set("Accept", "application/json; charset=utf-8")
	res, err := http.DefaultClient.Do(rqst)
	if err != nil {
		return
	}
	defer res.Body.Close()
	bf := bytes.Buffer{}
	if _, err = io.Copy(&bf, res.Body); err != nil {
		return
	}
	x := map[string]interface{}{}
	if err = json.Unmarshal(bf.Bytes(), &x); err != nil {
		return
	}
	if l, ok := x["errors"]; ok {
		fmt.Println(l)
		q := l.([]interface{})[0].(map[string]interface{})
		err = xerrors.Errorf("GraphQL error:", q["message"].(string))
		return
	}
	r = Result(reflect.ValueOf(x))
	return
}

func IfQuery(url, query string, args Args, f func(q Result) interface{}) interface{} {
	q, err := DoQuery(url, query, args)
	if err != nil { return err }
	return f(q)
}
