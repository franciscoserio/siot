package seed

import (
	"fmt"
	"log"

	"siot/api/models"

	"github.com/jinzhu/gorm"
)

func Load(db *gorm.DB) {

	// Drop all tables
	// errDrop := db.DropTableIfExists(&models.Rule{}, &models.Sensor{}, &models.Device{}, &models.UserTenant{}, &models.Tenant{}, &models.User{}).Error
	// if errDrop != nil {
	// 	log.Fatalf("cannot drop table: %v", errDrop)
	// }

	// Migration
	err := db.AutoMigrate(&models.User{}, &models.Tenant{}, &models.UserTenant{}, &models.Device{}, &models.Sensor{}, &models.Rule{}).Error
	if err != nil {
		log.Fatalf("cannot migrate table: %v", err)
	}

	// Add foreign keys
	// user_tenants
	db.Table("user_tenants").AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Table("user_tenants").AddForeignKey("tenant_id", "tenants(id)", "CASCADE", "CASCADE")

	// devices
	db.Table("devices").AddForeignKey("tenant_id", "tenants(id)", "CASCADE", "CASCADE")

	// sensors
	db.Table("sensors").AddForeignKey("device_id", "devices(id)", "CASCADE", "CASCADE")

	// rules
	db.Table("rules").AddForeignKey("device_id", "devices(id)", "CASCADE", "CASCADE")

	// Create super admin user if not exists
	superAdmin := models.User{}

	var errCreateSuperUser bool = superAdmin.CreateSuperAdminUser(db)
	if errCreateSuperUser {
		fmt.Println("Super user created. Do not forget to change the password!")
	}
}
