package v1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Credentials struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

func Login(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.Login")
	var creds Credentials
	err := c.BodyParser(&creds)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	is_secure, entropy, err := helpers.IsPasswordSecure(creds.Password, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("InsecurePassword(%v)", entropy), err, config.INSERCURE_PWD))
	}
	if !helpers.IsEmailValid(creds.Email) {
		return c.JSON(errRes("IsEmailValid()", errors.New("Email not valid."), config.NOT_ALLOWED))
	}
	employees := config.MI.DB.Collection(config.EMPLOYEES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	employeeCursor, err := employees.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"email":     creds.Email,
				"exited_on": nil,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"full_name": bson.M{"$concat": bson.A{"$name", " ", "$surname"}},
			},
		},
	})

	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer employeeCursor.Close(ctx)
	var employee models.EmployeeForLoginWithPassword
	if employeeCursor.Next(ctx) {
		err = employeeCursor.Decode(&employee)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		if err = helpers.ComparePassword(employee.Password, creds.Password); err != nil {
			return c.JSON(errRes("ComparePassword()", errors.New("Wrong password"), config.CREDENTIALS_ERROR))
		}
		// valid user, then create token
		ttl, err := time.ParseDuration(os.Getenv(config.ACCT_EXPIREDIN))
		if err != nil {
			return c.JSON(errRes("ParseDuration(acc_token_exp_in)", err, config.ACCT_TTL_NOT_VALID))
		}
		access_token, err := helpers.CreateToken(ttl, employee.WithoutPassword(), os.Getenv(config.ACCT_PRIVATE_KEY))
		if err != nil {
			return c.JSON(errRes("CreateToken(access_token)", err, config.ACCT_GENERATION_ERROR))
		}
		ttl, err = time.ParseDuration(os.Getenv(config.REFT_EXPIREDIN))
		if err != nil {
			return c.JSON(errRes("ParseDuration(ref_token_exp_in)", err, config.REFT_TTL_NOT_VALID))
		}
		refresh_token, err := helpers.CreateToken(ttl, employee.Id, os.Getenv(config.REFT_PRIVATE_KEY))
		if err != nil {
			return c.JSON(errRes("CreateToken(refresh_token)", err, config.REFT_GENERATION_ERROR))
		}

		return c.JSON(models.Response[fiber.Map]{
			IsSuccess: true,
			Result: fiber.Map{
				"access_token":  access_token,
				"refresh_token": refresh_token,
				"user":          employee.WithoutPassword(),
			},
		})
	}
	if err = employeeCursor.Err(); err != nil {
		return c.JSON(errRes("employeeCursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(errRes("employeeCursor.Next()", errors.New("cursor next() error"), config.SERVER_ERROR))
}
func StopWorking(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.StopWorking")
	var employeeObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	everydayWorkColl := config.MI.DB.Collection(config.EVERYDAYWORK)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	var empTodaysWork struct {
		Id          primitive.ObjectID `bson:"_id"`
		ChecksCount int64              `bson:"checks_count"`
	}

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"employee_id": employeeObjId,
				"date":        today,
			},
		},
		bson.M{
			"$limit": 1,
		},
		bson.M{
			"$project": bson.M{
				// "work_time":    1,
				"checks_count": 1,
				// "text":         1,
			},
		},
	}
	cursor, err := everydayWorkColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		err = cursor.Decode(&empTodaysWork)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if empTodaysWork.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Current employee has no data for today."), config.SERVER_ERROR))
	}
	text := fmt.Sprintf("%v checks", empTodaysWork.ChecksCount)
	if c.Locals(config.EMPLOYEE_JOB) == config.EMPLOYEES_MANAGER {
		var empManagerWork struct {
			Work string `json:"text"`
		}
		err = c.BodyParser(&empManagerWork)
		if err != nil {
			return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
		} else {
			text = empManagerWork.Work
		}
	}
	stop_working := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	_, err = everydayWorkColl.UpdateOne(ctx, bson.M{"_id": empTodaysWork.Id}, bson.M{
		"$set": bson.M{
			"work_time.end": stop_working,
			"text":          text,
		},
	})
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.UPDATED,
	})
}
func StartWorking(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.StartWorking")
	var employeeObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	everydayWorkColl := config.MI.DB.Collection(config.EVERYDAYWORK)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now() // Asia/Ashgabat timezone-a gora save edyas!!!
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	// eger su gun irden hem start_working eden bolsa, onda ony tapyp update et, bolmasa-da taze doret
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"employee_id": employeeObjId,
				"date":        today,
			},
		},
		bson.M{
			"$limit": 1,
		},
		bson.M{
			"$project": bson.M{
				// "work_time": 1,
				"_id": 1,
			},
		},
	}
	cursor, err := everydayWorkColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var empTodaysWork struct {
		Id primitive.ObjectID `bson:"_id"`
		// WorkTime models.WorkTime    `bson:"work_time"`
	}
	for cursor.Next(ctx) {
		err = cursor.Decode(&empTodaysWork)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		break
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if empTodaysWork.Id == primitive.NilObjectID { // diymek on start_working etmandir!
		hour := now.Hour()
		minute := now.Minute()
		starting_time := fmt.Sprintf("%02d:%02d", hour, minute)
		st := models.EverydayWork{
			EmployeeId:     employeeObjId,
			Date:           today,
			ChecksCount:    0,
			Auctions:       []primitive.ObjectID{},
			Notifications:  []primitive.ObjectID{},
			Posts:          []primitive.ObjectID{},
			Products:       []primitive.ObjectID{},
			SellerProfiles: []primitive.ObjectID{},
			WorkTime:       models.WorkTime{Start: starting_time, End: nil},
			Text:           "",
		}
		// fmt.Printf("\nstart working: %v\n", st)
		// fmt.Printf("\ntimezone info: %v\n", st.Date.Location())
		_, err = everydayWorkColl.InsertOne(ctx, st)
		if err != nil {
			return c.JSON(errRes("InsertOne()", err, config.CANT_INSERT))
		}
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.CREATED,
		})
	} else {
		_, err = everydayWorkColl.UpdateOne(ctx, bson.M{"_id": empTodaysWork.Id}, bson.M{
			"$set": bson.M{
				"work_time.end": nil,
				"text":          "",
			},
		})
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.UPDATED,
		})
	}
}

func RefreshAccessToken(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.RefreshAccessToken:")
	token, err := helpers.GetTokenFromHeader(c)
	if err != nil {
		// fmt.Println("refresh token error: ", err.Error())
		return c.Status(401).JSON(errRes("GetTokenFromHeader()", err, config.REFT_NOT_FOUND))
	}
	sub, err := helpers.ValidateToken(token, os.Getenv("REFRESH_TOKEN_PUBLIC_KEY"))
	if err != nil {
		return c.Status(401).JSON(errRes("ValidateToken()", err, config.REFT_EXPIRED))
	}
	employeeId, err := primitive.ObjectIDFromHex(sub.(string))
	if err != nil {
		return c.Status(401).JSON(errRes("ObjectIDFromHex()", err, config.CANT_DECODE))
	}
	employees := config.MI.DB.Collection(config.EMPLOYEES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := employees.FindOne(ctx, bson.M{
		"_id": employeeId,
	})
	if err = user.Err(); err != nil {
		return c.Status(401).JSON(errRes("FindOne()", err, config.NOT_FOUND))
	}
	var employee models.EmployeeForLoginWithoutPassword
	err = user.Decode(&employee)
	if err != nil {
		return c.Status(401).JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	ttl, err := time.ParseDuration(os.Getenv(config.ACCT_EXPIREDIN))
	if err != nil {
		return c.Status(401).JSON(errRes("ParseDuration(access_token)", err, ""))
	}
	access_token, err := helpers.CreateToken(ttl, employee, os.Getenv(config.ACCT_PRIVATE_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("CreateToken(access_token)", err, ""))
	}
	ttl, err = time.ParseDuration(os.Getenv(config.REFT_EXPIREDIN))
	if err != nil {
		return c.Status(401).JSON(errRes("ParseDuration(refresh_token)", err, ""))
	}
	refresh_token, err := helpers.CreateToken(ttl, employee.Id, os.Getenv(config.REFT_PRIVATE_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("CreateToken(refresh_token)", err, ""))
	}

	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"access_token":  access_token,
			"refresh_token": refresh_token,
			// "user":          employee,
		},
	})
}

// func DummyCookie(c *fiber.Ctx) error {

// 	c.Cookie(&fiber.Cookie{
// 		Name:    "koke",
// 		Value:   "koke diymek cookie diymek",
// 		Expires: time.Now().Add(5 * time.Minute),
// 		// HTTPOnly: false,
// 	})

// 	return c.JSON(models.Response[string]{
// 		IsSuccess: true,
// 		Result:    "cookie",
// 	})
// }
func ChangePassword(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.ChangePassword")
	var employeeObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	var passwords struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	err = c.BodyParser(&passwords)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	is_secure, entropy, err := helpers.IsPasswordSecure(passwords.OldPassword, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("OldPassword(%v)", entropy), err, config.INSERCURE_PWD))
	}
	is_secure, entropy, err = helpers.IsPasswordSecure(passwords.NewPassword, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("NewPassword(%v)", entropy), err, config.INSERCURE_PWD))
	}

	hashed_new := helpers.HashPassword(passwords.NewPassword)
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customerCursor, err := employeesColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id":       employeeObjId,
				"exited_on": nil,
			},
		},
		bson.M{
			"$project": bson.M{
				"password": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}

	var customerResult struct {
		Password string `bson:"password"`
	}

	if customerCursor.Next(ctx) {
		err = customerCursor.Decode(&customerResult)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	err = helpers.ComparePassword(customerResult.Password, passwords.OldPassword)
	if err != nil {
		return c.JSON(errRes("ComparePassword()", err, config.CREDENTIALS_ERROR))
	}

	result, err := employeesColl.UpdateOne(ctx, bson.M{
		"_id":       employeeObjId,
		"exited_on": nil,
	}, bson.M{
		"$set": bson.M{
			"password": hashed_new,
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	if result.ModifiedCount != 1 {
		return c.JSON(errRes("UpdateOne()", err, config.NOT_FOUND))
	}

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.UPDATED,
	})
}
