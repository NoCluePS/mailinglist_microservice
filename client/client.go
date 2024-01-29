package main

import (
	"context"
	"log"
	pb "mailinglist/proto"
	"time"

	"github.com/alexflint/go-arg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func logResponses(res *pb.EmailResponse, err error) {
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if res.EmailEntry == nil {
		log.Printf("No entry found\n")
	} else {
		log.Printf("Entry: %v\n", res.EmailEntry)
	}
}

func createEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Printf("Creating email %v\n", addr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.CreateEmail(ctx, &pb.CreateEmailRequest{EmailAddr: addr})
	logResponses(res, err)
	return res.EmailEntry;
}

func getEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Printf("Getting email %v\n", addr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmail(ctx, &pb.GetEmailRequest{EmailAddr: addr})
	logResponses(res, err)
	return res.EmailEntry;
}

func getEmailBatch(client pb.MailingListServiceClient, page int32, count int32) []*pb.EmailEntry {
	log.Println("Getting batch email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmailBatch(ctx, &pb.GetEmailBatchRequest{Page: page, Count: count})
	
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Println("response:")
	for i, entry := range res.EmailEntries {
		log.Printf(" item [%v of %v]: %s\n", i+1, len(res.EmailEntries), entry)
	}

	return res.EmailEntries;
}

func updateEmail(client pb.MailingListServiceClient, entry *pb.EmailEntry) *pb.EmailEntry {
	log.Printf("Updating email %v\n", entry.Email)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.UpdateEmail(ctx, &pb.UpdateEmailRequest{EmailEntry: entry})
	logResponses(res, err)
	return res.EmailEntry;
}

func deleteEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Printf("Updating email %v\n", addr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.DeleteEmail(ctx, &pb.DeleteEmailRequest{EmailAddr: addr})
	logResponses(res, err)
	return res.EmailEntry;
}

var args struct {
	GrpcAddr string `arg:"env:MAILINGLIST_GRPC_ADDR"`
}

func main() {
	arg.MustParse(&args)

	if args.GrpcAddr == "" {
		args.GrpcAddr = ":8001"
	}

	conn, err := grpc.Dial(args.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Error (couldn't connect): %v\n", err)
	}
	defer conn.Close()

	client := pb.NewMailingListServiceClient(conn)

	newEmail := createEmail(client, "9999@999.com")
	newEmail.ConfirmedAt = 10000
	updateEmail(client, newEmail)
	deleteEmail(client, newEmail.Email)
	getEmailBatch(client, 1, 10)
}