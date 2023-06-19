package ustring

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

func GetParamAsKey(param interface{}) (keyString string) {
	paramString, _ := json.Marshal(param)
	key := md5.Sum([]byte(paramString))
	keyString = fmt.Sprintf("%x", key)
	return
}
