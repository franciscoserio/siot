package models

import (
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Stream struct {
	ID          uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	Name        string    `validate:"required" gorm:"size:255;not null;" json:"name"`
	Description string    `gorm:"size:255;" json:"description"`
	Unit        string    `gorm:"size:255;" json:"unit"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Status      string    `gorm:"size:255;default:'active'" json:"status"`
	DeviceID    uuid.UUID `sql:"type:uuid REFERENCES devices(id) ON DELETE CASCADE" json:"-"`
}

func (d *Stream) Prepare() {
	d.Name = html.EscapeString(strings.TrimSpace(d.Name))
	d.Description = html.EscapeString(strings.TrimSpace(d.Description))
	d.Unit = html.EscapeString(strings.TrimSpace(d.Description))
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	d.Status = strings.ToLower(d.Status)

	if d.Status != "active" && d.Status != "inactive" {
		d.Status = "active"
	}
}

func (s *Stream) SaveStream(db *gorm.DB, device_id uuid.UUID) (*Stream, error) {
	var err error

	// get device for the association
	s.DeviceID = device_id

	// create tenant
	err = db.Model(&Stream{}).Omit("Devices").Create(&s).Error
	if err != nil {
		return &Stream{}, err
	}

	return s, nil
}
