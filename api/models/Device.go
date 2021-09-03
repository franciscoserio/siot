package models

import (
	"context"
	"crypto/rand"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"siot/api/utils/pagination"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type Device struct {
	ID          uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	Name        string    `validate:"required" gorm:"size:255;not null;" json:"name"`
	Description string    `gorm:"size:255;" json:"description"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Status      string    `gorm:"size:255;default:'active'" json:"status"`
	Latitude    float64   `validate:"required_with=Longitude,latitude" gorm:"type:decimal(10,8);default:0.0" json:"latitude"`
	Longitude   float64   `validate:"required_with=Longitude,latitude" gorm:"type:decimal(11,8);default:0.0" json:"longitude"`
	TenantID    uuid.UUID `sql:"type:uuid REFERENCES tenants(id)" json:"-"`
	SecretKey   string    `gorm:"size:255;" json:"secret_key"`
	Sensors     []Sensor  `gorm:"association_jointable_foreignkey:device_id, OnDelete:CASCADE" json:"sensors"`
}

func (d *Device) Prepare() {
	d.Name = html.EscapeString(strings.TrimSpace(d.Name))
	d.Description = html.EscapeString(strings.TrimSpace(d.Description))
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	d.Status = strings.ToLower(d.Status)
	d.SecretKey = html.EscapeString(strings.TrimSpace(d.SecretKey))

	if d.Status != "active" && d.Status != "inactive" {
		d.Status = "active"
	}

	if d.SecretKey == "" {
		d.SecretKey = randStr(25)
	}
}

func (d *Device) PrepareUpdate() {
	d.Name = html.EscapeString(strings.TrimSpace(d.Name))
	d.Description = html.EscapeString(strings.TrimSpace(d.Description))
	d.UpdatedAt = time.Now()
	d.Status = strings.ToLower(d.Status)
	d.SecretKey = html.EscapeString(strings.TrimSpace(d.SecretKey))

	if d.Status == "active" {
		d.Status = "active"
	} else if d.Status == "inactive" {
		d.Status = "inactive"
	} else {
		d.Status = ""
	}
}

func (d *Device) ValidateDevicePermission(db *gorm.DB, device_id uuid.UUID, tenant_id uuid.UUID) (*Device, error) {

	var err error = db.Where("id = ? and tenant_id = ?", device_id, tenant_id).Find(&d).Error
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Device) SaveDevice(dbm *mongo.Client, db *gorm.DB, tenant_id uuid.UUID) (*Device, error) {
	var err error

	// get user for the association
	d.TenantID = tenant_id

	// create tenant
	err = db.Model(&Device{}).Omit("Sensors").Create(&d).Error
	if err != nil {
		return &Device{}, err
	}

	// create collection
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := dbm.Database("siot").Collection(fmt.Sprintf("%v", d.ID))

	// create index
	index := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "collected_at", Value: bsonx.String("text")}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, errIndex := collection.Indexes().CreateMany(ctx, index, opts)
	if errIndex != nil {
		return &Device{}, errIndex
	}

	return d, nil
}

func (d *Device) FindAllDevices(db *gorm.DB, tenant_id string, r *http.Request) (interface{}, error) {

	devices := []Device{}

	var count int

	var err_count error = db.Where("tenant_id = ?", tenant_id).Find(&devices).Count(&count).Error
	if err_count != nil {
		return nil, err_count
	}

	// pagination
	offset, limit, page, totalPages, nextPage, previousPage, errPagination := pagination.ValidatePagination(r, count)
	if errPagination != nil {
		return nil, errPagination
	}

	// query
	var err error = db.Where("tenant_id = ?", tenant_id).Preload("Sensors").Limit(limit).Offset(offset).Order("updated_at desc").Find(&devices).Error
	if err != nil {
		return nil, err
	}

	return pagination.ListPaginationSerializer(limit, page, count, totalPages, nextPage, previousPage, devices), nil
}

func (d *Device) FindDevice(db *gorm.DB, device_id uuid.UUID) (*Device, error) {

	var err error = db.Model(&Device{}).Where("id = ?", device_id).Preload("Sensors").Take(&d).Error
	if err != nil {
		return &Device{}, err
	}
	return d, nil
}

func (d *Device) UpdateDevice(db *gorm.DB, device_id uuid.UUID) (*Device, error) {

	var err error = db.Model(&Device{}).Where("id = ?", device_id).Updates(&d).Error

	if err != nil {
		return &Device{}, err
	}

	// get the updated device
	var err_get_device error = db.Model(&Device{}).Where("id = ?", device_id).Take(&d).Error
	if err_get_device != nil {
		return &Device{}, err_get_device
	}
	return d, nil
}

func (d *Device) DeleteDevice(db *gorm.DB, device_id uuid.UUID) error {

	var err error = db.Where("id = ?", device_id).Delete(&Device{}).Error

	if err != nil {
		return err
	}
	return nil
}

func randStr(n int) (str string) {
	b := make([]byte, n)
	rand.Read(b)
	str = fmt.Sprintf("%x", b)
	return
}
