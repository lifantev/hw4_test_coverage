package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type UserDto struct {
	Id        int    `xml:"id"`
	About     string `xml:"about"`
	Age       int    `xml:"age"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
}

type UsersXMLDto struct {
	Version string    `xml:"version,attr"`
	Users   []UserDto `xml:"row"`
}

var validAccessToken = "abcd"

var validOrderField = map[string]struct{}{"Id": {}, "Name": {}, "Age": {}}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("AccessToken")
	if accessToken != validAccessToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")
	// query := r.URL.Query().Get("query")
	// orderBy := r.URL.Query().Get("order_by")
	orderField := r.URL.Query().Get("order_field")

	if orderField == "" {
		orderField = "Name"
	} else if _, ok := validOrderField[orderField]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		resp := SearchErrorResponse{Error: "ErrorBadOrderField"}
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

	usersResp := make([]User, 0, 3)
	for _, u := range users.Users[:2] {
		usersResp = append(usersResp, User{
			Id:     u.Id,
			Name:   u.FirstName + u.LastName,
			Age:    u.Age,
			About:  u.About,
			Gender: u.Gender,
		})
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(usersResp)
}

type TestCase struct {
	req     SearchRequest
	resp    *SearchResponse
	err     error
	isError bool
	sc      *SearchClient
}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	tcc := []TestCase{
		{
			req:     SearchRequest{OrderField: "Name"},
			resp:    &SearchResponse{},
			err:     nil,
			isError: false,
			sc: &SearchClient{
				AccessToken: validAccessToken,
				URL:         ts.URL,
			},
		},
		{
			req:     SearchRequest{OrderField: "Gender"},
			resp:    nil,
			err:     nil,
			isError: true,
			sc: &SearchClient{
				AccessToken: validAccessToken,
				URL:         ts.URL,
			},
		},
		{
			req:     SearchRequest{OrderField: "Age"},
			resp:    nil,
			err:     nil,
			isError: true,
			sc: &SearchClient{
				AccessToken: "bad_token",
				URL:         ts.URL,
			},
		},
	}

	for i, tc := range tcc {
		sr, err := tc.sc.FindUsers(tc.req)

		if err == nil && tc.isError {
			t.Errorf("[%d] expected error, got nil", i)
		} else if err != nil && !tc.isError {
			t.Errorf("[%d] unexpected error: %#v", i, err)
		} else if err == nil && sr == nil {
			t.Errorf("[%d] wrong return result, expected: %#v, got: %#v", i, tc.resp, sr)
		}
	}

	ts.Close()
}
