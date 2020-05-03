package main

import (
	"database/sql"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"strings"
	"crypto/sha256"
	"html"
	"time"
)

const (
	host     = "localhost"
	port     = 3306
	user     = "root"
	password = ""
	dbName   = "questions"
	encoding = "UTF8"
	salt     = "CwQaBVVCcDrvb2dJ"
)

var DBMap *gorp.DbMap

type User struct {
	Id              int64     `db:"id, primarykey, autoincrement" json:"-"`
	Name            string    `db:"name, size:255" json:"name"`
	Email           string    `db:"email, size:255, notnull" json:"email"`
	Password        string    `db:"password, notnull" json:"-"`
	AccessToken     string    `db:"access_token, size:64" json:"-"`
	TokenExpiration time.Time `db:"token_expiration, size:64" json:"-"`
}

func (u *User) Load(id int64) error {
	err := DBMap.SelectOne(u, "SELECT * FROM user WHERE id = ?", id)

	return err
}

func (u *User) LoadByEmailPass(email, password string) error {
	err := DBMap.SelectOne(u, "SELECT * FROM user WHERE email = ? AND password = ?", html.EscapeString(email), u.GetPasswordHash(password))

	return err
}

func (u *User) LoadByEmail(email string) error {
	err := DBMap.SelectOne(u, "SELECT * FROM user WHERE email = ?", html.EscapeString(email))

	return err
}

func (u *User) LoadByAccessToken(accesstoken string) error {
	err := DBMap.SelectOne(u, "SELECT * FROM user WHERE access_token = ?", html.EscapeString(accesstoken))

	return err
}

func (u User) GetTopUsers(limit int) ([]User, error) {
	var query string
	var err error
	var users []User

	query = fmt.Sprintf("SELECT * FROM user ORDER BY (SELECT COUNT(*) FROM answer WHERE user_id = user.id) DESC LIMIT %d", limit)

	_, err = DBMap.Select(&users, query)

	return users, err
}

func (u *User) GetPasswordHash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join([]string{password, salt}, ":"))))
}

func (u *User) GenerateAccessToken() error {
	u.AccessToken = fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join([]string{u.Email, time.Now().Format("2006-01-02 15:04:05")}, ":"))))
	u.TokenExpiration = time.Now().Add(time.Minute * 10)

	//fmt.Println(u.AccessToken, time.Now().Format("2006-01-02 15:04:05"),u.TokenExpiration.Format("2006-01-02 15:04:05"))
	return u.Save()
}

func (u *User) Save() error {
	var err error
	if u.Id == 0 {
		u.Password = u.GetPasswordHash(u.Password)
		err = DBMap.Insert(u)
	} else {
		_, err = DBMap.Update(u)
	}

	return err
}

func NewUser(name, email, password string) User {
	var u User = User{Name: name, Email: email, Password: password}
	return u
}

type Question struct {
	Id       int64    `db:"id, primarykey, autoincrement" json:"id"`
	Question string   `db:"question, size:255, notnull" json:"question"`
	UserId   int64    `db:"user_id, notnull" json:"-"`
	User     *User    `db:"-" json:"user"`
	Answers  []Answer `db:"-" json:"answers"`
}

func (q Question) GetByAnswersCount(page int, limit int, desc bool) ([]Question, error) {
	var query string
	var err error
	var questions []Question
	var sOrderDir string = "ASC"
	var sLimit string = fmt.Sprintf("%d,%d", (page-1)*limit, limit)

	if desc {
		sOrderDir = "DESC"
	}

	query = fmt.Sprintf("SELECT * FROM question ORDER BY (SELECT COUNT(*) FROM answer WHERE question_id = question.id) %s LIMIT %s", sOrderDir, sLimit)

	_, err = DBMap.Select(&questions, query)

	for k, _ := range questions {
		questions[k].AddUserData()
		questions[k].AddAnswersData()
	}

	return questions, err
}

func (q Question) GetAnswersByRate(page int, limit int, desc bool) ([]Answer, error) {
	var query string
	var err error
	var answers []Answer
	var sOrderDir string = "ASC"
	var sLimit string = fmt.Sprintf("%d,%d", (page-1)*limit, limit)

	if desc {
		sOrderDir = "DESC"
	}

	query = fmt.Sprintf("SELECT * FROM answer WHERE question_id = ? ORDER BY (SELECT COUNT(*) FROM answer_rate WHERE answer_id = answer.id) %s LIMIT %s", sOrderDir, sLimit)

	_, err = DBMap.Select(&answers, query, q.Id)

	for k, _ := range answers {
		answers[k].AddUserData()
	}

	return answers, err
}

func (q *Question) Save() error {
	var err error
	if q.Id == 0 {
		err = DBMap.Insert(q)
	} else {
		_, err = DBMap.Update(q)
	}

	return err
}

func (q *Question) Load(id int64) error {
	err := DBMap.SelectOne(q, "SELECT * FROM question WHERE id = ?", id)

	return err
}

func (q *Question) AddUserData() error {
	var u User

	err := u.Load(q.UserId)

	if err != nil {
		return err
	}

	q.User = &u

	return nil
}

func (q *Question) AddAnswersData() error {
	_, err := DBMap.Select(&q.Answers, "SELECT * FROM answer WHERE question_id = ? ORDER BY id DESC", q.Id)

	return err
}

func NewQuestion(question string, user User) Question {
	var q Question = Question{Question: question, UserId: user.Id}
	return q
}

type Answer struct {
	Id         int64  `db:"id, primarykey, autoincrement"`
	Answer     string `db:"answer, notnull" json:"answer"`
	QuestionId int64  `db:"question_id, notnull" json:"-"`
	UserId     int64  `db:"user_id, notnull" json:"-"`
	User       *User  `db:"-" json:"user"`
}

func (a *Answer) Save() error {
	var err error
	if a.Id == 0 {
		err = DBMap.Insert(a)
	} else {
		_, err = DBMap.Update(a)
	}

	return err
}

func (a *Answer) Load(id int64) error {
	err := DBMap.SelectOne(a, "SELECT * FROM answer WHERE id = ?", id)

	return err
}

func (a *Answer) AddUserData() error {
	var u User

	err := u.Load(a.UserId)

	if err != nil {
		return err
	}

	a.User = &u

	return nil
}

func NewAnswer(answer string, user User, question Question) Answer {
	var a Answer = Answer{Answer: answer, QuestionId: question.Id, UserId: user.Id}
	return a
}

type AnswerRate struct {
	Id         int64 `db:"id, primarykey, autoincrement" json:"-"`
	UserId     int64 `db:"user_id, notnull" json:"-"`
	AnswerId   int64 `db:"answer_id, notnull" json:"-"`
	QuestionId int64 `db:"question_id, notnull" json:"-"`
	Rate       int64 `db:"rate, notnull" json:"rate"`
	User       *User `db:"-" json:"user"`
}

func (ar *AnswerRate) AddUserData() error {
	var u User

	err := u.Load(ar.UserId)

	if err != nil {
		return err
	}

	ar.User = &u

	return nil
}

func (ar *AnswerRate) LoadByAnswerAndUser(answer Answer, user User) error {
	err := DBMap.SelectOne(ar, "SELECT * FROM answer_rate WHERE answer_id = ? AND user_id = ?", answer.Id, user.Id)

	return err
}

func (ar *AnswerRate) Save() error {
	var err error
	if ar.Id == 0 {
		err = DBMap.Insert(ar)
	} else {
		_, err = DBMap.Update(ar)
	}

	return err
}

func NewAnswerRate(rate int64, user User, answer Answer) AnswerRate {
	var ar AnswerRate = AnswerRate{UserId: user.Id, AnswerId: answer.Id, QuestionId: answer.QuestionId, Rate: rate}
	return ar
}

func init() {
	estabilishConnection(true)
}

func getSqlinfo(withDb bool) string {
	if withDb {
		return fmt.Sprintf("%s:%s@/%s?parseTime=true", user, password, dbName)
	}

	return fmt.Sprintf("%s:%s@/?parseTime=true", user, password, )
}

func estabilishConnection(build bool) {
	db, err := sql.Open("mysql", getSqlinfo(true))
	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		if build && strings.Contains(err.Error(), "Unknown database") {
			fmt.Println("no database, build database")
			db, err := sql.Open("mysql", getSqlinfo(false))
			if err != nil {
				panic(err)
			}

			query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s;", dbName, "utf8", "utf8_unicode_ci")
			_, err = db.Exec(query)
			if err != nil {
				panic(err)
			}
			estabilishConnection(false)

			return
		} else {
			panic(err)
		}
	}

	// construct a gorp DbMap
	DBMap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: fmt.Sprintf("%s", encoding)}}

	// add a table, setting the table name to 'posts' and
	// specifying that the Id property is an auto incrementing PK
	DBMap.AddTableWithName(Question{}, "question").SetKeys(true, "id")
	DBMap.AddTableWithName(Answer{}, "answer").SetKeys(true, "id")
	DBMap.AddTableWithName(AnswerRate{}, "answer_rate").SetKeys(true, "id")
	DBMap.AddTableWithName(User{}, "user").SetKeys(true, "id")

	err = DBMap.CreateTablesIfNotExists()

	if err != nil {
		panic(err)
	}
}
