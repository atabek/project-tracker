package model

import (
	"fmt"
	"time"

	"app/shared/database"

	"gopkg.in/mgo.v2/bson"
)

// *****************************************************************************
// Project
// *****************************************************************************

// Project table contains the information for each project
type Project struct {
	ObjectID  bson.ObjectId `bson:"_id"`
    // Don't use Id, use ProjectID() instead for consistency with MongoDB
	ID               uint32           `db:"id" bson:"id,omitempty"`
	ProjectName      string           `db:"project_name" bson:"project_name"`
	CustomerCompany  string           `db:"customer_company" bson:"customer_company"`
    EmployeeCompany  string           `db:"employee_company" bson:"employee_company"`
	Supervisor       string           `db:"supervisor" bson:"supervisor"`
	Priority         int64            `db:"priority" bson:"priority"`
    StartDate        string           `db:"start_date" bson:"start_date"`
    EndDate          string           `db:"end_date" bson:"end_date"`
	CreatedAt        time.Time        `db:"created_at" bson:"created_at"`
	UpdatedAt        time.Time        `db:"updated_at" bson:"updated_at"`
    Done             bool             `db:"done" bson:"done"`
	Deleted          uint8            `db:"deleted" bson:"deleted"`
	EmployeeIDs      []bson.ObjectId  `db:"employee_ids" bson:"employee_ids"`
}

// ProjectID returns the project id
func (p *Project) ProjectID() string {
	r := ""

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		r = fmt.Sprintf("%v", p.ID)
	case database.TypeMongoDB:
		r = p.ObjectID.Hex()
	case database.TypeBolt:
		r = p.ObjectID.Hex()
	}

	return r
}

// ProjectByID gets Project by ID
func ProjectByID(projectID string) (Project, error) {
	var err error

	result := Project{}

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("project")

			// Validate the object id
			if bson.IsObjectIdHex(projectID) {
				err = c.FindId(bson.ObjectIdHex(projectID)).One(&result)
				if err != nil {
					result = Project{}
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

// ProjectByName gets employee information from project name
func ProjectByName(projectName string) (Project, error) {
	var err error

	result := Project{}

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("project")
			err = c.Find(bson.M{"project_name": projectName}).One(&result)
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func GetAllProjects() ([]Project, error){
	var err error

	var result []Project

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("project")

			err = c.Find(nil).All(&result)
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func ProjectsByEmployeeID(employeeID string) ([]Project, error) {
	var err error
	var project_ids []bson.ObjectId

	results := []Project{}
	result  := Project{}

	employee := Employee{}


	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("employee")
			p := session.DB(database.ReadConfig().MongoDB.Database).C("project")

			err = c.FindId(bson.ObjectIdHex(employeeID)).One(&employee)
			if err != nil{
				employee = Employee{}
				err = ErrNoResult
			} else {
				project_ids = employee.ProjectIDs
				for _, project_id := range project_ids {
					err = p.FindId(project_id).One(&result)
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


func ProjectCreate(projectName, customerCompany, employeeCompany,
				   supervisor string, priority int64, startDate ,
				   endDate string, done bool) error {
	var err error
	now := time.Now()

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("project")

			project := &Project{
				ObjectID:  bson.NewObjectId(),
				ProjectName: projectName,
				CustomerCompany:  customerCompany,
				EmployeeCompany:     employeeCompany,
				Supervisor : supervisor,
				Priority:  priority,
				StartDate: startDate,
				EndDate: endDate,
				CreatedAt: now,
				UpdatedAt: now,
				Done: false,
				Deleted:   0,
			}
			err = c.Insert(project)
		} else {
			err = ErrUnavailable
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}


// ProjectUpdate updates a project
func ProjectUpdate(projectName, customerCompany, employeeCompany, supervisor string,
	priority int64, startDate, endDate string, done bool, projectID string) error {

	var err error

	now := time.Now()

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("project")
			var project Project
			project, err = ProjectByID(projectID)
			if err == nil {
				project.ProjectName     = projectName
				project.CustomerCompany = customerCompany
				project.EmployeeCompany = employeeCompany
				project.Supervisor      = supervisor
				project.Priority        = priority
				project.StartDate       = startDate
				project.EndDate         = endDate
				project.UpdatedAt       = now
				project.Done            = done

				err = c.UpdateId(bson.ObjectIdHex(projectID), &project)
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

func ProjectUpdateByEmployeeIDs(projectID string, employeeIDs ...string) error {
	var err error

	now := time.Now()
	var ids []bson.ObjectId


	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("project")
			var project Project
			project, err = ProjectByID(projectID)
			if err == nil {
				project.UpdatedAt   = now
				for _, v := range employeeIDs {
					ids = append(ids, bson.ObjectIdHex(v))
				}
				project.EmployeeIDs = ids
				err = c.UpdateId(bson.ObjectIdHex(projectID), &project)
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

// ProjectDelete deletes a project
func ProjectDelete(projectID string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMongoDB:
		if database.CheckConnection() {
			// Create a copy of mongo
			session := database.Mongo.Copy()
			defer session.Close()
			c := session.DB(database.ReadConfig().MongoDB.Database).C("project")

			_, err = ProjectByID(projectID)
			if err == nil {
				err = c.RemoveId(bson.ObjectIdHex(projectID))
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
