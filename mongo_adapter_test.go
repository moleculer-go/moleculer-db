package db

import (
	"time"

	snap "github.com/moleculer-go/cupaloy"
	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/payload"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//var MongoTestsHost = "mongodb://192.168.1.110"
var MongoTestsHost = "mongodb://localhost"

func mongoAdapter(database, collection string) *MongoAdapter {
	return &MongoAdapter{
		MongoURL:   MongoTestsHost,
		Timeout:    2 * time.Second,
		Database:   database,
		Collection: collection,
	}
}

type M map[string]interface{}

var _ = Describe("Mongo Adapter", func() {
	adapter := mongoAdapter("mongo_adapter_tests", "user")
	totalRecords := 6
	var johnSnow, marie moleculer.Payload
	BeforeEach(func() {
		johnSnow, marie, _ = connectAndLoadUsers(adapter)
	})

	AfterEach(func() {
		adapter.Disconnect()
	})

	Describe("Count", func() {
		It("should count the number of records properly", func() {
			result := adapter.Count(payload.New(M{}))
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Int()).Should(Equal(totalRecords))
		})

		It("should count the number of records and apply filter", func() {
			result := adapter.Count(payload.New(M{"query": M{
				"name": "John",
			}}))
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Int()).Should(Equal(2))
		})

	})

	Describe("Find", func() {

		It("should find using an empty query and return all records", func() {
			result := adapter.Find(payload.New(M{}))
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Len()).Should(Equal(totalRecords))
		})

		//Sort apprently not working in this client
		XIt("should sort the results", func() {
			result := adapter.Find(payload.New(M{
				"query": M{},
				"sort":  "name",
			}))
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Len()).Should(Equal(totalRecords))
			Expect(snap.SnapshotMulti("sort-1", result.Remove("_id"))).Should(Succeed())

			result2 := adapter.Find(payload.New(M{
				"query": M{},
				"sort":  "age",
			}))
			Expect(result2.IsError()).Should(BeFalse())
			Expect(result2.Len()).Should(Equal(totalRecords))
			result2 = result2.Remove("_id")
			Expect(snap.SnapshotMulti("sort-2", result2)).Should(Succeed())

			//shuold not match sort-1
			Expect(snap.SnapshotMulti("sort-1", result2)).ShouldNot(Succeed())
		})

		It("should offset the results", func() {
			result := adapter.Find(payload.New(M{
				"query":  M{},
				"sort":   "name",
				"offset": 2,
			}))
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Len()).Should(Equal(totalRecords - 2))
			Expect(snap.SnapshotMulti("offset-2", result.Remove("_id"))).Should(Succeed())

			result = adapter.Find(payload.New(M{
				"query":  M{},
				"sort":   "name",
				"offset": 4,
			}))
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Len()).Should(Equal(totalRecords - 4))
			Expect(snap.SnapshotMulti("offset-4", result.Remove("_id"))).Should(Succeed())
		})

		It("should find using an empty query and limit = 3 and return 3 records", func() {
			query := M{
				"query": M{},
				"limit": 3,
			}
			p := payload.New(query)
			r := adapter.Find(p)

			Expect(r.IsError()).Should(BeFalse())
			Expect(r.Len()).Should(Equal(3))
		})

		It("should search using search/searchFields params", func() {

			p := payload.New(map[string]interface{}{
				"search":       "John",
				"searchFields": []string{"name", "midlename"},
			})
			r := adapter.Find(p)

			Expect(r).ShouldNot(BeNil())
			Expect(r.Len()).Should(Equal(2))
		})

		It("should search using curtom query param", func() {
			query := M{
				"query": M{
					"age": M{
						"$gt": 60,
					},
				},
			}
			p := payload.New(query)
			r := adapter.Find(p)

			Expect(r.IsError()).Should(BeFalse())
			Expect(r.Len()).Should(Equal(2))

			query = M{
				"query": M{
					"$or": []M{
						M{"name": "John"},
						M{"lastname": "Claire"},
					},
				},
			}
			p = payload.New(query)
			r = adapter.Find(p)

			Expect(r.IsError()).Should(BeFalse())
			Expect(r.Len()).Should(Equal(3))
		})

	})

	Describe("FindById", func() {
		It("should find a records by its ID", func() {
			result := adapter.FindById(johnSnow.Get("_id"))
			Expect(result.Exists()).Should(BeTrue())
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Get("name").String()).Should(Equal(johnSnow.Get("name").String()))
			Expect(result.Get("lastname").String()).Should(Equal(johnSnow.Get("lastname").String()))
			Expect(result.Get("age").Int()).Should(Equal(johnSnow.Get("age").Int()))

			result = adapter.FindById(marie.Get("_id"))
			Expect(result.IsError()).Should(BeFalse())
			Expect(result.Get("name").String()).Should(Equal(marie.Get("name").String()))
			Expect(result.Get("lastname").String()).Should(Equal(marie.Get("lastname").String()))
			Expect(result.Get("age").Int()).Should(Equal(marie.Get("age").Int()))
		})
	})

})
