package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	base    uint64 = 62
	charSet        = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	collection *mongo.Collection
	last       uint64
	lastC      int
	DOMAIN     string
)

func init() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))

	if err != nil {
		log.Fatal("Could not connect to MongoDB")
	}

	database := client.Database("shorty")
	collection = database.Collection("url")
}

func main() {
	http.HandleFunc("/", handler)

	DOMAIN = os.Getenv("DOMAIN")
	var PORT = os.Getenv("PORT") // Get the PORT from the environment variables

	if PORT == "" {
		PORT = "8080"
	}

	if DOMAIN == "" {
		DOMAIN = "http://localhost:" + PORT + "/"
	}

	http.ListenAndServe(":"+PORT, Logd(http.DefaultServeMux))
}

func Logd(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		http.Header.Add(w.Header(), "content-type", "json/application")

		var input URL
		jErr := json.NewDecoder(r.Body).Decode(&input)

		if jErr != nil && input.URI == "" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "URI is required"})
			return
		}
		var url URL
		e := collection.FindOne(context.Background(), URL{URI: input.URI}).Decode(&url)
		if e == nil {
			json.NewEncoder(w).Encode(map[string]string{"uri": url.URI, "shorty": DOMAIN + url.Shorty})
			return
		}
		shorty := timeBase62withCount()

		_, err := collection.InsertOne(context.Background(), URL{URI: input.URI, Shorty: shorty, Count: 0})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"uri": input.URI, "shorty": DOMAIN + shorty})
		return
	}

	if r.Method == "GET" {
		shorty := r.URL.Path[1:]
		if shorty == "" {
			http.Error(w, "Bad Request", 400)
			return
		}
		var url URL
		err := collection.FindOne(context.Background(), URL{Shorty: shorty}).Decode(&url)

		if err != nil {
			http.Error(w, "Not Found", 404)
			return
		}
		url.Count++
		url.Logs = append(url.Logs, Log{IP: r.RemoteAddr, Refer: r.Referer(), Time: primitive.NewDateTimeFromTime(time.Now())})
		_, _ = collection.ReplaceOne(context.Background(), URL{Shorty: shorty}, url)
		http.Redirect(w, r, url.URI, http.StatusTemporaryRedirect)
		return
	}

}

func timeBase62withCount() string {
	var now = uint64(time.Now().Unix())
	if now == last {
		lastC++
	} else {
		last = now
		lastC = 0
	}

	var encoded string
	for now > 0 {
		encoded = string(charSet[now%base]) + encoded
		now = now / base
	}
	return encoded + string(charSet[lastC])
}

type URL struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	URI    string             `bson:"uri,omitempty"`
	Shorty string             `bson:"shorty,omitempty"`
	Count  int                `bson:"count,omitempty"`
	Logs   []Log              `bson:"logs,omitempty"`
}

type Log struct {
	IP    string             `bson:"ip,omitempty"`
	Refer string             `bson:"refer,omitempty"`
	Time  primitive.DateTime `bson:"time,omitempty"`
}
