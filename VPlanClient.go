package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
)

func isCreateOrWrite(op fsnotify.Op) bool {
	return op&fsnotify.Write == fsnotify.Write || op&fsnotify.Create == fsnotify.Create
}

func main() {
	var url = flag.String("url", "", "The URL to post the VPlan data to")
	var folder = flag.String("folder", "", "The folder to watch")
	var auth = flag.String("auth", "", "The authentication secret")

	flag.Parse()
	log.Println(*url, *folder)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if isCreateOrWrite(event.Op) {
					log.Println("modified file: ", event.Name)
					b, err := ioutil.ReadFile(event.Name)
					if err != nil {
						log.Fatal(err)
					}

					postVPlan(*url, toUtf8(b), *auth)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(*folder)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func toUtf8(buffer []byte) string {
	buf := make([]rune, len(buffer))
	for i, b := range buffer {
		buf[i] = rune(b)
	}
	return string(buf)
}

func postVPlan(url, data, auth string) {
	if len(data) < 20 {
		return
	}

	values := map[string]string{"vplan": data}

	jsonValue, _ := json.Marshal(values)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	if resp.StatusCode != 200 {
		log.Fatal("Vertretungsplan konnte nicht hochgeladen werden!")
		return
	}

	defer resp.Body.Close()
}
