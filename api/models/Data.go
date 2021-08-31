package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"siot/api/utils/pagination"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Data struct {
	Data []map[string]interface{} `validate:"required" json:"data" bson:"data"`
}

func (d *Data) ValidateAndSendData(dbm *mongo.Client, db *gorm.DB, device_id uuid.UUID) error {

	device := Device{}
	device_streams, _ := device.FindDevice(db, device_id)
	var values []interface{}

	// streams of the current device
	var list_device_streams []string

	for i := 0; i < len(device_streams.Streams); i++ {
		list_device_streams = append(list_device_streams, device_streams.Streams[i].Name)
	}

	// body streams
	var body_streams []string

	for i := 0; i < len(d.Data); i++ {

		// validate collection date
		if d.Data[i]["collected_at"] == nil {
			return errors.New("collect date is missing")
		}

		_, err := time.Parse("2006-01-02T15:04:05.000Z", fmt.Sprintf("%v", d.Data[i]["collected_at"]))
		if err != nil {
			return errors.New("collect date is in the wrong format")
		}

		// body keys
		for key, _ := range d.Data[i] {
			if !stringInSlice(key, body_streams) && key != "collected_at" {
				body_streams = append(body_streams, key)
			}
		}

		values = append(values, d.Data[i])
	}

	// check for active/inactive streams
	for i := 0; i < len(body_streams); i++ {
		for j := 0; j < len(device_streams.Streams); j++ {
			if body_streams[i] == device_streams.Streams[i].Name {
				if device_streams.Streams[i].Status != "active" {
					return errors.New("You can't send data with stream " + device_streams.Streams[i].Name + " because is inactive")
				}
			}
		}
	}

	// add streams
	for i := 0; i < len(body_streams); i++ {

		if !stringInSlice(body_streams[i], list_device_streams) {
			stream := Stream{}
			stream.Name = body_streams[i]
			stream.Prepare()
			stream.SaveStream(db, device_id)
		}
	}

	// send to mongodb
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := dbm.Database("siot").Collection(fmt.Sprintf("%v", device_id))
	_, err := collection.InsertMany(ctx, values)
	if err != nil {
		return err
	}

	// check rules
	_ = CheckRule(dbm, db, device_id)

	return nil
}

func (d *Data) GetData(dbm *mongo.Client, db *gorm.DB, device_id uuid.UUID, r *http.Request) (*Data, error) {

	// pagination
	offset, limit := pagination.ValidatePaginationData(r)

	// filter by date
	from, to := ValidateFromTo(r)

	// filter params
	var opt options.FindOptions
	streams, projections := SetStreams(db, r, device_id)

	opt.SetProjection(projections)
	opt.SetLimit(int64(limit))
	opt.SetSkip(int64(offset))

	// set filters to mongodb
	filter := bson.D{{"collected_at", bson.D{{"$gt", from}, {"$lt", to}}}}

	for i := 0; i < len(streams); i++ {
		filter = append(filter, bson.E{streams[i], bson.D{{"$exists", true}}})
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := dbm.Database("siot").Collection(fmt.Sprintf("%v", device_id))
	cur, err := collection.Find(ctx, filter, &opt)
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &d.Data); err != nil {
		return nil, errors.New("error returning data")
	}

	return d, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ValidateFromTo(r *http.Request) (string, string) {

	var from = ""
	var to = ""

	// validate page and limit
	if r.URL.Query().Get("from") != "" {

		_, err := time.Parse("2006-01-02T15:04:05.000Z", r.URL.Query().Get("from"))
		if err == nil {
			from = r.URL.Query().Get("from")
		}

	}

	if r.URL.Query().Get("to") != "" {

		_, err := time.Parse("2006-01-02T15:04:05.000Z", r.URL.Query().Get("to"))
		if err == nil {
			to = r.URL.Query().Get("to")
		}

	} else {
		datetime_now := time.Now()
		to = fmt.Sprintf("%v-%v-%vT%v:%v:%v.000Z", datetime_now.Year(), datetime_now.Month(), datetime_now.Day(), datetime_now.Hour(), datetime_now.Minute(), datetime_now.Second())
	}

	return from, to
}

func SetStreams(db *gorm.DB, r *http.Request, device_id uuid.UUID) ([]string, bson.M) {

	// device streams
	device := Device{}
	device_streams, _ := device.FindDevice(db, device_id)

	// url streams
	url_streams := r.URL.Query()["streams"]

	// final array of streams
	var streams []string
	streams = append(streams, "collected_at")

	if len(url_streams) == 0 {
		for i := 0; i < len(device_streams.Streams); i++ {
			streams = append(streams, device_streams.Streams[i].Name)
		}

	} else {

		for i := 0; i < len(url_streams); i++ {
			for j := 0; j < len(device_streams.Streams); j++ {

				if url_streams[i] == device_streams.Streams[j].Name {
					if !stringInSlice(url_streams[i], streams) && url_streams[i] != "collected_at" {
						streams = append(streams, url_streams[i])
					}

				}
			}
		}
	}

	if len(streams) == 1 {
		for i := 0; i < len(device_streams.Streams); i++ {
			streams = append(streams, device_streams.Streams[i].Name)
		}
	}

	proj := bson.M{}
	proj["_id"] = 0

	for i := 0; i < len(streams); i++ {
		proj[streams[i]] = 1
	}

	return streams, proj
}
