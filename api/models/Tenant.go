package models

import (
	"html"
	"net/http"
	"siot/api/utils/formaterror"
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
	Status    string    `gorm:"size:255;default:'active'" json:"status"`
}

func (t *Tenant) BeforeCreate() {

	if t.Status != "active" && t.Status != "inactive" {
		t.Status = "active"
	}

	t.Name = html.EscapeString(strings.TrimSpace(t.Name))
	t.Description = html.EscapeString(strings.TrimSpace(t.Description))
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
}

func (t *Tenant) PrepareUpdate() {

	if t.Status != "active" && t.Status != "inactive" {
		t.Status = ""
	}

	t.Name = html.EscapeString(strings.TrimSpace(t.Name))
	t.Description = html.EscapeString(strings.TrimSpace(t.Description))
	t.UpdatedAt = time.Now()
}

func (t *Tenant) TenantValidations() formaterror.GeneralError {

	var errors formaterror.GeneralError

	if t.Name == "" {
		errors.Errors = append(errors.Errors, "name is required")
	}
	if len(t.Name) > 255 {
		errors.Errors = append(errors.Errors, "name is too long")
	}
	if len(t.Description) > 255 {
		errors.Errors = append(errors.Errors, "description is too long")
	}
	return errors
}

func (t *Tenant) IsActive(db *gorm.DB, tenant_id uuid.UUID) (bool, error) {

	var tenant Tenant
	var err error = db.Where("id = ?", tenant_id).Find(&tenant).Error
	if err != nil {
		return false, err
	}

	if tenant.Status != "active" {
		return false, nil
	}

	return true, nil
}

func (t *UserTenant) ValidateTenantPermission(db *gorm.DB, user_id uuid.UUID, tenant_id uuid.UUID) int {

	var userTenants []UserTenant
	var err error = db.Where("user_id = ? AND tenant_id = ?", user_id, tenant_id).Find(&userTenants).Error
	if err != nil {
		return -1
	}

	if len(userTenants) > 0 {

		var user User
		var userError error = db.Where("id = ?", user_id).Find(&user).Error
		if userError != nil {
			return -1
		}

		// if admin, can access
		if user.IsAdmin {
			return 1

		} else {

			// if user is not active for that tenant
			if userTenants[0].Status != "active" {
				return -2
			}

			// if tenant is not active
			var tenant Tenant
			var errTenant error = db.Where("id = ?", tenant_id).Find(&tenant).Error
			if errTenant != nil {
				return -1
			}

			if tenant.Status != "active" {
				return -3
			}
		}

		return 1
	}

	return -1
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

func (t *Tenant) GetTenant(db *gorm.DB, tenant_id string) (*Tenant, error) {

	tenant := Tenant{}

	var err error = db.Where("id = ?", tenant_id).Find(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, err
}

func (t *Tenant) UpdateTenant(db *gorm.DB, tenant_id string) (*Tenant, error) {

	var err error = db.Model(&Tenant{}).Where("id = ?", tenant_id).Updates(&t).Error

	if err != nil {
		return nil, err
	}

	// get the updated device
	var err_get_device error = db.Model(&Device{}).Where("id = ?", tenant_id).Take(&t).Error
	if err_get_device != nil {
		return nil, err_get_device
	}
	return t, nil
}
