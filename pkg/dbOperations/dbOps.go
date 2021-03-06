package dbOperations

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
)
var URL string = os.Getenv("DB_URL")


type Project struct {
	Name string
	URL string
	SlackURL string
}
type ProjectLog struct {
	Name string
	URL string
}
func check(e error) int8 {
	if e != nil {
		return -1
	}
	return 0
}

func CheckIFExists(name string) int8 {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	build := client.Database("Projects").Collection("Builds")
	var m Project
	e := build.FindOne(ctx,bson.M{
		"name": name,
	}).Decode(&m)
	if e == nil {
		return 1
	} else {
		return 0
	}
}
func InsertProject(name string, URLProj  string, slackURL string) int8 {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	build := client.Database("Projects").Collection("Builds")
	var m Project
	e := build.FindOne(ctx,bson.M{
		"name": name,
	}).Decode(&m)
	if e == nil {
		return -1
	} else {
		_id, e := build.InsertOne(ctx, Project{name, URLProj, slackURL})
		if check(e) == -1 {
			return -1
		}
		fmt.Println("Inserted id :", _id)
	}
	return 0
}
func UpdateProject(name string, URLProj  string, slackURL string) int8 {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	build := client.Database("Projects").Collection("Builds")
	if CheckIFExists(name) == 0 {
		return -1
	}
	filter := bson.D{{"name", name}}
	update := bson.D{
		{"$set", bson.D{
			{"url", URLProj},
			{"slackurl", slackURL},
		}},
	}
	_, err = build.UpdateOne(ctx, filter, update)
	if err != nil {
		return -1
	} else {
		return 0
	}
}
func ReadProject(name string) (Project) {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	build := client.Database("Projects").Collection("Builds")
	var m Project
	build.FindOne(ctx,bson.M{
		"name": name,
	}).Decode(&m)
	return m
}
func DeleteProjects(name string) int8 {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	build := client.Database("Projects").Collection("Builds")
	delRes, err := build.DeleteMany(ctx, bson.M{
		"name": name,
	})
	if delRes.DeletedCount == 0 {
		return 2
	} else if err == nil {
		return 1
	} else {
		return 0
	}
}
func AllProjects() []Project {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	build := client.Database("Projects").Collection("Builds")
	results := make([]Project,0,200)
	cur, e := build.Find(ctx,bson.D{})
	if e != nil {
		return []Project{}
	}
	for cur.Next(context.TODO()) {
		var elem Project
		err := cur.Decode(&elem)
		if err != nil {
			return []Project{}
		}
		results = append(results, elem)
	}
	if err := cur.Err(); err != nil {
		return []Project{}
	}
	cur.Close(context.TODO())
	return results
}
func InsertLogs(name string, URLProj string) int8 {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	Logs := client.Database("Projects").Collection("Logs")
	var m ProjectLog
	e := Logs.FindOne(ctx,bson.M{
		"name": name,
	}).Decode(&m)
	if e == nil {
		filter := bson.D{{"name", name}}
		update := bson.D{
			{"$set", bson.D{
				{"url", URLProj},
			}},
		}
		_, err = Logs.UpdateOne(ctx, filter, update)
	} else {
		_id, e := Logs.InsertOne(ctx, ProjectLog{name, URLProj})
		if check(e) == -1 {
			return -1
		}
		fmt.Println("Inserted id :", _id)
	}
	return 0
}
func GetLogURL(name string) string {
	clientOptions := options.Client().ApplyURI(URL)
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	check(err)
	defer client.Disconnect(context.TODO())
	build := client.Database("Projects").Collection("Logs")
	var m ProjectLog
	build.FindOne(ctx,bson.M{
		"name": name,
	}).Decode(&m)
	return m.URL
}
func GetLogs(name string) string {
	URL := GetLogURL(name)
	if URL == "" {
		return ""
	}
	resp, _ := http.Get(URL)
	var m map[string] interface{}
	e := json.NewDecoder(resp.Body).Decode(&m)
	if e != nil {
		log.Fatalln(e)
	}
	keys := make([]string,len(m))
	count :=0
	for v,_ := range m {
		keys[count] = v
		count++
	}
	sort.Strings(keys)
	count = 0
	CompleteLogs := "LOGS ["+name+"]"
	for _,v := range keys {
		value := m[v].(map[string] interface{})
		logs := value["log"].(string)
		reg, err := regexp.Compile("[^a-zA-Z0-9/.]+")
		if err != nil {
			log.Fatal(err)
		}
		processedString := reg.ReplaceAllString(logs, " ")
		count++
		if processedString[0] != ' '{
			CompleteLogs+=fmt.Sprintf("\n[%d] %s", count, processedString)
		}
	}
	return CompleteLogs
}