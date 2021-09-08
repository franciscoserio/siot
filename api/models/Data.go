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
	device_sensors, _ := device.FindDevice(db, device_id)
	var values []interface{}

	// sensors of the current device
	var list_device_sensors []string

	for i := 0; i < len(device_sensors.Sensors); i++ {
		list_device_sensors = append(list_device_sensors, device_sensors.Sensors[i].Name)
	}

	// body sensors
	var body_sensors []string

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
			if !stringInSlice(key, body_sensors) && key != "collected_at" {
				body_sensors = append(body_sensors, key)
			}
		}

		values = append(values, d.Data[i])
	}

	// check for active/inactive sensors
	for i := 0; i < len(body_sensors); i++ {
		for j := 0; j < len(device_sensors.Sensors); j++ {
			if body_sensors[i] == device_sensors.Sensors[i].Name {
				if device_sensors.Sensors[i].Status != "active" {
					return errors.New("You can't send data with sensor " + device_sensors.Sensors[i].Name + " because is inactive")
				}
			}
		}
	}

	// add sensors
	for i := 0; i < len(body_sensors); i++ {

		if !stringInSlice(body_sensors[i], list_device_sensors) {
			sensor := Sensor{}
			sensor.Name = body_sensors[i]
			sensor.SaveSensor(db, device_id)
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
	go CheckRule(dbm, db, device_id, d.Data[len(d.Data)-1])

	return nil
}

func (d *Data) GetData(dbm *mongo.Client, db *gorm.DB, device_id uuid.UUID, r *http.Request) (*Data, error) {

	// pagination
	offset, limit := pagination.ValidatePaginationData(r)

	// filter by date
	from, to := ValidateFromTo(r)

	// filter params
	var opt options.FindOptions
	sensors, projections := SetSensors(db, r, device_id)

	opt.SetProjection(projections)
	opt.SetLimit(int64(limit))
	opt.SetSkip(int64(offset))

	// set filters to mongodb
	filter := bson.D{{"collected_at", bson.D{{"$gt", from}, {"$lt", to}}}}

	for i := 0; i < len(sensors); i++ {
		filter = append(filter, bson.E{sensors[i], bson.D{{"$exists", true}}})
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

func SetSensors(db *gorm.DB, r *http.Request, device_id uuid.UUID) ([]string, bson.M) {

	// device sensors
	device := Device{}
	device_sensors, _ := device.FindDevice(db, device_id)

	// url sensors
	url_sensors := r.URL.Query()["sensors"]

	// final array of sensors
	var sensors []string
	sensors = append(sensors, "collected_at")

	if len(url_sensors) == 0 {
		for i := 0; i < len(device_sensors.Sensors); i++ {
			sensors = append(sensors, device_sensors.Sensors[i].Name)
		}

	} else {

		for i := 0; i < len(url_sensors); i++ {
			for j := 0; j < len(device_sensors.Sensors); j++ {

				if url_sensors[i] == device_sensors.Sensors[j].Name {
					if !stringInSlice(url_sensors[i], sensors) && url_sensors[i] != "collected_at" {
						sensors = append(sensors, url_sensors[i])
					}

				}
			}
		}
	}

	if len(sensors) == 1 {
		for i := 0; i < len(device_sensors.Sensors); i++ {
			sensors = append(sensors, device_sensors.Sensors[i].Name)
		}
	}

	proj := bson.M{}
	proj["_id"] = 0

	for i := 0; i < len(sensors); i++ {
		proj[sensors[i]] = 1
	}

	return sensors, proj
}
