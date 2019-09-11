package main

import (
	"net/http"
	//"net/http/httputil"
	//"net/url"
	//"strings"
	//log "github.com/sirupsen/logrus"
)

func redirect(w http.ResponseWriter, req *http.Request) {
	//host := strings.Split(req.Host, ":")
	//host[1] = "443"

	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

//rpurl, _ := url.Parse("https://remoteAddr:433")
//proxy := httputil.NewSingleHostReverseProxy(rpurl)
//if err := http.ListenAndServe(":"+"8080", proxy); err != nil {
//log.Fatal("ListenAndServe:", err)
//}
