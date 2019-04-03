/*
Copyright 2019 Google LLC

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
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// SpeakerNotes provides the speaker notes of the "current" step.
func SpeakerNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] != "" {
		// TODO if a specific step ID is requested, then serve specific step.
	}

	snStateLock.Lock()
	stepHTML := currentStepHTML
	snStateLock.Unlock()

	if stepHTML == nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, "&lt; Speaker Notes: no steps displayed yet &gt;")
		// TODO maybe serve step 0?
		fmt.Fprint(w, speakerNotesTrailer)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(stepHTML)
	fmt.Fprintf(w, "\n\n<script>var currentStepID = %d;</script>", currentStepID)
	fmt.Fprint(w, speakerNotesTrailer)
	// Now serve JS to reveal Speaker Notes
}

func syncSpeakerNotes(stepID int, stepHTML []byte) {
	snStateLock.Lock()
	defer snStateLock.Unlock()

	currentStepID = stepID
	currentStepHTML = stepHTML

	// Disable the normal step interactivity
	stepScript := []byte(`<script src="/js/demoit.js"></script>`)
	if !bytes.Contains(currentStepHTML, stepScript) {
		currentStepHTML = []byte("Expected a very specific script tag for demoit.js, couldn't find it.")
		return
	}
	currentStepHTML = bytes.ReplaceAll(currentStepHTML, stepScript, nil)

	// TODO server push new state to currently open Speaker Notes clients.
	// Or just let them poll.
}

// The currentStepID is the id of the last step viewed.
// The currentStepHTML is the full HTML of the last step viewed.
// This global state is used to open the Speaker Notes in the
// correct step.
var (
	currentStepID   int
	currentStepHTML []byte

	snStateLock sync.Mutex
)

const speakerNotesTrailer = `
<style>
	#speaker-notes {
		display: flex;
	}
	#nav {
		display: none;
	}
</style>
<script>
	// TODO auto-refresh when the main presentation window moves to next step
	window.setTimeout( () => { location.reload(true); }, 1000);
</script>
`
