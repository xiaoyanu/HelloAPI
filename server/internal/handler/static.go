package handler

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"
)

//go:embed static/Logo.psd
var logo []byte

func serveLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/vnd.adobe.photoshop")
	http.ServeContent(w, r, "Logo.psd", time.Time{}, bytes.NewReader(logo))
}
