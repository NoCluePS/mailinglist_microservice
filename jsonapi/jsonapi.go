package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/NoCluePS/mailinglist_api/mdb"
)

func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, servErr := withData()

	if servErr != nil {
		w.WriteHeader(500)
		serverErrJson, err := json.Marshal(&servErr)

		if err != nil {
			log.Println(err)
			return 
		}

		w.Write(serverErrJson)
		return
	}

	dataJson, err := json.Marshal(&data)

	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return;
	}

	w.Write(dataJson)
}

func returnError(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMsg := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMsg, nil
	})
}

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			returnError(w, errors.New("must be POST request"), 400)
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnError(w, err, 400)
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON CreateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			returnError(w, errors.New("must be GET request"), 400)
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetEmailBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			returnError(w, errors.New("must be GET request"), 400)
			return
		}

		queryOptions := mdb.GetEmailBatchQueryParams{}
		fromJson(req.Body, &queryOptions)

		if queryOptions.Count <= 0 || queryOptions.Page <= 0 {
			returnError(w, errors.New("page and count fields are required, must be > 0"), 400)
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmailBatch: %v\n",queryOptions)
			return mdb.GetEmailBatch(db, queryOptions)
		})
	})
}

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			returnError(w, errors.New("must be PUT request"), 400)
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnError(w, err, 400)
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON UpdateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}


func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "DELETE" {
			returnError(w, errors.New("must be DELETE request"), 400)
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.DeleteEmail(db, entry.Email); err != nil {
			returnError(w, err, 400)
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON DeleteEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func Hello() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		returnJson(w, func() (interface{}, error) {
			return struct {
				Message string
			} {
				Message: "Hello, people of the internet!",
			}, nil
		})
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/", Hello())
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/get/batch", GetEmailBatch(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))
	log.Printf("JSON API serve listening on %v\n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatalf("JSON server error: %v\n", err)
	}
}