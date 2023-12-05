package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var key = "topsecretvalue!"
var s, _ = ioutil.ReadFile("./shedule.json")
var Shed interface{}
var shedule *mongo.Collection

func main() {
	err := bson.UnmarshalExtJSON(s, true, &Shed)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://Ann:Abc123@cluster0.oa3rxor.mongodb.net/?retryWrites=true&w=majority").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	shedule = client.Database("ShedulePI").Collection("group")

	InsertResult, err := shedule.InsertOne(context.TODO(), Shed)
	if err != nil {
		fmt.Print(InsertResult)
		panic(err)
	}

	http.HandleFunc("/get_shedule", Get_shedule)
	http.HandleFunc("/update_shedule", Update_shedule)
	http.ListenAndServe(":8080", nil) //??

}

func Update_shedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
	}

	var body []byte
	r.Body.Read(body)
	opts := options.Replace().SetUpsert(true)
	result, err := shedule.ReplaceOne(context.TODO(), body, opts)
	if err != nil {
		fmt.Print(result)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
}

func Get_shedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
	}

	token, _, err := new(jwt.Parser).ParseUnverified(r.URL.Query().Get("token"), jwt.MapClaims{})
	if err != nil {
		panic(err)
	}

	payload, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		fmt.Println("Can't convert token's claims to standard claims")
	}

	ex := payload["exp"].(float64)
	exp := int64(ex)
	t := time.Now().Unix()

	if exp > t && payload["group"] == "ПИ-б-о-231" && payload["action"] == "get_shedule" {
		var results bson.M
		if err = shedule.FindOne(context.TODO(), bson.M{}).Decode(&results); err != nil {
			panic(err)
		}
		jsonResponse, err := json.Marshal(results)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)

	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Ваш токен недействителен"))
	}

}
