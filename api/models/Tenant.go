package models

import (
	"errors"
	"html"
	"net/http"
	"siot/api/utils/pagination"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Tenant struct {
	ID          uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	Name        string    `validate:"required" gorm:"size:255;not null;" json:"name"`
	Description string    `gorm:"size:255;" json:"description"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Status      string    `gorm:"size:255;default:'active'" json:"status"`
	Users       []User    `gorm:"many2many:user_tenants;association_jointable_foreignkey:user_id" json:"-"`
}

type UserTenant struct {
	ID        uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID
	TenantID  uuid.UUID
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Status    bool      `gorm:"default:true"`
}

func (p *Tenant) Prepare() {
	p.Name = html.EscapeString(strings.TrimSpace(p.Name))
	p.Description = html.EscapeString(strings.TrimSpace(p.Description))
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
}

func (p *Tenant) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func (t *UserTenant) ValidateTenantPermission(db *gorm.DB, user_id uuid.UUID, tenant_id uuid.UUID) bool {

	count := int64(0)

	var err error = db.Where("user_id = ? AND tenant_id = ?", user_id, tenant_id).Find(&t).Count(&count).Error
	if err != nil {
		return false
	}

	if count > 0 {
		return true
	}

	return false
}

func (t *Tenant) SaveTenant(db *gorm.DB, user_id uuid.UUID) (*Tenant, error) {
	var err error

	// get user for the association
	user := User{
		ID: user_id,
	}

	// create tenant
	err = db.Model(&Tenant{}).Create(&t).Error
	if err != nil {
		return &Tenant{}, err
	}

	// create association
	var errAssociation error = db.Model(&t).Association("Users").Append([]User{user}).Error
	if errAssociation != nil {
		return &Tenant{}, errAssociation
	}

	return t, nil
}

func (t *Tenant) FindAllTenants(db *gorm.DB, user_id string, r *http.Request) (*[]Tenant, int, int, int, int, interface{}, interface{}, error) {

	tenants := []Tenant{}

	var count int

	var err_count error = db.Joins("JOIN user_tenants ON user_tenants.tenant_id = tenants.id").Where("user_tenants.user_id = ?", user_id).Find(&tenants).Count(&count).Error
	if err_count != nil {
		return &[]Tenant{}, 0, 0, 0, 0, nil, nil, err_count
	}

	// pagination
	offset, limit, page, totalPages, nextPage, previousPage, errPagination := pagination.ValidatePagination(r, count)
	if errPagination != nil {
		return &[]Tenant{}, 0, 0, 0, 0, nil, nil, errPagination
	}

	var err error = db.Joins("JOIN user_tenants ON user_tenants.tenant_id = tenants.id").Where("user_tenants.user_id = ?", user_id).Limit(limit).Offset(offset).Order("updated_at desc").Find(&tenants).Error
	if err != nil {
		return &[]Tenant{}, 0, 0, 0, 0, nil, nil, err
	}
	return &tenants, limit, page, count, totalPages, nextPage, previousPage, err
}
