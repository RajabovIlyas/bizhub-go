package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmployeeForLoginWithPassword struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id"`
	FullName string             `json:"full_name" bson:"full_name"`
	Job      Job                `json:"job" bson:"job"`
	Avatar   string             `json:"avatar" bson:"avatar"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password" bson:"password"`
}

type EverydayWorkWithoutTasks struct {
	Id                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	EmployeeId          primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	Date                time.Time          `json:"date" bson:"date"`
	CompletedTasksCount int64              `json:"completed_tasks_count" bson:"completed_tasks_count"`
	Note                string             `json:"note" bson:"note"`
}

func (e *EmployeeForLoginWithPassword) WithoutPassword() *EmployeeForLoginWithoutPassword {
	return &EmployeeForLoginWithoutPassword{
		Id:       e.Id,
		FullName: e.FullName,
		Job:      e.Job,
		Avatar:   e.Avatar,
		Email:    e.Email,
	}
}

type Job struct {
	Name        string `json:"name" bson:"name"`
	DisplayName string `json:"display_name" bson:"display_name"`
}
type EmployeeForLoginWithoutPassword struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id"`
	FullName string             `json:"full_name" bson:"full_name"`
	Job      Job                `json:"job" bson:"job"`
	Avatar   string             `json:"avatar" bson:"avatar"`
	Email    string             `json:"email" bson:"email"`
}
type EmployeeForAdmin struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id"`
	FullName string             `json:"full_name" bson:"full_name"`
	Job      Job                `json:"job" bson:"job"`
	Avatar   string             `json:"avatar" bson:"avatar"`
	ExitedOn *time.Time         `json:"exited_on" bson:"exited_on"`
	Reason   *ReasonOfEmployee  `json:"reason" bson:"reason"`
}
type ReasonOfEmployee struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	DisplayName string             `json:"display_name" bson:"display_name"`
}
type Reason struct {
	Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	DisplayName string             `json:"display_name" bson:"display_name"`
	From        time.Time          `json:"from" bson:"from"`
	Days        int                `json:"days" bson:"days"`
	To          time.Time          `json:"to" bson:"to"`
	Description string             `json:"description" bson:"description"`
	EmployeeId  primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	CreatedBy   primitive.ObjectID `json:"created_by" bson:"created_by"`
	CreatedAt   time.Time          `json:"created_at,string" bson:"created_at"`
}
type ReasonForEmpInfo struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	DisplayName string             `json:"display_name" bson:"display_name"`
	From        time.Time          `json:"from" bson:"from"`
	To          time.Time          `json:"to" bson:"to"`
	Description string             `json:"description" bson:"description"`
}
type ReasonFromFE struct {
	Name        string `json:"name" bson:"name"`
	DisplayName string `json:"display_name" bson:"display_name"`
	From        string `json:"from" bson:"from"`
	Days        int    `json:"days" bson:"days"`
	Description string `json:"description" bson:"description"`
}

func (r *Reason) ToString() {
	fmt.Printf("\nId : %v\n", r.Id)
	fmt.Printf("\nName : %v\n", r.Name)
	fmt.Printf("\nDisplayName : %v\n", r.DisplayName)
	fmt.Printf("\nFrom : %v\n", r.From)
	fmt.Printf("\nDays: %v\n", r.Days)
	fmt.Printf("\nTo : %v\n", r.To)
	fmt.Printf("\nDescription : %v\n", r.Description)
	fmt.Printf("\nEmployeeId : %v\n", r.EmployeeId)
	fmt.Printf("\nCreatedBy : %v\n", r.CreatedBy)
	fmt.Printf("\nCreatedAt : %v\n", r.CreatedAt)
}

type Note struct {
	Id         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IsLike     bool               `json:"is_like" bson:"is_like"`
	Text       string             `json:"text" bson:"text"`
	EmployeeId primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	CreatedBy  primitive.ObjectID `json:"created_by" bson:"created_by"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}
type NoteForEmpInfo struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IsLike    bool               `json:"is_like" bson:"is_like"`
	Text      string             `json:"text" bson:"text"`
	CreatedBy struct {
		Id       primitive.ObjectID `json:"_id" bson:"_id"`
		FullName string             `json:"full_name" bson:"full_name"`
	} `json:"created_by" bson:"created_by"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
type EmployeeInfo struct {
	Id             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name"`
	Surname        string             `json:"surname" bson:"surname"`
	MiddleName     string             `json:"middle_name" bson:"middle_name"`
	BirthDate      time.Time          `json:"birth_date" bson:"birth_date"`
	Address        string             `json:"address" bson:"address"`
	PassportCode   string             `json:"passport_code" bson:"passport_code"`
	PassportCopies []string           `json:"passport_copies" bson:"passport_copies"`
	Email          string             `json:"email" bson:"email"`
	Avatar         string             `json:"avatar" bson:"avatar"`
	Job            Job                `json:"job" bson:"job"`
	Salaries       []Salary           `json:"salaries" bson:"salaries"`
	WorkTime       WorkTime           `json:"work_time" bson:"work_time"`
	StartedOn      []StartedOn        `json:"started_on" bson:"started_on"`
	// Password string `json:"password" bson:"password"`
	ExitedOn     *time.Time                 `json:"exited_on" bson:"exited_on"`
	Reasons      []ReasonForEmpInfo         `json:"reasons" bson:"reasons"`
	Notes        []NoteForEmpInfo           `json:"notes" bson:"notes"`
	EverydayWork []EverydayWorkWithoutTasks `json:"everyday_work" bson:"everyday_work"`
}
type EmployeeInfoForEditing struct {
	Id             primitive.ObjectID `json:"_id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	Surname        string             `json:"surname" bson:"surname"`
	MiddleName     string             `json:"middle_name" bson:"middle_name"`
	BirthDate      time.Time          `json:"birth_date" bson:"birth_date"`
	Address        string             `json:"address" bson:"address"`
	PassportCode   string             `json:"passport_code" bson:"passport_code"`
	PassportCopies []string           `json:"passport_copies" bson:"passport_copies"`
	Email          string             `json:"email" bson:"email"`
	Password       string             `json:"password" bson:"password"`
	Avatar         string             `json:"avatar" bson:"avatar"`
	Job            Job                `json:"job" bson:"job"`
	Salary         float64            `json:"salary" bson:"salary"`
	WorkTime       WorkTime           `json:"work_time" bson:"work_time"`
}
type NewEmployee struct {
	Id             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name"`
	Surname        string             `json:"surname" bson:"surname"`
	MiddleName     string             `json:"middle_name" bson:"middle_name"`
	BirthDate      time.Time          `json:"birth_date" bson:"birth_date"`
	Address        string             `json:"address" bson:"address"`
	PassportCode   string             `json:"passport_code" bson:"passport_code"`
	PassportCopies []string           `json:"passport_copies" bson:"passport_copies"`
	Email          string             `json:"email" bson:"email"`
	Avatar         string             `json:"avatar" bson:"avatar"`
	Job            Job                `json:"job" bson:"job"`
	Salaries       []Salary           `json:"salaries" bson:"salaries"`
	WorkTime       WorkTime           `json:"work_time" bson:"work_time"`
	StartedOn      []StartedOn        `json:"started_on" bson:"started_on"`
	ExitedOn       *time.Time         `json:"exited_on" bson:"exited_on"`
	Password       string             `json:"password" bson:"password"`
	CreatedBy      primitive.ObjectID `json:"created_by" bson:"created_by"`
	Reason         *ReasonOfEmployee  `json:"reason" bson:"reason"`
}

func (e NewEmployee) HasEmptyFields() bool {
	if e.Name == "" || e.Surname == "" || e.MiddleName == "" || e.Address == "" ||
		e.PassportCode == "" || e.Email == "" || e.Avatar == "" || e.Job.Name == "" ||
		e.Job.DisplayName == "" || e.WorkTime.Start == "" || *e.WorkTime.End == "" ||
		e.Password == "" || len(e.PassportCopies) == 0 || e.Avatar == "" {
		return true
	}
	return false
}

type EmployeeForDismissing struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Salaries  []Salary           `json:"salaries" bson:"salaries"`
	StartedOn []StartedOn        `json:"started_on" bson:"started_on"`
}
type Salary struct {
	Amount float64    `json:"amount" bson:"amount"`
	From   time.Time  `json:"from" bson:"from"`
	To     *time.Time `json:"to" bson:"to"`
}
type WorkTime struct {
	Start string  `json:"start" bson:"start"`
	End   *string `json:"end" bson:"end"`
}
type StartedOn struct {
	From time.Time  `json:"from" bson:"from"`
	To   *time.Time `json:"to" bson:"to"`
}
type DailyWorkForEmpInfo struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Text        string             `json:"text" bson:"text"`
	Date        time.Time          `json:"date" bson:"date"`
	WorkTime    WorkTime           `json:"work_time" bson:"work_time"`
	ChecksCount int64              `json:"checks_count" bson:"checks_count"`
}

/*
	EmployeeId  primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	Products       []primitive.ObjectID `json:"products" bson:"products"`
	Posts          []primitive.ObjectID `json:"posts" bson:"posts"`
	SellerProfiles []primitive.ObjectID `json:"seller_profiles" bson:"seller_profiles"`
	Notificatons   []primitive.ObjectID `json:"notifications" bson:"notifications"`
	Auctions       []primitive.ObjectID `json:"auctions" bson:"auctions"`

62f78f2127cd91b2f3f8067b ok
62f8fe54dcb5dbfcfd9fe9d9 ok
62f8febedcb5dbfcfd9fe9da
62f8ff1adcb5dbfcfd9fe9db
employee_id
62fe46f450a80cbe24d585b9 ok
62fe4df950a80cbe24d585bc ok
62fe4e1b50a80cbe24d585bd ok
62fe4e4950a80cbe24d585be ok
Notificatons
62ce628c8cae982f654a3578 ok
62ce63488cae982f654a3583 ok
62ce63618cae982f654a3585 ok
SellerProfiles
6301ea54680dc66abc356db7 ok
6301f432680dc66abc356dbd ok
6301f4d5680dc66abc356dbe
Auctions
62cf9f30c48e57fb6702a74b ok
62d10733ddd382e3c517c7af ok
62d1078dddd382e3c517c7b0 ok
62d8f5569346332069ff7080 ok
62d8fa3e9346332069ff7088 ok
Products
*/
