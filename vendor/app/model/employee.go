package model

import (
	"fmt"
	"time"

	"app/shared/database"

	"gopkg.in/mgo.v2/bson"
)

// *****************************************************************************
// Employee
// *****************************************************************************

// Employee table contains the information for each employee
type Employee struct {
	ObjectID  bson.ObjectId `bson:"_id"`
    // Don't use Id, use EmployeeID() instead for consistency with MongoDB
	ID        uint32           `db:"id" bson:"id,omitempty"`
	FirstName string           `db:"first_name" bson:"first_name"`
	LastName  string           `db:"last_name" bson:"last_name"`
	Email     string           `db:"email" bson:"email"`
	StatusID  uint8            `db:"status_id" bson:"status_id"`
	CreatedAt time.Time        `db:"created_at" bson:"created_at"`
	UpdatedAt time.Time        `db:"updated_at" bson:"updated_at"`
	Deleted   uint8            `db:"deleted" bson:"deleted"`
	ProjectIDs []bson.ObjectId `db:"project_ids" bson:"project_ids"`
}

// // EmployeeStatus table contains every possible employee status (active/inactive)
// type EmployeeStatus struct {
// 	ID        uint8     `db:"id" bson:"id"`
// 	Status    string    `db:"status" bson:"status"`
// 	CreatedAt time.Time `db:"created_at" bson:"created_at"`
// 	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`
// 	Deleted   uint8     `db:"deleted" bson:"deleted"`
// }

// EmployeeID returns the employee id
func (e *Employee) EmployeeID() string {
	r := ""

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		r = fmt.Sprintf("%v", e.ID)
	case database.TypeMongoDB:
		r = e.ObjectID.Hex()
	case database.TypeBolt:
		r = e.ObjectID.Hex()
	}

	return r
}

// EmployeeByEmail gets employee information from email
func EmployeeByEmail(email string) (Employee, error) {
	var err error

	result := Employee{}

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")
			err = c.Find(bson.M{"email": email}).One(&result)
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}


// EmployeeByID gets Employee by ID
func EmployeeByID(employeeID string) (Employee, error) {
	var err error

	result := Employee{}

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")

			// Validate the object id
			if bson.IsObjectIdHex(employeeID) {
				err = c.FindId(bson.ObjectIdHex(employeeID)).One(&result)
				if err != nil {
					result = Employee{}
					err = ErrNoResult
				}
			} else {
				err = ErrNoResult
			}
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func EmployeesByProjectID(projectID string) ([]Employee, error) {
	var err error
	var employee_ids []bson.ObjectId

	results := []Employee{}
	result  := Employee{}

	project := Project{}


	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")
			p := session.DB(database.ReadConfig().MongoDB.Database).C("project")

			//err = c.FindId(bson.ObjectIdHex(employeeID)).One(&result)


				//fmt.Println(i, s)
			err = p.FindId(bson.ObjectIdHex(projectID)).One(&project)
			if err != nil{
				project = Project{}
				err = ErrNoResult
			} else {
				employee_ids = project.EmployeeIDs
				for _, employee_id := range employee_ids {
					err = c.FindId(employee_id).One(&result)
					if err == nil{
						results = append(results, result)
					} else {
						err = ErrNoResult
					}
				}
			}

		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return results, standardizeError(err)
}

func GetAllEmployees() ([]Employee, error){
	var err error

	var result []Employee

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")

			err = c.Find(nil).All(&result)
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EmployeeCreate creates employee
func EmployeeCreate(firstName, lastName, email string) error {
	var err error

	now := time.Now()

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")

			employee := &Employee{
				ObjectID:  bson.NewObjectId(),
				FirstName: firstName,
				LastName:  lastName,
				Email:     email,
				StatusID:  1,
				CreatedAt: now,
				UpdatedAt: now,
				Deleted:   0,
			}
			err = c.Insert(employee)
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// EmployeeUpdate updates a Employee
func EmployeeUpdate(firstName, lastName, email, employeeID string) error {
	var err error

	now := time.Now()

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")
			var employee Employee
			employee, err = EmployeeByID(employeeID)
			fmt.Println(employee)
			fmt.Println(err)
			if err == nil {
				employee.UpdatedAt = now
                employee.FirstName = firstName
				employee.LastName =  lastName
				employee.Email    = email
				err = c.UpdateId(bson.ObjectIdHex(employeeID), &employee)
			} else {
					err = ErrUnauthorized
			}
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EmployeeUpdateByProjectIDs(employeeID string, projectIDs ...string) error {
	var err error

	now := time.Now()
	var ids []bson.ObjectId


	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")
			var employee Employee
			employee, err = EmployeeByID(employeeID)
			if err == nil {
				employee.UpdatedAt   = now
				for _, v := range projectIDs {
					ids = append(ids, bson.ObjectIdHex(v))
				}
				employee.ProjectIDs = ids
				err = c.UpdateId(bson.ObjectIdHex(employeeID), &employee)
			} else {
					err = ErrUnauthorized
			}
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}


// EmployeeDelete updates a Employee
func EmployeeDelete(employeeID string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")

			_, err = EmployeeByID(employeeID)
			if err == nil {
				err = c.RemoveId(bson.ObjectIdHex(employeeID))
				if err != nil{
					err = ErrUnavailable
				}
			}
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}


	return standardizeError(err)
}
