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

package files

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jaschaephraim/lrserver"
	"github.com/radovskyb/watcher"
)

func Watch(root string, port uint16) error {
	lr := lrserver.New(lrserver.DefaultName, port)
	go func() {
		log.Fatal(lr.ListenAndServe())
	}()

	w := watcher.New()
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove, watcher.Rename, watcher.Move)
	if err := w.Ignore(filepath.Join(root, ".git")); err != nil {
		return err
	}
	if err := w.AddRecursive(root); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event)
				lr.Reload(event.Name())
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.Start(time.Millisecond * 100); err != nil {
		return err
	}

	return nil
}

func LiveReloadPort() uint16 {
	port := lrserver.DefaultPort
	for {
		if isPortAvailable(port) {
			break
		}
		// If port is already in use (e.g. another Demoit instance), then
		// choose an alternative port
		log.Println("\tCan't use live reload port", port, "(already in use)")
		port++
		if port == lrserver.DefaultPort+10 {
			log.Fatalln("Couldn't find a live reload port after 10 attempts")
		}
	}
	return port
}

func isPortAvailable(port uint16) bool {
	portstr := strconv.Itoa(int(port))
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("", portstr), time.Second)
	_ = err
	if conn != nil {
		// Connection established means "Port was already in use!"
		conn.Close()
		return false
	}
	return true
}
