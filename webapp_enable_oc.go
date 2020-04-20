package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/spanner"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

var (
	projectID  = os.Getenv("PROJECT_ID")
	instanceID = os.Getenv("INSTANCE_ID")
	databaseID = os.Getenv("DATABASE_ID")
	dbPath     = fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)
	client     *spanner.Client
)

type user struct {
	ID        string `spanner:"id"`
	Firstname string `spanner:"first_name"`
	Lastname  string `spanner:"last_name"`
	Email     string `spanner:"email"`
}

func listUsers() []byte {
	ctx := context.Background()
	stmt := spanner.Statement{SQL: `
    SELECT id, first_name, last_name, email FROM users
    `}
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	users := []user{}

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var u user

		if err := row.ToStruct(&u); err != nil {
			log.Fatal(err)
		}

		users = append(users, u)
	}
	// fmt.Println("users", users)

	// Convert to json string format.
	byteArray, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}
	return byteArray
}

func writeUser(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		fmt.Fprintf(w, "Failed to parse the form.: %v", err)
		http.Error(w, "Failed to parse the form.", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Post form: %v\n", req.PostForm)
	id, err := uuid.NewUUID()
	if err != nil {
		fmt.Fprintf(w, "Failed to generate UUID: %v", err)
		http.Error(w, "Failed to generate UUID.", http.StatusBadRequest)
		return
	}
	firstname := req.FormValue("firstname")
	lastname := req.FormValue("lastname")
	email := req.FormValue("email")
	fmt.Fprintf(w, "ID = %s\n", id.String())
	fmt.Fprintf(w, "First name = %s\n", firstname)
	fmt.Fprintf(w, "Last name = %s\n", lastname)
	fmt.Fprintf(w, "Email = %s\n", email)

	columns := []string{"id", "email", "first_name", "last_name"}
	m := []*spanner.Mutation{
		spanner.InsertOrUpdate("users", columns, []interface{}{id.String(), email, firstname, lastname}),
	}
	_, err = client.Apply(context.Background(), m)
	if err != nil {
		fmt.Fprintf(w, "Failed to write data: %v", err)
		http.Error(w, "Failed to write data.", http.StatusBadRequest)
		return
	}
}

func users(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Write(listUsers())
	case "POST":
		writeUser(w, req)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func main() {
	port := flag.Int("port", 8080, "port to run a testing server.")
	flag.Parse()

	// Enable all default views.
	spanner.EnableStatViews()

	// Set up the stackdriver exporter.
	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID:         projectID,
		ReportingInterval: 60 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create the StackDriver exporter: %v", err)
	}
	defer sd.Flush()

	sd.StartMetricsExporter()
	defer sd.StopMetricsExporter()

	// Intialize a Spanner client.
	ctx := context.Background()
	client, err = spanner.NewClient(ctx, dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Start the server.
	http.HandleFunc("/users", users)
	fmt.Printf("Starting server for testing metrics. Listening to port %d.\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal(err)
	}
}
