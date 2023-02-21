package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

type UserDto struct {
	Id        int    `xml: "id"`
	About     string `xml: "about"`
	Age       int    `xml: "age"`
	FirstName string `xml: "first_name"`
	LastName  string `xml: "last_name"`
}

type UsersXMLDto struct {
	Version string    `xml: "version, attr"`
	Users   []UserDto `xml: "row"`
}

var validOrderField = map[string]struct{}{"Id": struct{}{}, "Name": struct{}{}, "Age": struct{}{}}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy := r.URL.Query().Get("order_by")

	if orderField == "" {
		orderField = "Name"
	} else if _, ok := validOrderField[orderField]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		resp := SearchErrorResponse{Error: ErrorBadOrderField}
		json.NewEncoder(w).Encode(resp)
		return
	}

	f, err := os.Open("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	var users UsersXMLDto
	err = xml.Unmarshal(b, &users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

}

func TestFindUsers(t *testing.T) {

}
