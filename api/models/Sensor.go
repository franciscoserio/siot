package models

import (
	"context"
	"errors"
	"html"
	"net/http"
	"siot/api/utils/formaterror"
	"siot/api/utils/pagination"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Sensor struct {
	ID          uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	Name        string    `validate:"required" gorm:"size:255;not null;" json:"name"`
	Description string    `gorm:"size:255;" json:"description"`
	Unit        string    `gorm:"size:255;" json:"unit"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Status      string    `gorm:"size:255;default:'active'" json:"status"`
	DeviceID    uuid.UUID `sql:"type:uuid REFERENCES devices(id) ON DELETE CASCADE" json:"-"`
}

func (s *Sensor) BeforeCreate() {

	s.Name = html.EscapeString(strings.TrimSpace(s.Name))
	s.Description = html.EscapeString(strings.TrimSpace(s.Description))
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.Status = strings.ToLower(s.Status)
	s.Unit = html.EscapeString(strings.TrimSpace(s.Unit))

	if s.Status != "active" && s.Status != "inactive" {
		s.Status = "active"
	}
}

func (s *Sensor) PrepareUpdate() {
	s.Name = html.EscapeString(strings.TrimSpace(s.Name))
	s.Description = html.EscapeString(strings.TrimSpace(s.Description))
	s.UpdatedAt = time.Now()
	s.Status = strings.ToLower(s.Status)
	s.Unit = html.EscapeString(strings.TrimSpace(s.Unit))

	if s.Status != "active" && s.Status != "inactive" {
		s.Status = ""
	}
}

func (s *Sensor) SensorValidations() formaterror.GeneralError {

	var errors formaterror.GeneralError

	if s.Name == "" {
		errors.Errors = append(errors.Errors, "name is required")
	}
	if len(s.Name) > 255 {
		errors.Errors = append(errors.Errors, "name is too long")
	}
	if len(s.Description) > 255 {
		errors.Errors = append(errors.Errors, "description is too long")
	}
	if len(s.Unit) > 255 {
		errors.Errors = append(errors.Errors, "unit is too long")
	}
	return errors
}

func (s *Sensor) SaveSensor(db *gorm.DB, device_id uuid.UUID) (*Sensor, error) {

	// get device for the association
	s.DeviceID = device_id

	sensors := []Sensor{}

	// check if sensor already exists for that device
	var count int

	var err_get_sensor_count error = db.Select("sensors.name").Joins("join devices on sensors.device_id = devices.id").Where("sensors.name = ? AND sensors.device_id = ?", s.Name, device_id).Find(&sensors).Count(&count).Error
	if err_get_sensor_count != nil {
		return nil, err_get_sensor_count
	}

	if count > 0 {
		return nil, errors.New("sensor already exists")
	}

	// create sensor
	err := db.Model(&Sensor{}).Create(&s).Error
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sensor) FindAllSensors(db *gorm.DB, device_id string, r *http.Request) (interface{}, error) {

	sensors := []Sensor{}

	var count int

	var err_count error = db.Select("sensors.name").Joins("join devices on sensors.device_id = devices.id").Where("sensors.device_id = ?", device_id).Find(&sensors).Count(&count).Error
	if err_count != nil {
		return nil, err_count
	}

	// pagination
	offset, limit, page, totalPages, nextPage, previousPage, errPagination := pagination.ValidatePagination(r, count)
	if errPagination != nil {
		return nil, errPagination
	}

	// query
	err := db.Select("sensors.id, sensors.name, sensors.description, sensors.unit, sensors.created_at, sensors.updated_at, sensors.status").Joins("join devices on sensors.device_id = devices.id").Where("sensors.device_id = ?", device_id).Limit(limit).Offset(offset).Find(&sensors).Error
	if err != nil {
		return nil, err
	}
	return pagination.ListPaginationSerializer(limit, page, count, totalPages, nextPage, previousPage, sensors), nil
}

func (s *Sensor) IsValidSensor(db *gorm.DB, sensor_id uuid.UUID, device_id uuid.UUID) (bool, error) {

	sensors := []Sensor{}

	// query
	err := db.Select("sensors.id, sensors.device_id").Joins("join devices on sensors.device_id = devices.id").Where("sensors.id = ? AND sensors.device_id = ?", sensor_id, device_id).Find(&sensors).Error
	if err != nil {
		return false, err
	}

	if len(sensors) > 0 {
		return true, nil
	}

	return false, nil
}

func (s *Sensor) IsValidSensorName(db *gorm.DB, sensor_name string, device_id uuid.UUID) bool {

	sensors := []Sensor{}

	// query
	err := db.Select("sensors.name, sensors.device_id").Joins("join devices on sensors.device_id = devices.id").Where("sensors.name = ? AND sensors.device_id = ?", sensor_name, device_id).Find(&sensors).Error
	if err != nil {
		return false
	}

	if len(sensors) > 0 {
		return true
	}

	return false
}

func (s *Sensor) GetSensor(db *gorm.DB, sensor_id string) (*Sensor, error) {

	sensor := Sensor{}

	// query
	err := db.Model(&Sensor{}).Where("id = ?", sensor_id).Take(&sensor).Error
	if err != nil {
		return nil, err
	}
	return &sensor, nil
}

func (s *Sensor) UpdateSensor(dbm *mongo.Client, db *gorm.DB, sensor_id string, device_id string) (*Sensor, error) {

	sensor, _ := s.GetSensor(db, sensor_id)

	// if updating
	if sensor.Name != s.Name {

		var count int
		sensors := []Sensor{}
		var err_get_sensor_count error = db.Select("sensors.name").Joins("join devices on sensors.device_id = devices.id").Where("sensors.name = ? AND sensors.device_id = ?", s.Name, device_id).Find(&sensors).Count(&count).Error
		if err_get_sensor_count != nil {
			return nil, err_get_sensor_count
		}

		if count > 0 {
			return nil, errors.New("sensor with that name already exists")
		}

	}

	var err error = db.Model(&Sensor{}).Where("id = ?", sensor_id).Updates(&s).Error

	if err != nil {
		return nil, err
	}

	// get the updated sensor
	var err_get_sensor error = db.Model(&Sensor{}).Where("id = ?", sensor_id).Take(&s).Error
	if err_get_sensor != nil {
		return nil, err_get_sensor
	}

	if sensor.Name != s.Name {
		// update all sensor names in mongodb
		go UpdateSensorMongo(dbm, sensor.Name, s.Name, sensor.DeviceID.String())
	}

	return s, nil
}

func UpdateSensorMongo(dbm *mongo.Client, oldSensorName, newSensorNamedeviceId, deviceId string) {

	var filter primitive.D

	filter = append(filter, bson.E{oldSensorName, bson.D{{"$exists", true}}})

	update := bson.M{"$rename": bson.M{oldSensorName: newSensorNamedeviceId}}

	collection := dbm.Database("siot").Collection(deviceId)
	_, _ = collection.UpdateMany(context.Background(), filter, update)
}

func DeleteSensorMongo(db *gorm.DB, dbm *mongo.Client, sensorName, deviceId string) {

	collection := dbm.Database("siot").Collection(deviceId)

	// delete documents that have only the sensor
	sensors := SetSensorsToDelete(db, sensorName, deviceId)

	// set filters to mongodb
	var filter primitive.D
	filter = append(filter, bson.E{sensorName, bson.D{{"$exists", true}}})

	for i := 0; i < len(sensors); i++ {
		filter = append(filter, bson.E{sensors[i], bson.D{{"$exists", false}}})
	}

	_, _ = collection.DeleteMany(context.Background(), filter)

	// delete key of documents that have more than one sensor
	var filter_sensor_exists primitive.D

	filter_sensor_exists = append(filter_sensor_exists, bson.E{sensorName, bson.D{{"$exists", true}}})

	update := bson.M{"$unset": bson.M{sensorName: true}}

	_, _ = collection.UpdateMany(context.Background(), filter_sensor_exists, update)
}

func SetSensorsToDelete(db *gorm.DB, sensorName, deviceId string) []string {

	// convert device id to uuid
	did_uuid, _ := uuid.Parse(deviceId)

	// device sensors
	device := Device{}
	deviceSensors, _ := device.FindDevice(db, did_uuid)

	// final array of sensors
	var sensors []string

	for i := 0; i < len(deviceSensors.Sensors); i++ {
		if deviceSensors.Sensors[i].Name != sensorName {
			sensors = append(sensors, deviceSensors.Sensors[i].Name)
		}
	}

	return sensors
}

func (s *Sensor) DeleteSensor(dbm *mongo.Client, db *gorm.DB, sensor_id string) error {

	sensor, _ := s.GetSensor(db, sensor_id)

	var err error = db.Where("id = ?", sensor_id).Delete(&Sensor{}).Error

	if err != nil {
		return err
	}

	go DeleteSensorMongo(db, dbm, sensor.Name, sensor.DeviceID.String())

	return nil
}
