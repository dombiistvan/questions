package main

import (
	"net/http"
	"time"
	"encoding/json"
	"io/ioutil"
	"errors"
)

type handler func(w http.ResponseWriter, r *http.Request)

func GetOnly(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			h(w, r)
			return
		}
		http.Error(w, getErrorByStatusCode(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func PostOnly(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			h(w, r)
			return
		}
		http.Error(w, getErrorByStatusCode(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func AuthUser(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var at string = r.Header.Get("access-token")
		var user User

		if len(at) != 64 || user.LoadByAccessToken(at) != nil || user.TokenExpiration.Unix() < time.Now().Unix() {
			http.Error(w, getErrorByStatusCode(http.StatusForbidden), http.StatusForbidden)
			return
		}

		h(w, r)
	}
}

func getErrorByStatusCode(statusCode int) string {
	switch statusCode {
	case http.StatusMethodNotAllowed:
		return "method is not allowed"
	case http.StatusConflict:
		return "entity has already been found"
		break
	case http.StatusInternalServerError:
		return "internal error"
		break
	case http.StatusExpectationFailed:
		return "expectations failed"
		break
	case http.StatusForbidden:
		return "access forbidden"
		break
	}

	return "error message has not been specified"
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	var err error
	var byteResp []byte
	byteResp,err = json.Marshal(data)
	if err != nil{
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(byteResp)
}

func getJsonData(r *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("missing request body")
	}

	var data interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, errors.New("bad request format")
	}

	return data, nil
}

func getAuthUser(r *http.Request) (User,error){
	var at string = r.Header.Get("access-token")
	var user User
	var err error

	err = user.LoadByAccessToken(at)

	return user,err
}
