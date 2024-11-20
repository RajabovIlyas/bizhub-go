package v1

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	everydayworkservice "github.com/devzatruk/bizhubBackend/everydaywork_service"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UploadEmployeesChatFile(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UploadEmployeesChatFile")
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(errRes("file.err", err, config.BODY_NOT_PROVIDED))
	}

	// ext := strings.Split(file.Header["Content-Type"][0], "/")[1]
	rootPath := config.RootPath

	err = c.SaveFile(file, path.Join(rootPath, fmt.Sprintf("public/files/chat/%v", file.Filename)))

	if err != nil {
		return c.JSON(errRes("c.SaveFile()", err, config.SERVER_ERROR))
	}

	type EmployeesChatFileUploadResponse struct {
		MimeType string `json:"mimetype"`
		Name     string `json:"name"`
		Size     int64  `json:"size"`
	}

	result := EmployeesChatFileUploadResponse{
		Size:     file.Size,
		MimeType: file.Header["Content-Type"][0],
		Name:     file.Filename,
	}

	return c.JSON(models.Response[EmployeesChatFileUploadResponse]{
		IsSuccess: true,
		Result:    result,
	})
}

func GivePermission(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GivePermission")
	var managerObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &managerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	now := time.Now()
	// y, m, d := now.Date()
	// today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var feData models.ReasonFromFE
	err = c.BodyParser(&feData)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	var data models.Reason
	data.EmployeeId = employeeObjId
	data.CreatedBy = managerObjId
	data.CreatedAt = now
	newFrom, err := helpers.StringToDate(feData.From)
	if err != nil {
		c.JSON(errRes("StringToDate(from)", err, config.CANT_DECODE))
	}
	data.From = newFrom
	data.To = newFrom.AddDate(0, 0, feData.Days)
	data.Days = feData.Days
	data.Name = feData.Name
	data.DisplayName = feData.DisplayName
	data.Description = feData.Description
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_reasonsColl := transaction_manager.Collection(config.REASONS)
	insert_model := ojoTr.NewModel().SetDocument(data)
	insertResult, err := tr_reasonsColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(reason)", err, config.CANT_INSERT))
	}
	// on berilip, yatdan cykan rugsat bolsa
	if data.To.Before(now) {
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.CREATED,
		})
	}
	newReasonId := insertResult.InsertedID.(primitive.ObjectID)
	var batch []*ojocronservice.OjoCronJobModel
	// on berilip, hazirem dowam edyan rugsat bolsa
	if data.From.Before(now) && data.To.After(now) {
		// ilki employee.reason = set et
		update_model := ojoTr.NewModel().
			SetFilter(bson.M{"_id": employeeObjId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"reason": bson.M{
						"name":         data.Name,
						"display_name": data.DisplayName,
						"_id":          newReasonId,
					},
				},
			}).
			SetRollbackUpdate(bson.M{
				"$set": bson.M{
					"reason": nil,
				},
			})
		tr_employeesColl := transaction_manager.Collection(config.EMPLOYEES)
		_, err = tr_employeesColl.FindOneAndUpdate(update_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(employee)", err, config.CANT_UPDATE))
		}
		// cron job doret ki To yetende employee.reason = null bolsun
		modelF := ojocronservice.NewOjoCronJobModel()
		modelF.ListenerName(config.PERMISSION_ENDED)
		modelF.Payload(map[string]interface{}{"employee_id": employeeObjId})
		modelF.RunAt(data.To)
		batch = append(batch, modelF)
	} else if data.From.After(now) { // ertire rugsat beryan bolsa
		// cron job From yetende employee.reason = {...} we To yetende employee.reason = null bolsun
		modelF := ojocronservice.NewOjoCronJobModel()
		modelF.ListenerName(config.PERMISSION_STARTED)
		modelF.Payload(map[string]interface{}{
			"employee_id": employeeObjId,
			"reason": bson.M{
				"name":         data.Name,
				"display_name": data.DisplayName,
				"_id":          newReasonId,
			},
		})
		modelF.RunAt(data.From)
		batch = append(batch, modelF)

		modelF = ojocronservice.NewOjoCronJobModel()
		modelF.ListenerName(config.PERMISSION_ENDED)
		modelF.Payload(map[string]interface{}{"employee_id": employeeObjId})
		modelF.RunAt(data.To)
		batch = append(batch, modelF)
	}

	if err = transaction_manager.Err(); err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("transaction_manager.Err()", err, config.TRANSACTION_FAILED))
	}
	for _, jobModel := range batch {
		config.OjoCronService.NewJob(jobModel)
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.TRANSACTION_SUCCESSFUL,
	})
}
func AddNote(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.AddNote")
	var managerObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &managerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var note models.Note
	err = c.BodyParser(&note)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	note.CreatedBy = managerObjId
	note.EmployeeId = employeeObjId
	note.CreatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	notesColl := config.MI.DB.Collection(config.NOTES)
	result, err := notesColl.InsertOne(ctx, note)
	if err != nil {
		return c.JSON(errRes("InsertOne()", err, config.CANT_INSERT))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    fmt.Sprintf("Added a new note with ID: %v", result.InsertedID),
	})
}

func AddNewEmployee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.AddNewEmployee")
	var employeeObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	now := time.Now()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	var emp models.NewEmployee
	emp.Name = c.FormValue("name")
	emp.Surname = c.FormValue("surname")
	emp.MiddleName = c.FormValue("middle_name")
	birthDate, err := helpers.StringToDate(c.FormValue("birth_date"))
	// fmt.Printf("birth date: %v - parsed: %v", c.FormValue("birth_date"), birthDate)
	if err != nil {
		return c.JSON(errRes("StringToDate(birth_date)", err, config.CANT_DECODE))
	}
	emp.BirthDate = birthDate
	emp.Address = c.FormValue("address")
	emp.PassportCode = c.FormValue("passport_code")
	emp.PassportCopies = make([]string, 0)
	emp.Email = c.FormValue("email")
	if !helpers.IsEmailValid(emp.Email) {
		return c.JSON(errRes("IsEmailValid()", errors.New("Email not valid."), config.NOT_ALLOWED))
	}
	emp.Password = c.FormValue("password")
	is_secure, entropy, err := helpers.IsPasswordSecure(emp.Password, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("InsecurePassword(%v)", entropy), err, config.INSERCURE_PWD))
	}
	emp.Password = helpers.HashPassword(emp.Password)

	emp.Job.Name = c.FormValue("job.name")
	emp.Job.DisplayName = c.FormValue("job.display_name")
	salary, err := strconv.ParseFloat(c.FormValue("salary"), 64)
	if err != nil {
		salary = 0
	}
	emp.Salaries = []models.Salary{
		{
			Amount: salary,
			From:   today,
			To:     nil,
		},
	}
	emp.ExitedOn = nil
	workTimeEnd := c.FormValue("end")
	emp.WorkTime = models.WorkTime{
		Start: c.FormValue("start"),
		End:   &workTimeEnd,
	}
	emp.StartedOn = []models.StartedOn{
		{
			From: today,
			To:   nil,
		},
	}
	emp.Reason = nil
	emp.CreatedBy = employeeObjId
	// passport copyalary almaga calysaly
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
	}
	unimportant_errors := []string{}
	if copies, ok := form.File["passport_copies"]; ok && len(copies) > 0 {
		for _, copy := range copies {
			imagePath, err := helpers.SaveFileheader(c, copy, config.FOLDER_EMPLOYEE_PASSPORTS)
			if err != nil {
				unimportant_errors = append(unimportant_errors, fmt.Sprintf("Couldn't save a passport copy: %v", err))
			} else {
				emp.PassportCopies = append(emp.PassportCopies, imagePath)
			}
		}
		if len(emp.PassportCopies) == 0 {
			unimportant_errors = append(unimportant_errors, fmt.Sprintf("Attemp to save passport copies failed."))
		}
	} // else {
	// fmt.Printf("\nNo passport copies uploaded.\n")
	// }
	if avatars, ok := form.File["avatar"]; ok && len(avatars) > 0 {
		avatar := avatars[0]
		imagePath, err := helpers.SaveFileheader(c, avatar, config.FOLDER_EMPLOYEE_AVATARS)
		if err != nil {
			unimportant_errors = append(unimportant_errors, fmt.Sprintf("Couldn't save avatar image: %v", err))
		} else {
			emp.Avatar = imagePath
		}
	} //else {
	// fmt.Printf("\nAvatar not provided.\n")
	// }
	if emp.HasEmptyFields() {
		helpers.DeleteImages(emp.PassportCopies)
		helpers.DeleteImageFile(emp.Avatar)
		return c.JSON(errRes("HasEmptyFields()", errors.New("Some data not provided."), config.BODY_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	insertResult, err := employeesColl.InsertOne(ctx, emp)
	if err != nil {
		helpers.DeleteImageFile(emp.Avatar)
		helpers.DeleteImages(emp.PassportCopies)
		return c.JSON(errRes("InsertOne(employee)", err, config.CANT_INSERT))
	}
	emp.Id = insertResult.InsertedID.(primitive.ObjectID)
	return c.JSON(models.Response[any]{
		IsSuccess: true,
		Result: fiber.Map{
			"new_employee":       emp,
			"unimportant_errors": unimportant_errors, // save edip bilmedik passport_copy ya avatar bar bolsa, solaryn listesi
		},
	})
}

// hem employee info-ny EDIT etmek hem-de RECRUIT etmek ucin!!!
func EditEmployeeInfo(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.EditEmployeeInfo")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	is_recruit := false
	if c.Query("recruit") == "true" {
		is_recruit = true
	}
	// edit edilyan employee ve edit edyan employee bar!!!
	/* request payload:
	{
		"new_name": "maysa",
		"new_surname":"godekowa",
		"new_middle_name":"kabayewna",
		"new_birth_date":"2000-09-12",
		"new_address":"cingizhan kocesi 13",
		"new_passport_copy":"I-LB123456",
		"deleted_passport_copies": ["image1.jpg", "image2.jpg"]string,
		"new_passport_copies": []Image,
		"new_avatar": "avatar.png"Image,
		"new_job": {
			"name": "new job name",
			"display_name": "new job display name"
		},
		"new_login": "maysayka@godek.com",
		"new_password": "maysayka",
		"new_salary": 2345,

			"new_worktime_start": "10:00",
			"new_worktime_end": "15:00"
		}
	}
	*/
	var oldData models.NewEmployee
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	findResult := employeesColl.FindOne(ctx, bson.M{"_id": employeeObjId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne()", err, config.NOT_FOUND))
	}
	err = findResult.Decode(&oldData)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	now := time.Now()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	var newData = bson.M{}
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
	}
	if len(c.FormValue("new_name")) > 0 {
		newData["name"] = c.FormValue("new_name")
	}
	if len(c.FormValue("new_surname")) > 0 {
		newData["surname"] = c.FormValue("new_surname")
	}
	if len(c.FormValue("new_middle_name")) > 0 {
		newData["middle_name"] = c.FormValue("new_middle_name")
	}
	if len(c.FormValue("new_birth_date")) > 0 {
		birthDate, err := helpers.StringToDate(c.FormValue("new_birth_date"))
		if err != nil {
			return c.JSON(errRes("StringToDate(bith_date)", err, config.CANT_DECODE))
		}
		newData["birth_date"] = birthDate
	}
	if len(c.FormValue("new_address")) > 0 {
		newData["address"] = c.FormValue("new_address")
	}
	if len(c.FormValue("new_passport_code")) > 0 {
		newData["passport_code"] = c.FormValue("new_passport_code")
	}
	if len(c.FormValue("new_email")) > 0 {
		newData["email"] = c.FormValue("new_email")
		if !helpers.IsEmailValid(c.FormValue("new_email")) {
			return c.JSON(errRes("IsEmailValid()", errors.New("Email not valid."), config.NOT_ALLOWED))
		}
	}
	if len(c.FormValue("new_password")) > 0 {
		new_password := c.FormValue("new_password")
		is_secure, entropy, err := helpers.IsPasswordSecure(new_password, config.MIN_PWD_ENTROPY)
		if is_secure == false {
			return c.JSON(errRes(fmt.Sprintf("NewPassword(%v)", entropy), err, config.INSERCURE_PWD))
		}
		newData["password"] = helpers.HashPassword(new_password)
	}
	if len(c.FormValue("new_job_name")) > 0 && len(c.FormValue("new_job_display_name")) > 0 {
		newData["job"] = bson.M{
			"name":         c.FormValue("new_job_name"),
			"display_name": c.FormValue("new_job_display_name"),
		}
	}
	if len(c.FormValue("new_salary")) > 0 {
		salary, err := strconv.ParseFloat(c.FormValue("new_salary"), 64)
		if err != nil {
			return c.JSON(errRes("ParseFloat(salary)", err, config.CANT_DECODE))
		}
		// in sonky salary-ni tap, update et, taze gos
		newSalaries := []models.Salary{}
		for _, s := range oldData.Salaries {
			if s.To == nil {
				s.To = &today
			}
			newSalaries = append(newSalaries, s)
		}
		newSalaries = append(newSalaries, models.Salary{
			Amount: salary,
			From:   today,
			To:     nil,
		})
		newData["salaries"] = newSalaries
		if is_recruit {
			oldData.StartedOn = append(oldData.StartedOn, models.StartedOn{
				From: today,
				To:   nil,
			})
			newData["started_on"] = oldData.StartedOn
		}
	}
	recruitData := bson.M{}
	if is_recruit && len(c.FormValue("new_salary")) == 0 {
		recruitData =
			bson.M{
				"salaries": bson.M{"$each": bson.A{models.Salary{
					Amount: oldData.Salaries[len(oldData.Salaries)-1].Amount,
					From:   today,
					To:     nil,
				}}},
				"started_on": bson.M{"$each": bson.A{models.StartedOn{
					From: today,
					To:   nil,
				}}},
			}
	}
	if is_recruit {
		newData["exited_on"] = nil
	}
	newWorkTime := bson.M{}
	if len(c.FormValue("new_worktime_start")) > 0 {
		start := c.FormValue("new_worktime_start")
		newWorkTime["start"] = start
		if len(c.FormValue("new_worktime_end")) == 0 {
			newWorkTime["end"] = oldData.WorkTime.End
		}
	}
	if len(c.FormValue("new_worktime_end")) > 0 {
		end := c.FormValue("new_worktime_end")
		newWorkTime["end"] = end
		if len(c.FormValue("new_worktime_start")) == 0 {
			newWorkTime["start"] = oldData.WorkTime.Start
		}
	}
	if len(newWorkTime) > 0 {
		newData["work_time"] = newWorkTime
	}
	newPassportCopies := []string{}
	remainingPassportCopies := []string{}
	deletedPCopies, ok := form.Value["deleted_passport_copies"]
	if ok && len(deletedPCopies) > 0 {
		for _, copy := range oldData.PassportCopies {
			if !helpers.SliceContains(deletedPCopies, copy) {
				remainingPassportCopies = append(remainingPassportCopies, copy)
			}
		}
	} else {
		deletedPCopies = []string{}
		remainingPassportCopies = oldData.PassportCopies
	}

	if copies, ok := form.File["new_passport_copies"]; ok && len(copies) > 0 {
		for _, copy := range copies {
			imagePath, err := helpers.SaveFileheader(c, copy, config.FOLDER_EMPLOYEE_PASSPORTS)
			if err != nil {
				if len(newPassportCopies) > 0 {
					helpers.DeleteImages(newPassportCopies)
				}
				return c.JSON(errRes("SaveFileheader(passport_copy)", err, config.CANT_DECODE))
			} else {
				remainingPassportCopies = append(remainingPassportCopies, imagePath)
				newPassportCopies = append(newPassportCopies, imagePath)
			}
		}
	}
	if len(newPassportCopies) > 0 || len(deletedPCopies) > 0 {
		newData["passport_copies"] = remainingPassportCopies
	}
	if avatars, ok := form.File["new_avatar"]; ok && len(avatars) > 0 {
		avatar := avatars[0]
		imagePath, err := helpers.SaveFileheader(c, avatar, config.FOLDER_EMPLOYEE_AVATARS)
		if err != nil {
			if len(newPassportCopies) > 0 {
				helpers.DeleteImages(newPassportCopies)
			}
			return c.JSON(errRes("SaveFileheader(avatar)", err, config.CANT_DECODE))
		} else {
			newData["avatar"] = imagePath
		}
	}
	update_model := bson.M{"$set": newData}
	if len(recruitData) > 0 {
		update_model["$push"] = recruitData
	}
	updateResult, err := employeesColl.UpdateOne(ctx, bson.M{"_id": employeeObjId}, update_model)
	if err != nil || updateResult.MatchedCount == 0 {
		if len(newData["avatar"].(string)) > 0 {
			helpers.DeleteImageFile(newData["avatar"].(string))
		}
		if len(newPassportCopies) > 0 {
			helpers.DeleteImages(newPassportCopies)
		}
		if err != nil {
			return c.JSON(errRes("UpdateOne(employee)", err, config.CANT_UPDATE))
		}
		if updateResult.MatchedCount == 0 {
			return c.JSON(errRes("UpdateOne(employee)", errors.New("Employee not found."), config.NOT_FOUND))
		}
	}
	// update success ise, delete etmeli file-lary delete et
	if len(deletedPCopies) > 0 {
		helpers.DeleteImages(deletedPCopies)
	}

	return c.JSON(models.Response[any]{
		IsSuccess: true,
		Result: fiber.Map{
			"avatar": newData["avatar"],
		},
	})
}
func GetAllEmployees(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetAllEmployees")
	var employeeObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	reason := c.Query("reason", "all")
	job := c.Query("job", "all")
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	aggregateSlice := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": bson.M{
					"$not": bson.M{
						"$eq": employeeObjId,
					},
				},
			},
		},
	}

	if reason != "all" {
		aggregateSlice = append(aggregateSlice, bson.M{
			"$match": bson.M{
				"reason.name": reason,
			},
		})
	}
	if job != "all" {
		aggregateSlice = append(aggregateSlice, bson.M{
			"$match": bson.M{
				"job.name": job,
			},
		})
	}

	if c.Locals(config.EMPLOYEE_JOB) == config.EMPLOYEES_MANAGER {
		aggregateSlice = append(aggregateSlice, bson.M{
			"$match": bson.M{
				"created_by": employeeObjId,
			},
		})
	}
	aggregateSlice = append(aggregateSlice, bson.A{
		bson.M{
			"$skip": limit * pageIndex,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"full_name": bson.M{
					"$concat": bson.A{"$name", " ", "$surname"},
				},
				"job":       1,
				"avatar":    1,
				"exited_on": 1,
				"reason":    1,
			},
		},
	}...)

	cursor, err := employeesColl.Aggregate(ctx, aggregateSlice)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var employees []models.EmployeeForAdmin
	for cursor.Next(ctx) {
		var emp models.EmployeeForAdmin
		err = cursor.Decode(&emp)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		employees = append(employees, emp)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if employees == nil {
		employees = make([]models.EmployeeForAdmin, 0)
	}
	return c.JSON(models.Response[[]models.EmployeeForAdmin]{
		IsSuccess: true,
		Result:    employees,
	})
}
func GetNotesOfEmployee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetNotesOfEmployee")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", errors.New("Employee Id not provided."), config.PARAM_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.CANT_DECODE))
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.CANT_DECODE))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	notesColl := config.MI.DB.Collection(config.NOTES)
	cursor, err := notesColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"employee_id": employeeObjId,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "employees",
				"localField":   "created_by",
				"foreignField": "_id",
				"as":           "created_by",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"full_name": bson.M{
								"$concat": bson.A{"$name", " ", "$surname"},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$created_by",
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var empNotes []models.NoteForEmpInfo
	for cursor.Next(ctx) {
		var note models.NoteForEmpInfo
		err = cursor.Decode(&note)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		empNotes = append(empNotes, note)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if empNotes == nil {
		empNotes = make([]models.NoteForEmpInfo, 0)
	}
	return c.JSON(models.Response[[]models.NoteForEmpInfo]{
		IsSuccess: true,
		Result:    empNotes,
	})
}

func GetReasonsOfEmployee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetReasonsOfEmployee")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", errors.New("Employee Id not provided."), config.PARAM_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	reasonsColl := config.MI.DB.Collection(config.REASONS)
	cursor, err := reasonsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"employee_id": employeeObjId,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"_id":          1,
				"name":         1,
				"display_name": 1,
				"from":         1,
				"to":           1,
				"description":  1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var empReasons []models.ReasonForEmpInfo
	for cursor.Next(ctx) {
		var reason models.ReasonForEmpInfo
		err = cursor.Decode(&reason)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		empReasons = append(empReasons, reason)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if empReasons == nil {
		empReasons = make([]models.ReasonForEmpInfo, 0)
	}
	return c.JSON(models.Response[[]models.ReasonForEmpInfo]{
		IsSuccess: true,
		Result:    empReasons,
	})
}

func GetEverydayWorkOfEmployee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEverydayWorkOfEmployee")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", errors.New("Employee Id not provided."), config.PARAM_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	empColl := config.MI.DB.Collection(config.EMPLOYEES)
	employeeResult := empColl.FindOne(ctx, bson.M{"_id": employeeObjId})
	if err = employeeResult.Err(); err != nil {
		return c.JSON(errRes("FindOne()", err, config.NOT_FOUND))
	}
	var employee struct {
		Id  primitive.ObjectID `bson:"_id"`
		Job models.Job         `bson:"job"`
	}
	err = employeeResult.Decode(&employee)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}

	service := config.EverydayWorkService.Of(employeeObjId)
	works, err := service.Reader.Works(pageIndex, limit)
	if err != nil {
		return c.JSON(errRes("EverydayWorkService.Works()", err, config.CANT_DECODE))
	}
	for i := 0; i < len(works); i++ {
		if employee.Job.Name == config.ADMIN_CHECKER {
			works[i].Note = fmt.Sprintf("%v checked", works[i].CompletedTasksCount)
		} else if employee.Job.Name == config.CASHIER {
			works[i].Note = fmt.Sprintf("%v requests", works[i].CompletedTasksCount)
		}
	}

	return c.JSON(models.Response[[]everydayworkservice.EverydayWorkWithoutTasks]{
		IsSuccess: true,
		Result:    works,
	})
}

// func GetEverydayWorkOfEmployee(c *fiber.Ctx) error {
// 	errRes := helpers.ErrorResponse("Admin.GetEverydayWorkOfEmployee")
// 	var employeeObjId primitive.ObjectID
// 	empId := c.Params("id", "noid")
// 	if empId == "noid" {
// 		return c.JSON(errRes("Params(id)", errors.New("Employee Id not provided."), config.NO_PERMISSION))
// 	}
// 	employeeObjId, err := primitive.ObjectIDFromHex(empId)
// 	if err != nil {
// 		return c.JSON(errRes("ObjectIDFromHex()", err, config.NO_PERMISSION))
// 	}
// 	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
// 	if err != nil {
// 		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
// 	}
// 	limit, err := strconv.Atoi(c.Query("limit", "10"))
// 	if err != nil {
// 		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
// 	}
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	empColl := config.MI.DB.Collection(config.EMPLOYEES)
// 	employeeResult := empColl.FindOne(ctx, bson.M{"_id": employeeObjId})
// 	if err = employeeResult.Err(); err != nil {
// 		return c.JSON(errRes("FindOne()", err, config.CANT_DECODE))
// 	}
// 	var employee struct {
// 		Id  primitive.ObjectID `bson:"_id"`
// 		Job models.Job         `bson:"job"`
// 	}
// 	err = employeeResult.Decode(&employee)
// 	if err != nil {
// 		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
// 	}
// 	isAdminChecker := false
// 	if employee.Job.Name == config.ADMIN_CHECKER {
// 		isAdminChecker = true
// 	}
// 	everyday_workColl := config.MI.DB.Collection(config.EVERYDAYWORK)
// 	cursor, err := everyday_workColl.Aggregate(ctx, bson.A{
// 		bson.M{
// 			"$match": bson.M{
// 				"employee_id": employeeObjId,
// 			},
// 		},
// 		bson.M{
// 			"$skip": pageIndex * limit,
// 		},
// 		bson.M{
// 			"$limit": limit,
// 		},
// 		bson.M{
// 			"$project": bson.M{
// 				"_id":          1,
// 				"date":         1,
// 				"work_time":    1,
// 				"text":         1,
// 				"checks_count": 1,
// 			},
// 		},
// 		bson.M{
// 			"$addFields": bson.M{
// 				"text": bson.M{"$cond": bson.A{isAdminChecker, bson.M{
// 					"$concat": bson.A{bson.M{"$toString": "$checks_count"}, " ", "checked"},
// 				}, "$text"}},
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return c.JSON(errRes("Aggregate()", err, ""))
// 	}
// 	defer cursor.Close(ctx)
// 	var empWorks []models.DailyWorkForEmpInfo
// 	for cursor.Next(ctx) {
// 		var reason models.DailyWorkForEmpInfo
// 		err = cursor.Decode(&reason)
// 		if err != nil {
// 			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
// 		}
// 		empWorks = append(empWorks, reason)
// 	}
// 	if err = cursor.Err(); err != nil {
// 		return c.JSON(errRes("cursor.Err()", err, config.CANT_DECODE))
// 	}
// 	if empWorks == nil {
// 		empWorks = make([]models.DailyWorkForEmpInfo, 0)
// 	}
// 	return c.JSON(models.Response[[]models.DailyWorkForEmpInfo]{
// 		IsSuccess: true,
// 		Result:    empWorks,
// 	})
// }

func GetEmployeeInfoForEditing(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEmployeeInfoForEditing")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": employeeObjId,
			},
		},
		bson.M{
			"$project": bson.M{
				"name":            1,
				"surname":         1,
				"middle_name":     1,
				"address":         1,
				"passport_code":   1,
				"passport_copies": 1,
				"email":           1,
				"job":             1,
				"salary": bson.M{
					"$last": "$salaries.amount",
				},
				"work_time":  1,
				"avatar":     1,
				"birth_date": 1,
			},
		},
	}
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	cursor, err := employeesColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var employeeInfo models.EmployeeInfoForEditing
	if cursor.Next(ctx) {
		err = cursor.Decode(&employeeInfo)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if employeeInfo.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Employee not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.EmployeeInfoForEditing]{
		IsSuccess: true,
		Result:    employeeInfo,
	})
}
func GetEverydayworkPostDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEverydayworkPostDetail")
	postObjId, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		return c.JSON(errRes("Params(postId)", err, config.PARAM_NOT_PROVIDED))
	}
	postsColl := config.MI.DB.Collection(config.POSTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": postObjId,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "related_products",
				"foreignField": "_id",
				"as":           "related_products",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"_id": 1,
							"image": bson.M{
								"$first": "$images",
							},
						},
					},
				},
			},
		},
	}
	cursor, err := postsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var postDetail models.PostDetailWithTranslationWithoutSeller
	if cursor.Next(ctx) {
		err = cursor.Decode(&postDetail)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	} else {
		return c.JSON(errRes("cursor.Next()", errors.New("Post not found."), config.NOT_FOUND))
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[models.PostDetailWithTranslationWithoutSeller]{
		IsSuccess: true,
		Result:    postDetail,
	})
}
func GetEverydayworkProductDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEverydayworkProductDetail")
	prodObjId, err := primitive.ObjectIDFromHex(c.Params("productId"))
	if err != nil {
		return c.JSON(errRes("Params(productId)", err, config.PARAM_NOT_PROVIDED))
	}
	products := config.MI.DB.Collection(config.PRODUCTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	aggregationArray := bson.A{

		bson.M{
			"$match": bson.M{
				"_id": prodObjId,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"localField":   "seller_id",
				"foreignField": "_id",
				"as":           "seller",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "cities",
							"localField":   "city_id",
							"foreignField": "_id",
							"as":           "city",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"name": "$name.en",
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$city",
						},
					},
					bson.M{
						"$project": bson.M{
							"name": 1,
							"type": 1,
							"city": 1,
							"logo": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "attributes",
				"localField":   "attrs.attr_id",
				"foreignField": "_id",
				"as":           "attrs_detail",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":        "$name.en",
							"units_array": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"attrs": bson.M{
					"$map": bson.M{
						"input": "$attrs",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$this",
								bson.M{
									"attr": bson.M{
										"$arrayElemAt": bson.A{
											"$attrs_detail",
											bson.M{
												"$indexOfArray": bson.A{"$attrs", "$$this"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"attrs": bson.M{
					"$map": bson.M{
						"input": "$attrs",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$this",
								bson.M{
									"unit": bson.M{
										"$arrayElemAt": bson.A{
											"$$this.attr.units_array",
											"$$this.unit_index",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "category_id",
				"foreignField": "_id",
				"as":           "category",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "categories",
							"localField":   "parent",
							"foreignField": "_id",
							"as":           "parent",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"name": "$name.en",
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$parent",
						},
					},
					bson.M{
						"$project": bson.M{
							"name":   "$name.en",
							"parent": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "brands",
				"localField":   "brand_id",
				"foreignField": "_id",
				"as":           "brand",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "brands",
							"localField":   "parent",
							"foreignField": "_id",
							"as":           "parent",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"name": 1,
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$parent",
						},
					},
					bson.M{
						"$project": bson.M{
							"name":   1,
							"parent": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$category",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$brand",
			},
		},
		bson.M{
			"$project": bson.M{
				// "seller":        1,
				"category":      1,
				"brand":         1,
				"attrs":         1,
				"heading":       1, // <-
				"more_details":  1, // <-
				"likes":         1,
				"viewed":        1,
				"status":        1,
				"discount":      1,
				"price":         1,
				"discount_data": 1,
				"brand_id":      1,
				"category_id":   1,
				"images":        1,
				"seller_id":     1,
			},
		},
	}
	cursor, err := products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var productDetail models.ProductDetailWithTranslationWithoutSeller
	if cursor.Next(ctx) {
		err = cursor.Decode(&productDetail)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	} else {
		return c.JSON(errRes("cursor.Next()", errors.New("Product not found."), config.NOT_FOUND))
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[models.ProductDetailWithTranslationWithoutSeller]{
		IsSuccess: true,
		Result:    productDetail,
	})

}
func GetEmployeeEverydayworkProducts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEmployeeEverydayworkProducts")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	workId, err := primitive.ObjectIDFromHex(c.Params("workId"))
	if err != nil {
		return c.JSON(errRes("Params(workId)", err, config.PARAM_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	service := config.EverydayWorkService.Of(employeeObjId)
	work := service.Reader.Work(workId)
	products, err := work.Products(page, limit)
	if err != nil {
		return c.JSON(errRes("Products()", err, config.CANT_DECODE))
	}

	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}
func GetEmployeeEverydayworkPosts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEmployeeEverydayworkPosts")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	workId, err := primitive.ObjectIDFromHex(c.Params("workId"))
	if err != nil {
		return c.JSON(errRes("Params(workId)", err, config.PARAM_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	service := config.EverydayWorkService.Of(employeeObjId)
	work := service.Reader.Work(workId)
	posts, err := work.Posts(page, limit)
	if err != nil {
		return c.JSON(errRes("Posts()", err, config.CANT_DECODE))
	}

	return c.JSON(models.Response[[]models.Post]{
		IsSuccess: true,
		Result:    posts,
	})
}
func GetEmployeeEverydayworkCashierActivities(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEmployeeEverydayworkPosts")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	workId, err := primitive.ObjectIDFromHex(c.Params("workId"))
	if err != nil {
		return c.JSON(errRes("Params(workId)", err, config.PARAM_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	service := config.EverydayWorkService.Of(employeeObjId)
	work := service.Reader.Work(workId)
	cashier_works, err := work.CashierActivities(page, limit)
	if err != nil {
		return c.JSON(errRes("CashierActivities()", err, config.CANT_DECODE))
	}

	return c.JSON(models.Response[[]models.CompletedCashierWork]{
		IsSuccess: true,
		Result:    cashier_works,
	})
}
func GetEmployeeInfo(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetEmployeeInfo")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var empInfo models.EmployeeInfo
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	employeeCursor, err := employeesColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": employeeObjId,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "reasons",
				"localField":   "_id",
				"foreignField": "employee_id",
				"as":           "reasons",
				"pipeline": bson.A{
					bson.M{
						"$sort": bson.M{
							"created_at": -1,
						},
					},
					bson.M{
						"$limit": 2,
					},
					bson.M{
						"$project": bson.M{
							"employee_id": 0,
							"created_at":  0,
							"created_by":  0,
							"days":        0,
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "notes",
				"localField":   "_id",
				"foreignField": "employee_id",
				"as":           "notes",
				"pipeline": bson.A{
					bson.M{
						"$sort": bson.M{
							"created_at": -1,
						},
					},
					bson.M{
						"$limit": 2,
					},
					bson.M{
						"$lookup": bson.M{
							"from":         "employees",
							"localField":   "created_by",
							"foreignField": "_id",
							"as":           "created_by",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"_id": 1,
										"full_name": bson.M{
											"$concat": bson.A{"$name", " ", "$surname"},
										},
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$created_by",
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "everyday_work",
				"localField":   "_id",
				"foreignField": "employee_id",
				"as":           "everyday_work",
				"pipeline": bson.A{
					bson.M{
						"$sort": bson.M{
							"date": -1,
						},
					},
					bson.M{
						"$limit": 2,
					},
				},
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	if employeeCursor.Next(ctx) {
		err = employeeCursor.Decode(&empInfo)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = employeeCursor.Err(); err != nil {
		return c.JSON(errRes("employeeCursor.Err()", err, config.DBQUERY_ERROR))
	}

	if empInfo.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Employee not found."), config.NOT_FOUND))
	}

	for i := 0; i < len(empInfo.EverydayWork); i++ {
		if empInfo.Job.Name == config.ADMIN_CHECKER {
			empInfo.EverydayWork[i].Note = fmt.Sprintf("%v checked", empInfo.EverydayWork[i].CompletedTasksCount)
		} else if empInfo.Job.Name == config.CASHIER {
			empInfo.EverydayWork[i].Note = fmt.Sprintf("%v requests", empInfo.EverydayWork[i].CompletedTasksCount)
		}
	}

	return c.JSON(models.Response[models.EmployeeInfo]{
		IsSuccess: true,
		Result:    empInfo,
	})

}

// TODO: dismiss edenson access_token & refresh_token lary nadip gecersiz kilariz?
func DismissEmployee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.DismissEmployee")
	employeeObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var oldData models.EmployeeForDismissing
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	cursor, err := employeesColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": employeeObjId,
			},
		},
		bson.M{
			"$project": bson.M{
				"salaries":   1,
				"started_on": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		err = cursor.Decode(&oldData)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	now := time.Now()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	update_model := bson.M{
		"exited_on": today,
		"reason":    nil,
	}
	salaries := []models.Salary{}
	for _, s := range oldData.Salaries {
		if s.To == nil {
			s.To = &today
		}
		salaries = append(salaries, s)
	}
	update_model["salaries"] = salaries
	started_ons := []models.StartedOn{}
	for _, s := range oldData.StartedOn {
		if s.To == nil {
			s.To = &today
		}
		started_ons = append(started_ons, s)
	}
	update_model["started_on"] = started_ons
	updateResult, err := employeesColl.UpdateOne(ctx, bson.M{"_id": employeeObjId}, bson.M{
		"$set": update_model,
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		return c.JSON(errRes("UpdateOne()", errors.New("Employee not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "Employee dismissed successfully.",
	})
}
