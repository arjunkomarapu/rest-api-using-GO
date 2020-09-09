package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

const connectionString = "mongodb+srv://root:root@cluster0.c4nt3.mongodb.net/test?retryWrites=true&w=majority"
const dbName = "test"
const collName = "todolist"

var collection *mongo.Collection

func CreatePersonEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	var person Person
	json.NewDecoder(req.Body).Decode(&person)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(res).Encode(result)

}

func GetPeopleEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	var people []Person
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err := cursor.Err(); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(people)
}

func GetPersonEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(person)
}

func DelPersonEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	var person Person
	json.NewDecoder(req.Body).Decode(&person)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.DeleteOne(ctx, person)
	json.NewEncoder(res).Encode(result)

}

func UpdatePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var person Person
	filter := bson.M{}
	_ = json.NewDecoder(req.Body).Decode(&person)
	update := bson.D{
		{"$set", bson.D{
			{"firstname", person.Firstname},
			{"lastname", person.Lastname},
		}},
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOneAndUpdate(ctx, filter, update).Decode(&person)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	person.ID = id

	json.NewEncoder(w).Encode(person)

}
func main() {
	fmt.Println("Go-Mongo DB API")
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// client, _ := mongo.Connect(ctx, "mongodb+srv://root:root@cluster0.c4nt3.mongodb.net/test?retryWrites=true&w=majority")
	// client, _ := mongo.Connect(ctx, "mongodb://localhost:27017")
	// router := mux.NewRouter()
	// http.ListenAndServe(":12345", router)
	// Set client options
	clientOptions := options.Client().ApplyURI(connectionString)
	// connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	collection = client.Database(dbName).Collection(collName)
	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", UpdatePersonEndpoint).Methods("PUT")
	router.HandleFunc("/delperson/{id}", DelPersonEndpoint).Methods("DELETE")
	http.ListenAndServe(":12345", router)
	fmt.Println("Collection instance created!")
}
