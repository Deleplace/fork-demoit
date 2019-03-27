/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/dgageot/demoit/files"
	"github.com/dgageot/demoit/flags"
	"github.com/dgageot/demoit/templates"
	"github.com/gorilla/mux"
	"github.com/jaschaephraim/lrserver"
)

// Page describes a page of the demo.
type Page struct {
	WorkingDir  string
	HTML        []byte
	URL         string
	PrevURL     string
	NextURL     string
	CurrentStep int
	StepCount   int
	DevMode     bool
}

// LiveReloadPort is the port used for live-reload of page contents, on change.
// Default is 35729.
// Can be set to an alternative value, before rendering.
var LiveReloadPort uint16 = lrserver.DefaultPort

// Step renders a given page.
func Step(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	steps, err := readSteps(files.Root)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read steps: %v", err), http.StatusInternalServerError)
		return
	}

	id := 0
	if vars["id"] != "" {
		id, err = strconv.Atoi(vars["id"])
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	step := steps[id]
	pageTemplate, err := template.New("page").
		Funcs(template.FuncMap{"lrport": func() uint16 { return LiveReloadPort }}).
		Parse(templates.Index(step.HTML))
	if err != nil {
		http.Error(w, "Unable to parse page", http.StatusInternalServerError)
		return
	}

	var html bytes.Buffer
	err = pageTemplate.Execute(&html, step)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to render page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html.WriteTo(w)
}

func readSteps(folder string) ([]Page, error) {
	var steps []Page

	content, err := ioutil.ReadFile(filepath.Join(folder, "demoit.html"))
	if err != nil {
		return nil, err
	}

	parts := bytes.Split(content, []byte("---"))
	for i, part := range parts {
		var url string
		if i == 0 {
			url = "/"
		} else {
			url = fmt.Sprintf("/%d", i)
		}

		steps = append(steps, Page{
			WorkingDir:  folder,
			HTML:        part,
			DevMode:     *flags.DevMode,
			CurrentStep: i,
			URL:         url,
		})
	}

	for i := range steps {
		steps[i].StepCount = len(steps) - 1
		if i > 0 {
			steps[i].PrevURL = steps[i-1].URL
		}
		if i < len(steps)-1 {
			steps[i].NextURL = steps[i+1].URL
		}
	}

	return steps, nil
}

// VerifyStepsFile returns non-nil error if it can't read demoit.html
func VerifyStepsFile() error {
	_, err := readSteps(files.Root)
	return err
}
