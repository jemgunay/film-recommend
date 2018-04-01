package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Respond to HTTP request clean up request body.
func httpRespond(w http.ResponseWriter, r *http.Request, response interface{}, status int) {
	w.WriteHeader(status)

	// write response
	if _, err := fmt.Fprintf(w, "%v", response); err != nil {
		fmt.Printf("error writing response, error:[%v] response:[%v (%v)]", err, response, status)
	}

	r.Body.Close()
}

// Get & validate a parameter from the URL query.
func getURLParams(r *http.Request) (params map[string]string) {
	params = make(map[string]string)

	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	return
}

// Get & validate a parameter from the request body/data.
func getDataParams(r *http.Request) (params map[string]string, err error) {
	params = make(map[string]string)

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return params, fmt.Errorf("error reading request body, query=[%v]", r.URL.RawQuery)
	}

	// collect all parameter key:value pairs
	paramsPairs, err := url.ParseQuery(string(requestBody))
	if err != nil {
		return params, fmt.Errorf("error parsing body to URL query, data=[%v] error=[%v]", string(requestBody), err)
	}

	// put into map
	for k, v := range paramsPairs {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	return
}

// Replace variables in HTML templates with corresponding values in TemplateData.
func completeTemplate(filePath string, data interface{}) (result template.HTML) {
	filePath = rootPath + filePath

	// load HTML template from disk
	htmlTemplate, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// parse HTML template & register template functions
	templateParsed, err := template.New("t").Funcs(template.FuncMap{
		"formatEpoch": func(epoch int64) string {
			t := time.Unix(epoch, 0)
			return t.Format("02/01/2006 [15:04]")
		},
		"toTitleCase": func(text string) string {
			return strings.Title(text)
		},
	}).Parse(string(htmlTemplate))
	if err != nil {
		fmt.Println(err)
		return
	}

	// perform template variable replacement
	buffer := new(bytes.Buffer)
	if err = templateParsed.Execute(buffer, data); err != nil {
		fmt.Println(err)
		return
	}

	return template.HTML(buffer.String())
}

// Convert a target object into a JSON string.
func toJSON(target interface{}) (JSON string, err error) {
	jsonResponse, err := json.MarshalIndent(target, "", "\t")
	if err != nil {
		return
	}
	return string(jsonResponse), nil
}
