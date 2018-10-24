package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	appName = "lellaChatterella"
	version = "0.0.3.1"
)

// Handles when a client requests info on the last
// version available of the client apps (android or desktop).
// We simply reply by sending a JSON string containing the current
// version of the app
func handleUpdateInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Update info requested")
	w.Header().Set("Content-Type", "application/json")
	versionFile, err := ioutil.ReadFile("/var/www/chatterella/version.json")
	if err != nil {
		log.Printf("Error while reading the version file: %s\n", err)
		replyInternalServerError(w, []byte(`{"error": "something unexpected occurred"}`))
		return
	}
	sendReply(w, versionFile)
}

// Handles download of the client apps. Client needs to send the platform
// for which it wants to download the app and if it's recognized we send it
// the app back as an apk (Android) or tar.gz (Desktop)
func handleDownload(w http.ResponseWriter, r *http.Request) {
	log.Println("Download requested")
	platform, ok := r.URL.Query()["platform"]
	if !ok || (platform[0] != "desktop" && platform[0] != "android") {
		w.Header().Set("Content-Type", "application/json")
		replyBadRequest(w, []byte(`{"error": "your request was not understood"}`))
		return
	}

	http.ServeFile(w, r, getClientAppPath(platform[0]))
}

func getClientAppPath(platform string) string {
	if platform == "android" {
		return fmt.Sprintf("/var/www/chatterella/android/%s-%s.apk", appName, version)
	} else {
		return fmt.Sprintf("/var/www/chatterella/desktop/%s-%s.tar.gz", appName, version)
	}
}
