package main

import (
	"database/sql"
	"log"
	"mailinglist/api"
	"mailinglist/grpcapi"
	"mailinglist/mdb"
	"sync"

	"github.com/alexflint/go-arg"
)

var args struct {
	DbPath string `arg:"env:MAILINGLIST_DB"`
	BindGrpc string `arg:"env:MAILINGLIST_BIND_GRPC"`
	BindJson string `arg:"env:MAILINGLIST_BIND_JSON"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "list.db";
	}

	if args.BindJson == "" {
		args.BindJson = ":8000";
	}

	if args.BindGrpc == "" {
		args.BindGrpc = ":8001";
	}

	log.Printf("Using database %v\n", args.DbPath)
	db, err := sql.Open("sqlite3", args.DbPath)

	if err != nil { 
		log.Fatal(err);
	}
	defer db.Close();

	mdb.TryCreate(db);
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Println("Starting JSON API...")
		api.Serve(db, args.BindJson)
		wg.Done()

	}()

	wg.Add(1)
	go func() {
		log.Println("Starting gRPC API...")
		grpcapi.Serve(db, args.BindGrpc)
		wg.Done()

	}()
	wg.Wait() 
}