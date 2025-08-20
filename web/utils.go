package web

import "net/http"

func HxRefresh(w http.ResponseWriter) error {
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
	return nil
}

func HxRedirect(w http.ResponseWriter, target string) error {
	w.Header().Set("HX-Redirect", target)
	w.WriteHeader(http.StatusOK)
	return nil
}

func HxLocation(w http.ResponseWriter, target string) error {
	w.Header().Set("HX-Location", target)
	w.WriteHeader(http.StatusOK)
	return nil
}

func Redirect(w http.ResponseWriter, target string) error {
	w.Header().Set("Location", target)
	w.WriteHeader(http.StatusFound)
	return nil
}
