package fake

import (
	"encoding/json"

	"github.com/cloudnativego/cfmgo"
	"gopkg.in/mgo.v2"
)

var TargetCount int = 1

//FakesNewCollectionDialer
func FakesNewCollectionDialer(c interface{}) func(url, dbname, collectionname string) (col cfmgo.Collection, err error) {
	b, err := json.Marshal(c)
	if err != nil {
		panic("Unexpected Error: Unable to marshal fake data.")
	}

	return func(url, dbname, collectioname string) (col cfmgo.Collection, err error) {
		col = &FakeCollection{
			Data: b,
		}
		return
	}
}

//FakeCollection
type FakeCollection struct {
	mgo.Collection
	Data  []byte
	Error error
}

//Close
func (s *FakeCollection) Close() {}

//Wake
func (s *FakeCollection) Wake() {}

//Find -- finds all records matching given selector
func (s *FakeCollection) Find(params cfmgo.Params, result interface{}) (count int, err error) {
	count = TargetCount
	err = json.Unmarshal(s.Data, result)

	return
}
