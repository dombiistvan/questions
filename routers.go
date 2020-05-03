package main

import (
	"net/http"
	"encoding/json"
	"database/sql"
)

/**
create user by eamil, password and name mandatory fields.
also check if the email is in the database already
 */
func CreateUser(w http.ResponseWriter, r *http.Request) {
	jsonData, err := getJsonData(r)

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	var user User
	email, eok := jsonData.(map[string]interface{})["email"]
	name, nok := jsonData.(map[string]interface{})["name"]
	password, pok := jsonData.(map[string]interface{})["password"]

	if !eok || !nok || !pok {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}
	err = user.LoadByEmail(email.(string))

	if err == nil {
		http.Error(w, getErrorByStatusCode(http.StatusConflict), http.StatusConflict)
		return
	} else if err != nil && err != sql.ErrNoRows {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user = NewUser(name.(string), email.(string), password.(string))
	err = user.Save()

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/**
login user by email and password, returns access token must be sent in header in the following
 */
func LoginUser(w http.ResponseWriter, r *http.Request) {
	jsonData, err := getJsonData(r)

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	var user User
	email, eok := jsonData.(map[string]interface{})["email"]
	password, pok := jsonData.(map[string]interface{})["password"]

	if !eok ||  !pok {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}
	err = user.LoadByEmailPass(email.(string),password.(string))

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	err = user.GenerateAccessToken()

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var mapResponse map[string]interface{} = make(map[string]interface{})
	var byteResponse []byte
	mapResponse["token"] = user.AccessToken
	mapResponse["expiration"] = user.TokenExpiration.Format("2006-01-02 15:04:05")
	byteResponse, err = json.Marshal(mapResponse)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(byteResponse)

}

/**
returns top 5 users by answer count to authenticated user request
 */
func UsersTopFive(w http.ResponseWriter, r *http.Request) {
	var topUsers []User
	var user User

	_,err := getAuthUser(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	topUsers,err = user.GetTopUsers(5)

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonResponse(w,topUsers)
}

/**
save question to db to authenticated user request
 */
func PostQuestion(w http.ResponseWriter, r *http.Request) {
	var user User
	var qstring interface{}
	var question Question

	jsonData, err := getJsonData(r)

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	user,err = getAuthUser(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	qstring, qok := jsonData.(map[string]interface{})["question"]
	if !qok {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	question = NewQuestion(qstring.(string),user)
	err = question.Save()

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var response map[string]interface{} = make(map[string]interface{})
	response["id"] = question.Id

	jsonResponse(w,response)
}
/**
 post answer to question to authenticated user request
 */
func PostAnswer(w http.ResponseWriter, r *http.Request) {
	var astring interface{}
	var qid interface{}

	var question Question
	var answer Answer
	var user User

	user,err := getAuthUser(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonData, err := getJsonData(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	astring, aok := jsonData.(map[string]interface{})["answer"]
	if !aok {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	qid,qok := jsonData.(map[string]interface{})["question_id"]
	if !qok || question.Load(int64(qid.(float64))) != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	answer = NewAnswer(astring.(string),user,question)
	err = answer.Save()

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var response map[string]interface{} = make(map[string]interface{})
	response["id"] = answer.Id

	jsonResponse(w,response)
}
/**
 list questions ordered by answer count to authenticated user request
 */
func QuestionsByAnswer(w http.ResponseWriter, r *http.Request) {
	var pageNum interface {}
	var page int = 1

	var questions []Question
	var question Question

	_,err := getAuthUser(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonData, err := getJsonData(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	pageNum,pok := jsonData.(map[string]interface{})["page"]
	if pok && int64(pageNum.(float64))>1{
		page = int(pageNum.(float64))
	}

	questions,err = question.GetByAnswersCount(page,5,true)

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonResponse(w,questions)
}

/**
 list questions ordered by answers rate to authenticated user request
 */
func QuestionAnswersByRate(w http.ResponseWriter, r *http.Request) {
	var qid interface{}
	var pageNum interface {}
	var page int = 1

	var question Question
	var answers []Answer

	_,err := getAuthUser(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonData, err := getJsonData(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	qid, qok := jsonData.(map[string]interface{})["question_id"]
	if !qok || question.Load(int64(qid.(float64))) != nil{
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	pageNum,pok := jsonData.(map[string]interface{})["page"]
	if pok && int64(pageNum.(float64))>1{
		page = int(pageNum.(float64))
	}

	answers,err = question.GetAnswersByRate(page,5,true)

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonResponse(w,answers)
}
/**
 rate answer to authenticated user request
 */
func RateAnswer(w http.ResponseWriter, r *http.Request) {
	var aid interface{}

	var answer Answer
	var answerRate AnswerRate
	var user User

	user,err := getAuthUser(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonData, err := getJsonData(r)
	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	aid, aok := jsonData.(map[string]interface{})["answer_id"]
	if !aok || answer.Load(int64(aid.(float64))) != nil{
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	_,rok := jsonData.(map[string]interface{})["rate"]
	if !rok {
		http.Error(w, getErrorByStatusCode(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	err = answerRate.LoadByAnswerAndUser(answer,user)

	if err == nil {
		http.Error(w, getErrorByStatusCode(http.StatusConflict), http.StatusConflict)
		return
	} else if err != sql.ErrNoRows{
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	answerRate = NewAnswerRate(1,user,answer)
	err = answerRate.Save()

	if err != nil {
		http.Error(w, getErrorByStatusCode(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
