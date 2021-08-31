package models

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Rule struct {
	ID                      uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	Stream                  string    `validate:"required" gorm:"size:255;not null;" json:"stream"`
	Description             string    `gorm:"size:255;" json:"description"`
	Operation               string    `gorm:"size:255;" json:"operation"`
	CountLatest             int64     `gorm:"default:1;" json:"count_latest"`
	Email                   string    `gorm:"size:255;" json:"email"`
	EndpointUrl             string    `gorm:"size:255;" json:"endpoint_url"`
	EndpointHeader          string    `sql:"type:jsonb" gorm:"size:255;" json:"endpoint_header"`
	EndpointPayload         string    `sql:"type:jsonb" gorm:"size:255;" json:"endpoint_payload"`
	Operator                string    `gorm:"size:255;" json:"operator"`
	Value                   string    `gorm:"size:255;" json:"value"`
	TimeBetweenNotification string    `gorm:"size:255;" json:"time_between_notification"`
	LastNotification        time.Time `gorm:"size:255;" json:"last_notification"`
	DeviceID                uuid.UUID `sql:"type:uuid REFERENCES devices(id) ON DELETE CASCADE" json:"-"`
	CreatedAt               time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (r *Rule) updateNotificationTime(db *gorm.DB) error {

	r.LastNotification = time.Now()
	var err error = db.Model(&Rule{}).Where("id = ?", r.DeviceID).Updates(&r).Error

	if err != nil {
		return err
	}

	return nil
}

// func (r *Rule) sendNotificationRequest(db *gorm.DB, value string) error {

// 	return nil
// }

func CheckRule(dbm *mongo.Client, db *gorm.DB, device_id uuid.UUID) error {

	var rules []Rule

	db.Where("device_id = ?", device_id).Find(&rules)

	for i := 0; i < len(rules); i++ {

		// only one device value
		if rules[i].CountLatest < 2 {
			onlyOneDeviceValue(dbm, db, rules[i], device_id)

			// N latest device data
		} else if rules[i].CountLatest > 1 {
			latestDeviceData(dbm, db, rules[i], device_id)
		}
	}

	return nil
}

func onlyOneDeviceValue(dbm *mongo.Client, db *gorm.DB, rule Rule, device_id uuid.UUID) error {

	// check notification time
	if !freeNotificationTime(rule.TimeBetweenNotification, rule.LastNotification) {
		return nil
	}

	// filter params
	var opt options.FindOptions
	opt.SetLimit(1)
	opt.SetSort(bson.M{"$natural": -1})

	// set filters to mongodb
	filter := bson.D{{rule.Stream, bson.D{{"$exists", true}}}}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := dbm.Database("siot").Collection(fmt.Sprintf("%v", device_id))
	cur, err := collection.Find(ctx, filter, &opt)
	if err != nil {
		return err
	}

	var data Data
	if err = cur.All(ctx, &data.Data); err != nil {
		return errors.New("error returning data")
	}

	if len(data.Data) > 0 {

		if rule.Operator == ">" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Stream]), 64); err == nil {
					if deviceValue > ruleValue {
						rule.updateNotificationTime(db)
						fmt.Println("É MAIOR. ENVIAR NOTIFICAÇÃO")
					}
				}
			}

		} else if rule.Operator == ">=" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Stream]), 64); err == nil {
					if deviceValue >= ruleValue {
						rule.updateNotificationTime(db)
						fmt.Println("É MAIOR OU IGUAL. ENVIAR NOTIFICAÇÃO")
					}
				}
			}
		} else if rule.Operator == "<" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Stream]), 64); err == nil {
					if deviceValue < ruleValue {
						rule.updateNotificationTime(db)
						fmt.Println("É MENOR. ENVIAR NOTIFICAÇÃO")
					}
				}
			}
		} else if rule.Operator == "<=" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Stream]), 64); err == nil {
					if deviceValue <= ruleValue {
						rule.updateNotificationTime(db)
						fmt.Println("É MENOR OU IGUAL. ENVIAR NOTIFICAÇÃO")
					}
				}
			}
		} else {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Stream]), 64); err == nil {
					if deviceValue == ruleValue {
						rule.updateNotificationTime(db)
						fmt.Println("É IGUAL. ENVIAR NOTIFICAÇÃO")
					}
				} else {
					if fmt.Sprintf("%v", data.Data[0][rule.Stream]) == rule.Value {
						rule.updateNotificationTime(db)
						fmt.Println("É IGUAL. ENVIAR NOTIFICAÇÃO")
					}
				}
			} else {
				if fmt.Sprintf("%v", data.Data[0][rule.Stream]) == rule.Value {
					rule.updateNotificationTime(db)
					fmt.Println("É IGUAL. ENVIAR NOTIFICAÇÃO")
				}
			}

		}
	}

	return nil
}

func latestDeviceData(dbm *mongo.Client, db *gorm.DB, rule Rule, device_id uuid.UUID) error {

	// check notification time
	if !freeNotificationTime(rule.TimeBetweenNotification, rule.LastNotification) {
		return nil
	}

	// filter params
	var opt options.FindOptions
	opt.SetLimit(rule.CountLatest)
	opt.SetSort(bson.M{"$natural": -1})

	// set filters to mongodb
	filter := bson.D{{rule.Stream, bson.D{{"$exists", true}}}}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := dbm.Database("siot").Collection(fmt.Sprintf("%v", device_id))
	cur, err := collection.Find(ctx, filter, &opt)
	if err != nil {
		return err
	}

	var data Data
	if err = cur.All(ctx, &data.Data); err != nil {
		return errors.New("error returning data")
	}

	if len(data.Data) > 0 {

		var calculatedValue float64

		// operation
		// sum
		if rule.Operation == "sum" {

			for _, value := range data.Data {

				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Stream]), 64); err == nil {
					calculatedValue = calculatedValue + deviceValue
				}

			}

		} else if rule.Operation == "mean" {

			var auxValue float64

			for _, value := range data.Data {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Stream]), 64); err == nil {
					auxValue = auxValue + deviceValue
				}
			}

			calculatedValue = auxValue / float64(len(data.Data))

		} else if rule.Operation == "median" {
			var auxValues []float64

			for _, value := range data.Data {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Stream]), 64); err == nil {
					auxValues = append(auxValues, deviceValue)
				}
			}

			// sort list of values
			sort.Float64s(auxValues)

			mNumber := len(auxValues) / 2

			// is even
			if len(auxValues)%2 == 0 {
				calculatedValue = (auxValues[mNumber-1] + auxValues[mNumber]) / 2

				// is odd
			} else {
				calculatedValue = auxValues[mNumber]
			}

		} else if rule.Operation == "max" {
			var auxValue float64

			for _, value := range data.Data {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Stream]), 64); err == nil {
					if deviceValue > auxValue {
						auxValue = deviceValue
					}
				}
			}

			calculatedValue = auxValue

		} else if rule.Operation == "min" {
			var auxValue float64

			for _, value := range data.Data {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Stream]), 64); err == nil {
					if auxValue == 0 {
						auxValue = deviceValue

					} else {
						if deviceValue < auxValue {
							auxValue = deviceValue
						}
					}
				}
			}

			calculatedValue = auxValue

		}

		// operator
		if rule.Operator == ">" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue > ruleValue {
					rule.updateNotificationTime(db)
					fmt.Println("SUM É MAIOR. ENVIAR NOTIFICAÇÃO")
				}
			}

		} else if rule.Operator == ">=" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue >= ruleValue {
					rule.updateNotificationTime(db)
					fmt.Println("SUM É MAIOR OU IGUAL. ENVIAR NOTIFICAÇÃO")
				}
			}
		} else if rule.Operator == "<" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue < ruleValue {
					rule.updateNotificationTime(db)
					fmt.Println("SUM É MENOR. ENVIAR NOTIFICAÇÃO")
				}
			}
		} else if rule.Operator == "<=" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue <= ruleValue {
					rule.updateNotificationTime(db)
					fmt.Println("SUM É MENOR OU IGUAL. ENVIAR NOTIFICAÇÃO")
				}
			}
		} else {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue == ruleValue {
					rule.updateNotificationTime(db)
					fmt.Println("SUM É IGUAL. ENVIAR NOTIFICAÇÃO")
				}
			}

		}

	}

	return nil
}

func freeNotificationTime(timeBetween string, lastNotification time.Time) bool {

	possibleTimeBetween := []string{"s", "m", "h"}

	for key, value := range possibleTimeBetween {
		if strings.Contains(timeBetween, value) {

			t := strings.Split(timeBetween, value)
			timeConverted, _ := strconv.ParseInt(t[0], 10, 64)
			var nowSubtracted time.Time

			if possibleTimeBetween[key] == "s" {
				nowSubtracted = time.Now().Add(-time.Second * time.Duration(timeConverted))

			} else if possibleTimeBetween[key] == "m" {
				nowSubtracted = time.Now().Add(-time.Minute * time.Duration(timeConverted))

			} else if possibleTimeBetween[key] == "h" {
				nowSubtracted = time.Now().Add(-time.Hour * time.Duration(timeConverted))

			}

			// compare datetimes
			if nowSubtracted.Equal(lastNotification) {
				return false

			} else if nowSubtracted.After(lastNotification) {
				return true
			}
		}
	}

	return false
}
