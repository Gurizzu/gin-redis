package model

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	Id          string `json:"id" bson:"_id"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Created_At  int64  `json:"created_at" bson:"created_at"`
	Updated_At  int64  `json:"updated_at" bson:"updated_at"`
}

type Item_View struct {
	Item `bson:",inline"`
}

type Item_Search struct {
	Name string `json:"search"`
}

func (o *Item_Search) HandlerFilter(listFilterAnd *[]bson.M) {
	if search := o.Name; search != "" {
		*listFilterAnd = append(*listFilterAnd, bson.M{"name": primitive.Regex{Pattern: search, Options: "i"}})
	}
}
