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
	stepID := currentStepID
	snStateLock.Unlock()

	if stepHTML == nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, "<meta http-equiv=\"refresh\" content=\"1\" />")
		fmt.Fprintln(w, "&lt; Speaker Notes: no steps displayed yet &gt;")
		// TODO maybe serve step 0?
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(stepHTML)
	fmt.Fprintf(w, "\n\n<script>var currentStepID = %d;</script>", stepID)
	fmt.Fprint(w, speakerNotesTrailer)
	// Now serve JS to reveal Speaker Notes
}

// CurrentStepID returns the "current" StepID.
// It is used to reduce FOUC by reloading the Speaker Notes page only
// when the step has actually changed.
func CurrentStepID(w http.ResponseWriter, r *http.Request) {
	snStateLock.Lock()
	fmt.Fprintln(w, currentStepID)
	snStateLock.Unlock()
}

func syncSpeakerNotes(stepID int, stepHTML []byte) {
	snStateLock.Lock()
	defer snStateLock.Unlock()

	currentStepID = stepID
	currentStepHTML = stepHTML

	// Disable the normal step interactivity
	stepScript := []byte(`<script src="/js/demoit.js`)
	if !bytes.Contains(currentStepHTML, stepScript) {
		currentStepHTML = []byte("Expected a very specific script tag for demoit.js, couldn't find it.")
		return
	}
	scritTagLen := len(`<script src="/js/demoit.js?hash=419827a85a"></script>`)
	pos := bytes.Index(currentStepHTML, stepScript)
	if pos+scritTagLen <= len(currentStepHTML) {
		scriptTag := currentStepHTML[pos : pos+scritTagLen]
		currentStepHTML = bytes.ReplaceAll(currentStepHTML, scriptTag, []byte(`<!-- removed demoit.js, no interactivity in speaker notes -->`))
	}

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
	// Auto-refresh when the main presentation window moves to next step
	window.setInterval( () => { 
		fetch('/currentstep').then( (response) => {
			return response.text();
  		}).then( (serverStepID) => {
			if (serverStepID == currentStepID) {
				console.debug("Still " + currentStepID);
			} else {
				console.debug("Transition " + currentStepID + " => " + serverStepID);
				location.reload(true); 
			}
		});
	}, 1000);

	// Keep user line breaks from user speaker notes.
	let sn = document.getElementById("speaker-notes");
	sn.innerHTML = sn.innerHTML.replace(new RegExp('\n', 'g'), ' <br/>\n');
</script>
`
