package seed

import (
	"log"

	"siot/api/models"

	"github.com/jinzhu/gorm"
)

// var users = []models.User{
// 	models.User{
// 		FirstName: "Steven victor",
// 		Email:    "steven@gmail.com",
// 		Password: "password",
// 	},
// 	models.User{
// 		FirstName: "Martin Luther",
// 		Email:    "luther@gmail.com",
// 		Password: "password",
// 	},
// }

// var posts = []models.Post{
// 	models.Post{
// 		Title:   "Title 1",
// 		Content: "Hello world 1",
// 	},
// 	models.Post{
// 		Title:   "Title 2",
// 		Content: "Hello world 2",
// 	},
// }

func Load(db *gorm.DB) {

	// errDrop := db.DropTableIfExists(&models.Rule{}, &models.Stream{}, &models.Device{}, &models.UserTenant{}, &models.Post{}, &models.User{}, &models.Tenant{}).Error
	// if errDrop != nil {
	// 	log.Fatalf("cannot drop table: %v", errDrop)
	// }

	err := db.AutoMigrate(&models.User{}, &models.Tenant{}, &models.UserTenant{}, &models.Device{}, &models.Stream{}, &models.Rule{}).Error
	if err != nil {
		log.Fatalf("cannot migrate table: %v", err)
	}

	// user_tenants
	db.Table("user_tenants").AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Table("user_tenants").AddForeignKey("tenant_id", "tenants(id)", "CASCADE", "CASCADE")

	// devices
	db.Table("devices").AddForeignKey("tenant_id", "tenants(id)", "CASCADE", "CASCADE")

	// streams
	db.Table("streams").AddForeignKey("device_id", "devices(id)", "CASCADE", "CASCADE")

	// rules
	db.Table("rules").AddForeignKey("device_id", "devices(id)", "CASCADE", "CASCADE")

	/*
		err = db.Model(&models.Post{}).AddForeignKey("author_id", "users(id)", "cascade", "cascade").Error
		if err != nil {
			log.Fatalf("attaching foreign key error: %v", err)
		}
	*/

	// for i, _ := range users {
	// 	err = db.Model(&models.User{}).Create(&users[i]).Error
	// 	if err != nil {
	// 		log.Fatalf("cannot seed users table: %v", err)
	// 	}
	// 	posts[i].AuthorID = users[i].ID

	// 	err = db.Model(&models.Post{}).Create(&posts[i]).Error
	// 	if err != nil {
	// 		log.Fatalf("cannot seed posts table: %v", err)
	// 	}
	// }
}
