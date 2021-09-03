package models

import (
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
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

func (s *Sensor) SaveSensor(db *gorm.DB, device_id uuid.UUID) (*Sensor, error) {
	var err error

	// get device for the association
	s.DeviceID = device_id

	// create tenant
	err = db.Model(&Sensor{}).Omit("Devices").Create(&s).Error
	if err != nil {
		return &Sensor{}, err
	}

	return s, nil
}
