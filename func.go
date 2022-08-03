/*
 * Copyright © 2016-2022 Iury Braun
 * Copyright © 2017-2022 Neo7i
 */

package go_mongodb_dao

import (
    //"log"
    "regexp"
    
    //"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IsValidObjectId(ObjectId string) bool {
    re := regexp.MustCompile("^[0-9a-fA-F]{24}$")
    if ok := re.FindString(ObjectId); ok != "" {
        return true
    }
    
    return false
}

func NewObjectID() primitive.ObjectID {
    return primitive.NewObjectID()
}



/* INDEX */
/*err_index := mgodao.CreateIndex(req, "properties", []string{"$text:location.neighborhood", "$text:location.city"})
if err_index != nil {
    log.Println(err_index)
}*/
/***
func (m *Dao) CreateIndex(IndexString []string) error {
    index := mgo.Index{
        Key: IndexString,
    }
    
    if err:= db.C(m.Collection).EnsureIndex(index); err != nil{
        if err.Error() != "not found" {
            log.Println(err)
        }
		return err
	}
    
	return nil
}
**/
/*
func GetIndex(req *http.Request, Collection string, id interface{}) error {
    return collection(req, Collection).Remove(bson.M{"_id": id})
    
    count, err := collection(req, Collection).Find(&qry).Select(&prj).Count()
    if err != nil {
        log.Println(err)
        return nil, count, err
    }
}
*/
