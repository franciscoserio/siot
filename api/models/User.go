package models

import (
	"errors"
	"html"
	"log"
	"net/http"
	"os"
	"siot/api/serializers"
	"siot/api/utils/email"
	"siot/api/utils/formaterror"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	FirstName       string    `validate:"required" gorm:"size:255;not null" json:"first_name"`
	LastName        string    `validate:"required" gorm:"size:255;not null" json:"last_name"`
	Email           string    `validate:"email,required" gorm:"size:100;not null;unique" json:"email"`
	Password        string    `validate:"required" gorm:"size:100;not null;" json:"password"`
	IsAdmin         bool      `gorm:"default:true"`
	IsSuperAdmin    bool      `gorm:"default:false"`
	Status          string    `gorm:"size:255;default:'active'"`
	InvitationToken string    `gorm:"size:255;"`
	CreatedAt       time.Time `validate:"required" gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `validate:"required" gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Tenants         []Tenant  `gorm:"many2many:user_tenants;association_jointable_foreignkey:tenant_id" json:"-"`
}

func (u *User) ShowUserSerializer() serializers.ShowUserSerializer {
	return serializers.ShowUserSerializer{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Tenants:   u.Tenants,
	}
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// func (u *User) BeforeSave() error {
// 	u.Status = "invited"
// 	u.InvitationToken = randStr(30)
// 	u.Password = ""
// 	return nil
// }

func (u *User) Prepare() {
	u.FirstName = html.EscapeString(strings.TrimSpace(u.FirstName))
	u.LastName = html.EscapeString(strings.TrimSpace(u.LastName))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}

func (u *User) Validate(action string) []string {
	var errors []string

	switch strings.ToLower(action) {
	case "update":
		if u.FirstName == "" {
			errors = append(errors, "error 1")
		}
		if u.Password == "" {
			errors = append(errors, "error 1")
		}
		if u.Email == "" {
			errors = append(errors, "error 1")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			errors = append(errors, "error 1")
		}

		return errors

	case "login":
		if u.Password == "" {
			errors = append(errors, "password is required")
		}
		if u.Email == "" {
			errors = append(errors, "email is required")
		}

		return errors

	default:
		if u.FirstName == "" {
			errors = append(errors, "error 1")
		}
		if u.Password == "" {
			errors = append(errors, "error 1")
		}
		if u.Email == "" {
			errors = append(errors, "error 1")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			errors = append(errors, "error 1")
		}

		return errors
	}
}

func (u *User) UserValidations(action string, db *gorm.DB) formaterror.GeneralError {

	var errors formaterror.GeneralError

	switch strings.ToLower(action) {
	case "confirm":

		if u.Password == "" {
			errors.Errors = append(errors.Errors, "password is required")
		}
		if len(u.Password) < 5 {
			errors.Errors = append(errors.Errors, "password must have at least 5 characters")
		}

		return errors

	default:
		if u.FirstName == "" {
			errors.Errors = append(errors.Errors, "first name is required")
		}
		if len(u.FirstName) > 255 {
			errors.Errors = append(errors.Errors, "first name is too long")
		}
		if u.LastName == "" {
			errors.Errors = append(errors.Errors, "last name is required")
		}
		if len(u.LastName) > 255 {
			errors.Errors = append(errors.Errors, "last name is too long")
		}
		if u.Email == "" {
			errors.Errors = append(errors.Errors, "email is required")
		}
		if len(u.Email) > 100 {
			errors.Errors = append(errors.Errors, "email is too long")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			errors.Errors = append(errors.Errors, "invalid email")
		}
		if EmailAlreadyExists(db, u.Email) {
			errors.Errors = append(errors.Errors, "an account with this email already exists")
		}
		return errors
	}
}

func EmailAlreadyExists(db *gorm.DB, email string) bool {

	var users []User
	db.Model(User{}).Where("email = ?", email).Find(&users)
	return len(users) > 0
}

func (u *User) CreateSuperAdminUser(db *gorm.DB) bool {

	// get super admin email
	var superAdminEmail string = os.Getenv("SUPER_ADMIN_EMAIL")
	if err := checkmail.ValidateFormat(superAdminEmail); err != nil {
		log.Fatalln("invalid super admin email")
		return false
	}

	// super admin password
	hashedPassword, _ := Hash("admin")
	var stringHashedPassword = string(hashedPassword)

	u.FirstName = "admin"
	u.LastName = "admin"
	u.Email = superAdminEmail
	u.Password = stringHashedPassword
	u.IsSuperAdmin = true

	// check if super admin user was already created
	var users []User
	db.Model(User{}).Where("is_super_admin = ?", true).Find(&users)
	if len(users) > 0 {
		return false
	}

	// create super admin user
	db.Create(&u)
	return true
}

func (u *User) SaveUser(db *gorm.DB) (*User, error) {

	var err error = db.Create(&u).Error
	if err != nil {
		return &User{}, err
	}

	to := []string{
		u.Email,
	}

	confirmationUrl := os.Getenv("SERVER_URL") + "/users/confirmation?confirmation_token=" + u.InvitationToken

	go email.SendConfirmationEmail(to, "[SIOT] Confirmation email", u.FirstName, u.LastName, confirmationUrl)

	return u, nil
}

func (u *User) ConfirmUser(db *gorm.DB, r *http.Request) (*User, error) {

	var user User
	err := db.Model(User{}).Where("invitation_token = ?", r.URL.Query().Get("confirmation_token")).Preload("Tenants").Take(&user).Error
	if err != nil {
		return nil, errors.New("wrong confirmation token")
	}

	if user.Status == "active" {
		return nil, errors.New("user already confirmed")
	}

	hashedPassword, _ := Hash(u.Password)
	user.Password = string(hashedPassword)
	user.Status = "active"

	var err_update error = db.Model(&User{}).Where("id = ?", user.ID).Updates(&user).Error

	if err_update != nil {
		return nil, err_update
	}

	return &user, nil
}

func (u *User) FindAllUsers(db *gorm.DB) (*[]User, error) {

	var err error
	users := []User{}
	err = db.Model(&User{}).Preload("Tenants").Limit(100).Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindUserByEmail(db *gorm.DB, email string) (*User, error) {

	var err error = db.Model(User{}).Where("email = ?", email).Preload("Tenants").Find(&u).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return u, err
}

func (u *User) FindUserByID(db *gorm.DB, uid string) (*User, error) {

	var err error = db.Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return u, err
}

// func (u *User) UpdateAUser(db *gorm.DB, uid uint32) (*User, error) {

// 	// To hash the password
// 	err := u.BeforeSave()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	db = db.Model(&User{}).Where("id = ?", uid).Take(&User{}).UpdateColumns(
// 		map[string]interface{}{
// 			"password":   u.Password,
// 			"first_name": u.FirstName,
// 			"email":      u.Email,
// 			"updated_at": time.Now(),
// 		},
// 	)
// 	if db.Error != nil {
// 		return &User{}, db.Error
// 	}
// 	// This is the display the updated user
// 	err = db.Model(&User{}).Where("id = ?", uid).Take(&u).Error
// 	if err != nil {
// 		return &User{}, err
// 	}
// 	return u, nil
// }

func (u *User) DeleteAUser(db *gorm.DB, uid uint32) (int64, error) {

	db = db.Model(&User{}).Where("id = ?", uid).Take(&User{}).Delete(&User{})

	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
