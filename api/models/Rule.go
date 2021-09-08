package models

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/smtp"
	"os"
	"siot/api/utils/formaterror"
	"siot/api/utils/pagination"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type Rule struct {
	ID                      uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v4()" json:"id"`
	Sensor                  string    `validate:"required" gorm:"size:255;not null;" json:"sensor"`
	Description             string    `gorm:"size:255;" json:"description"`
	Operation               string    `gorm:"size:255;" json:"operation"`
	CountLatest             int64     `gorm:"default:1;" json:"count_latest"`
	Email                   string    `gorm:"size:255;" json:"email"`
	EmailSubject            string    `gorm:"size:255;" json:"email_subject"`
	EmailBody               string    `gorm:"size:255;" json:"email_body"`
	EndpointUrl             string    `gorm:"size:255;" json:"endpoint_url"`
	EndpointHeader          JSONB     `sql:"type:jsonb" gorm:"size:255;" json:"endpoint_header"`
	EndpointPayload         JSONB     `sql:"type:jsonb" gorm:"size:255;" json:"endpoint_payload"`
	Operator                string    `gorm:"size:255;" json:"operator"`
	Value                   string    `gorm:"size:255;" json:"value"`
	TimeBetweenNotification string    `gorm:"size:255;" json:"time_between_notification"`
	LastNotification        time.Time `gorm:"size:255;" json:"last_notification"`
	DeviceID                uuid.UUID `sql:"type:uuid REFERENCES devices(id) ON DELETE CASCADE" json:"device_id"`
	TenantID                uuid.UUID `sql:"type:uuid REFERENCES tenants(id) ON DELETE CASCADE" json:"-"`
	Status                  string    `gorm:"size:255;" json:"status"`
	CreatedAt               time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (r *Rule) BeforeCreate() {

	r.Description = html.EscapeString(strings.TrimSpace(r.Description))
	r.Status = strings.ToLower(r.Status)
	r.Operation = html.EscapeString(strings.TrimSpace(r.Operation))
	r.Email = html.EscapeString(strings.TrimSpace(r.Email))
	r.EndpointUrl = html.EscapeString(strings.TrimSpace(r.EndpointUrl))
	r.Operator = html.EscapeString(strings.TrimSpace(r.Operator))
	r.Value = html.EscapeString(strings.TrimSpace(r.Value))
	r.TimeBetweenNotification = html.EscapeString(strings.TrimSpace(r.TimeBetweenNotification))
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	if r.Status != "active" && r.Status != "inactive" {
		r.Status = "active"
	}
}

func (r *Rule) PrepareUpdate() {

	r.Description = html.EscapeString(strings.TrimSpace(r.Description))
	r.Status = strings.ToLower(r.Status)
	r.Operation = html.EscapeString(strings.TrimSpace(r.Operation))
	r.Email = html.EscapeString(strings.TrimSpace(r.Email))
	r.EndpointUrl = html.EscapeString(strings.TrimSpace(r.EndpointUrl))
	r.Operator = html.EscapeString(strings.TrimSpace(r.Operator))
	r.Value = html.EscapeString(strings.TrimSpace(r.Value))
	r.TimeBetweenNotification = html.EscapeString(strings.TrimSpace(r.TimeBetweenNotification))
	r.UpdatedAt = time.Now()

	if r.Status != "active" && r.Status != "inactive" {
		r.Status = ""
	}
}

func (r *Rule) updateNotificationTime(db *gorm.DB) error {

	r.LastNotification = time.Now()
	var err error = db.Model(&Rule{}).Where("id = ?", r.ID).Updates(&r).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *Rule) SendNotificationEmail(lastData map[string]interface{}) {

	// email info
	var to []string
	to = append(to, r.Email)
	subject := r.EmailSubject
	msg := r.EmailBody

	// replace subject and msg with data values if exists
	subject = strings.Replace(subject, "$value", fmt.Sprintf("%v", lastData[r.Sensor]), -1)
	subject = strings.Replace(subject, "$collected_at", fmt.Sprintf("%v", lastData["collected_at"]), -1)
	subject = strings.Replace(subject, "$device_id", fmt.Sprintf("%v", r.DeviceID), -1)
	subject = strings.Replace(subject, "$sensor", r.Sensor, -1)

	msg = strings.Replace(msg, "$value", fmt.Sprintf("%v", lastData[r.Sensor]), -1)
	msg = strings.Replace(msg, "$collected_at", fmt.Sprintf("%v", lastData["collected_at"]), -1)
	msg = strings.Replace(msg, "$device_id", fmt.Sprintf("%v", r.DeviceID), -1)
	msg = strings.Replace(msg, "$sensor", r.Sensor, -1)

	// Sender data
	from := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	// smtp server configuration
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: "+subject+" \n%s\n\n"+msg, mimeHeaders)))

	// Sending email.
	smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
}

func (r *Rule) sendRequestNotification(lastData map[string]interface{}) error {

	// json payload
	// replace values and collected_at if exists
	payload, _ := r.EndpointPayload.Value()
	stringPayload := fmt.Sprintf("%v", payload)

	newPayload := strings.Replace(stringPayload, "$value", fmt.Sprintf("%v", lastData[r.Sensor]), -1)
	newPayload = strings.Replace(newPayload, "$collected_at", fmt.Sprintf("%v", lastData["collected_at"]), -1)
	newPayload = strings.Replace(newPayload, "$device_id", fmt.Sprintf("%v", r.DeviceID), -1)
	newPayload = strings.Replace(newPayload, "$sensor", r.Sensor, -1)

	json.Unmarshal([]byte(newPayload), &r.EndpointPayload)

	// get the new payload
	json_data, _ := json.Marshal(r.EndpointPayload)

	client := http.Client{}
	req, err := http.NewRequest("POST", r.EndpointUrl, bytes.NewBuffer(json_data))

	if err != nil {
		return err
	}

	// headers
	headers, _ := json.Marshal(r.EndpointHeader)

	var h map[string]interface{}
	_ = json.Unmarshal(headers, &h)

	for key, value := range h {
		req.Header.Set(key, fmt.Sprintf("%v", value))
	}

	// send request
	_, errRequest := client.Do(req)
	if errRequest != nil {
		return errRequest
	}

	return nil
}

func (r *Rule) RuleValidations(db *gorm.DB, tenant_id uuid.UUID) formaterror.GeneralError {

	var errors formaterror.GeneralError

	if r.Value == "" {
		errors.Errors = append(errors.Errors, "value is required")

	}
	if len(r.Value) > 255 {
		errors.Errors = append(errors.Errors, "value is too long")

	}
	if len(r.Description) > 255 {
		errors.Errors = append(errors.Errors, "description is too long")
	}
	if r.Email != "" {
		if len(r.Email) > 255 {
			errors.Errors = append(errors.Errors, "email is too long")
		}
		if err := checkmail.ValidateFormat(r.Email); err != nil {
			errors.Errors = append(errors.Errors, "invalid email")
		}
	}
	if len(r.EmailSubject) > 255 {
		errors.Errors = append(errors.Errors, "email_subject is too long")
	}
	if len(r.EmailBody) > 255 {
		errors.Errors = append(errors.Errors, "email_body is too long")
	}
	if r.EndpointUrl != "" {
		if len(r.EndpointUrl) > 255 {
			errors.Errors = append(errors.Errors, "endpoint_url is too long")
		}
	}

	// validate sensor name
	var sensor Sensor
	if !sensor.IsValidSensorName(db, r.Sensor, r.DeviceID) {
		errors.Errors = append(errors.Errors, "invalid sensor name")
	}

	// validate device id
	if r.DeviceID.String() == "" {
		errors.Errors = append(errors.Errors, "device_id is required")
	}
	var device Device
	if !device.IsValidDevice(db, r.DeviceID, tenant_id) {
		errors.Errors = append(errors.Errors, "invalid device_id")
	}

	if r.TimeBetweenNotification != "" {
		if !isValidTimeBetweenNotification(r.TimeBetweenNotification) {
			errors.Errors = append(errors.Errors, "invalid time_between_notification")
		}
	}

	if r.CountLatest < 1 {
		errors.Errors = append(errors.Errors, "count_latest is required")

	} else if r.CountLatest > 1 {
		if r.Operator == "" {
			errors.Errors = append(errors.Errors, "operator is required")
		}
		if r.Operator != "lt" && r.Operator != "lte" && r.Operator != "gt" && r.Operator != "gte" && r.Operator != "eq" {
			errors.Errors = append(errors.Errors, "invalid operator. The available operators are: lt, lte, gt, gte and eq")
		}
		if r.Operation == "" {
			errors.Errors = append(errors.Errors, "operation is required")
		}
		if r.Operation != "sum" && r.Operation != "mean" && r.Operation != "median" && r.Operation != "max" && r.Operation != "min" {
			errors.Errors = append(errors.Errors, "invalid operation. The available operations are: sum, mean, median, max and min")
		}

	} else {
		if r.Operator != "" {
			if r.Operator != "lt" && r.Operator != "lte" && r.Operator != "gt" && r.Operator != "gte" && r.Operator != "eq" {
				errors.Errors = append(errors.Errors, "invalid operator. The available operators are: lt, lte, gt, gte and eq")
			}
		}
		if r.Operation != "" {
			if r.Operation != "sum" && r.Operation != "mean" && r.Operation != "median" && r.Operation != "max" && r.Operation != "min" {
				errors.Errors = append(errors.Errors, "invalid operation. The available operations are: sum, mean, median, max and min")
			}
		}
	}

	return errors
}

func (r *Rule) IsValidRule(db *gorm.DB, tenant_id uuid.UUID, rule_id uuid.UUID) (bool, error) {

	rules := []Rule{}

	// query
	err := db.Where("tenant_id = ? AND id = ?", tenant_id, rule_id).Find(&rules).Error
	if err != nil {
		return false, err
	}

	if len(rules) > 0 {
		return true, nil
	}

	return false, nil
}

func (r *Rule) SaveRule(db *gorm.DB, tenant_id uuid.UUID) (*Rule, error) {

	r.TenantID = tenant_id

	// create rule
	err := db.Model(&Rule{}).Create(&r).Error
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Rule) FindAllRules(db *gorm.DB, tenant_id string, req *http.Request) (interface{}, error) {

	rules := []Rule{}

	var count int

	var err_count error = db.Where("tenant_id = ?", tenant_id).Find(&rules).Count(&count).Error
	if err_count != nil {
		return nil, err_count
	}

	// pagination
	offset, limit, page, totalPages, nextPage, previousPage, errPagination := pagination.ValidatePagination(req, count)
	if errPagination != nil {
		return nil, errPagination
	}

	// query
	var err error = db.Where("tenant_id = ?", tenant_id).Limit(limit).Offset(offset).Order("updated_at desc").Find(&rules).Error
	if err != nil {
		return nil, err
	}

	return pagination.ListPaginationSerializer(limit, page, count, totalPages, nextPage, previousPage, rules), nil
}

func (r *Rule) GetRule(db *gorm.DB, rule_id string) (*Rule, error) {

	rule := Rule{}

	// query
	err := db.Model(&Rule{}).Where("id = ?", rule_id).Take(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *Rule) UpdateRule(db *gorm.DB, rule_id string) (*Rule, error) {

	var err error = db.Model(&Rule{}).Where("id = ?", rule_id).Updates(&r).Error

	if err != nil {
		return nil, err
	}

	// get the updated rule
	var err_get_rule error = db.Model(&Rule{}).Where("id = ?", rule_id).Take(&r).Error
	if err_get_rule != nil {
		return nil, err_get_rule
	}
	return r, nil
}

func (e *Rule) DeleteRule(db *gorm.DB, rule_id string) error {

	var err error = db.Where("id = ?", rule_id).Delete(&Rule{}).Error

	if err != nil {
		return err
	}
	return nil
}

func CheckRule(dbm *mongo.Client, db *gorm.DB, device_id uuid.UUID, lastData map[string]interface{}) {

	var rules []Rule

	db.Where("device_id = ?", device_id).Find(&rules)

	for i := 0; i < len(rules); i++ {

		if rules[i].Status == "active" {

			// only one device value
			if rules[i].CountLatest < 2 {
				onlyOneDeviceValue(dbm, db, rules[i], device_id, lastData)

				// N latest device data
			} else if rules[i].CountLatest > 1 {
				latestDeviceData(dbm, db, rules[i], device_id, lastData)
			}
		}
	}
}

func onlyOneDeviceValue(dbm *mongo.Client, db *gorm.DB, rule Rule, device_id uuid.UUID, lastData map[string]interface{}) error {

	// check notification time
	if !freeNotificationTime(rule.TimeBetweenNotification, rule.LastNotification) {
		return nil
	}

	// filter params
	var opt options.FindOptions
	opt.SetLimit(1)
	opt.SetSort(bson.M{"$natural": -1})

	// set filters to mongodb
	filter := bson.D{{rule.Sensor, bson.D{{"$exists", true}}}}

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

		if rule.Operator == "gt" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Sensor]), 64); err == nil {
					if deviceValue > ruleValue {
						rule.updateNotificationTime(db)
						if rule.Email != "" {
							rule.SendNotificationEmail(lastData)
						}
						if rule.EndpointUrl != "" {
							rule.sendRequestNotification(lastData)
						}
					}
				}
			}

		} else if rule.Operator == "gte" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Sensor]), 64); err == nil {
					if deviceValue >= ruleValue {
						rule.updateNotificationTime(db)
						if rule.Email != "" {
							rule.SendNotificationEmail(lastData)
						}
						if rule.EndpointUrl != "" {
							rule.sendRequestNotification(lastData)
						}
					}
				}
			}
		} else if rule.Operator == "lt" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Sensor]), 64); err == nil {
					if deviceValue < ruleValue {
						rule.updateNotificationTime(db)
						if rule.Email != "" {
							rule.SendNotificationEmail(lastData)
						}
						if rule.EndpointUrl != "" {
							rule.sendRequestNotification(lastData)
						}
					}
				}
			}
		} else if rule.Operator == "lte" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Sensor]), 64); err == nil {
					if deviceValue <= ruleValue {
						rule.updateNotificationTime(db)
						if rule.Email != "" {
							rule.SendNotificationEmail(lastData)
						}
						if rule.EndpointUrl != "" {
							rule.sendRequestNotification(lastData)
						}
					}
				}
			}
		} else {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Data[0][rule.Sensor]), 64); err == nil {
					if deviceValue == ruleValue {
						rule.updateNotificationTime(db)
						if rule.Email != "" {
							rule.SendNotificationEmail(lastData)
						}
						if rule.EndpointUrl != "" {
							rule.sendRequestNotification(lastData)
						}
					}
				} else {
					if fmt.Sprintf("%v", data.Data[0][rule.Sensor]) == rule.Value {
						rule.updateNotificationTime(db)
						if rule.Email != "" {
							rule.SendNotificationEmail(lastData)
						}
						if rule.EndpointUrl != "" {
							rule.sendRequestNotification(lastData)
						}
					}
				}
			} else {
				if fmt.Sprintf("%v", data.Data[0][rule.Sensor]) == rule.Value {
					rule.updateNotificationTime(db)
					if rule.Email != "" {
						rule.SendNotificationEmail(lastData)
					}
					if rule.EndpointUrl != "" {
						rule.sendRequestNotification(lastData)
					}
				}
			}

		}
	}

	return nil
}

func latestDeviceData(dbm *mongo.Client, db *gorm.DB, rule Rule, device_id uuid.UUID, lastData map[string]interface{}) error {

	// check notification time
	if !freeNotificationTime(rule.TimeBetweenNotification, rule.LastNotification) {
		return nil
	}

	// filter params
	var opt options.FindOptions
	opt.SetLimit(rule.CountLatest)
	opt.SetSort(bson.M{"$natural": -1})

	// set filters to mongodb
	filter := bson.D{{rule.Sensor, bson.D{{"$exists", true}}}}

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

				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Sensor]), 64); err == nil {
					calculatedValue = calculatedValue + deviceValue
				}

			}

		} else if rule.Operation == "mean" {

			var auxValue float64

			for _, value := range data.Data {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Sensor]), 64); err == nil {
					auxValue = auxValue + deviceValue
				}
			}

			calculatedValue = auxValue / float64(len(data.Data))

		} else if rule.Operation == "median" {
			var auxValues []float64

			for _, value := range data.Data {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Sensor]), 64); err == nil {
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
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Sensor]), 64); err == nil {
					if deviceValue > auxValue {
						auxValue = deviceValue
					}
				}
			}

			calculatedValue = auxValue

		} else if rule.Operation == "min" {
			var auxValue float64

			for _, value := range data.Data {
				if deviceValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value[rule.Sensor]), 64); err == nil {
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
		if rule.Operator == "gt" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue > ruleValue {
					rule.updateNotificationTime(db)
					if rule.Email != "" {
						rule.SendNotificationEmail(lastData)
					}
					if rule.EndpointUrl != "" {
						rule.sendRequestNotification(lastData)
					}
				}
			}

		} else if rule.Operator == "gte" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue >= ruleValue {
					rule.updateNotificationTime(db)
					if rule.Email != "" {
						rule.SendNotificationEmail(lastData)
					}
					if rule.EndpointUrl != "" {
						rule.sendRequestNotification(lastData)
					}
				}
			}
		} else if rule.Operator == "lt" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue < ruleValue {
					rule.updateNotificationTime(db)
					if rule.Email != "" {
						rule.SendNotificationEmail(lastData)
					}
					if rule.EndpointUrl != "" {
						rule.sendRequestNotification(lastData)
					}
				}
			}
		} else if rule.Operator == "lte" {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue <= ruleValue {
					rule.updateNotificationTime(db)
					if rule.Email != "" {
						rule.SendNotificationEmail(lastData)
					}
					if rule.EndpointUrl != "" {
						rule.sendRequestNotification(lastData)
					}
				}
			}
		} else {
			if ruleValue, err := strconv.ParseFloat(rule.Value, 64); err == nil {
				if calculatedValue == ruleValue {
					rule.updateNotificationTime(db)
					if rule.Email != "" {
						rule.SendNotificationEmail(lastData)
					}
					if rule.EndpointUrl != "" {
						rule.sendRequestNotification(lastData)
					}
				}
			}

		}

	}

	return nil
}

func isValidTimeBetweenNotification(timeBetween string) bool {

	isValid := false

	possibleTimeBetween := []string{"s", "m", "h"}

	for _, value := range possibleTimeBetween {
		if strings.Contains(timeBetween, value) {

			t := strings.Split(timeBetween, value)

			if len(t[1]) > 1 || len(t) > 2 || t[len(t)-1] != "" {
				return false
			}

			// convert time
			_, err := strconv.ParseInt(t[0], 10, 64)
			if err == nil {
				isValid = true

			} else {
				return false
			}
		}
	}
	return isValid
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
