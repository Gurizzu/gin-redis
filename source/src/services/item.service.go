package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"practice-redis/src/model"
	"practice-redis/src/util/db"
	"practice-redis/src/util/db/ustring"
	"practice-redis/src/util/enum"
	"time"
)

type ItemService struct {
	collectionName string
	ctx            context.Context
	dbUtil         *db.MongoDbUtil
	redisUtil      *db.RedisUtil
}

func NewItemService() *ItemService {
	o := &ItemService{
		collectionName: enum.MongoCollection_Items.String(),
		ctx:            context.Background(),
	}

	o.dbUtil = db.NewMongoDbUtil("mongodb://localhost:27017", "redis-practice", enum.MongoCollection_Items.String())
	o.redisUtil = db.NewRedisConnection("localhost:6379", "", 0)
	return o
}

func (o *ItemService) Upsert(param model.Item, isUpdate bool) (resp model.Response) {
	dbUtil := o.dbUtil
	_, col := dbUtil.GetCollection()

	paramId := param.Id
	if paramId == "" {
		param.Id = ustring.GenerateID()
	}

	currentTime := time.Now().Unix()
	if param.Created_At == 0 {
		param.Created_At = currentTime
	}
	if param.Updated_At == 0 {
		param.Updated_At = currentTime
	}

	if !isUpdate {
		res, err := col.InsertOne(o.ctx, param)
		if err != nil {
			log.Println(err)
		}
		resp.Data = res.InsertedID
		return
	} else {
		param.Updated_At = time.Now().Unix()
		if updateRes, err := col.UpdateByID(o.ctx, bson.M{"_id": param.Id}, bson.M{"$set": param}); err != nil {
			log.Println(err)
		} else {
			if updateRes.MatchedCount == 0 && updateRes.UpsertedID == "" {
				err = errors.New("data not found. nothing updated")
				log.Println(err)
				return
			}
			resp.Data = updateRes.UpsertedID
		}

	}
	return
}

func (o *ItemService) FindOne(key string, value string) (pointerDecodeTo model.Item, err string) {
	_, col := o.dbUtil.GetCollection()

	filter := bson.D{{key, value}}
	res := col.FindOne(o.ctx, filter)
	if res.Err() != nil {
		log.Println(res.Err(), filter)
		err = "data not found"
		return
	}

	if err := res.Decode(&pointerDecodeTo); err != nil {
		log.Println(err)
	}
	return
}

func (o *ItemService) FindAll(param model.Item_Search) (data []model.Item_View, errMessage string) {

	key := ustring.GetParamAsKey(param)
	fmt.Println(key)

	//cacheData, err := o.redisUtil.Get(key)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()
	cacheData, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("ini chache: %s", cacheData)
	fmt.Printf("ini err: %s", err)

	filter := bson.M{}
	listFilterAnd := []bson.M{}
	param.HandlerFilter(&listFilterAnd)

	if len(listFilterAnd) > 0 {
		filter["$and"] = listFilterAnd
	}

	_, col := o.dbUtil.GetCollection()
	findRes, err := col.Find(o.ctx, filter)
	if err != nil {
		log.Println(err)
		return
	}
	if findRes.Err() != nil {
		errMessage = "not found"
	}

	if err := findRes.All(o.ctx, &data); err != nil {
		log.Println(err)
	}

	jString, _ := json.Marshal(data)

	err = client.Set(ctx, key, string(jString), 1*time.Minute).Err()
	if err != nil {
		log.Println(err)
	}
	//err = o.redisUtil.Set(key, string(jString), 1*time.Minute)
	//if err != nil {
	//	log.Println(err)
	//}

	return

}
