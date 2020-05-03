package main

import (
	"net/http"
	"log"
	"fmt"
)

const listenPort = "8080"

func main(){
	http.HandleFunc("/user/create", PostOnly(CreateUser))
	http.HandleFunc("/user/login", PostOnly(LoginUser))
	http.HandleFunc("/user/list/top5", GetOnly(UsersTopFive))

	http.HandleFunc("/question/new", PostOnly(AuthUser(PostQuestion)))
	http.HandleFunc("/question/list/byanswers", GetOnly(AuthUser(QuestionsByAnswer)))

	http.HandleFunc("/answer/new", PostOnly(AuthUser(PostAnswer)))
	http.HandleFunc("/answer/rate", PostOnly(AuthUser(RateAnswer)))
	http.HandleFunc("/answer/list/byrate", GetOnly(AuthUser(QuestionAnswersByRate)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s",listenPort), nil))
}
