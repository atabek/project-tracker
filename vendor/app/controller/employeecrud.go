package controller

import (
	"log"
	"net/http"
	"fmt"

	"app/model"
	"app/shared/recaptcha"
	"app/shared/session"
	"app/shared/view"

	"github.com/gorilla/context"
	"github.com/josephspurrier/csrfbanana"
	"github.com/julienschmidt/httprouter"
)

// EmployeeReadGET displays the entries in the employee
func EmployeeReadGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// userID := fmt.Sprintf("%s", sess.Values["id"])
	// fmt.Println(userID)

	employees, err := model.GetAllEmployees()
	if err != nil {
		log.Println(err)
		employees = []model.Employee{}
	}

	// Display the view
	v := view.New(r)
	v.Name = "employee/read"
	v.Vars["first_name"] = sess.Values["first_name"]
	v.Vars["employees"] = employees
	v.Render(w)
}

// EmployeeCreateGET displays the register page
func EmployeeCreateGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Display the view
	v := view.New(r)
	v.Name = "employee/create"
	v.Vars["token"] = csrfbanana.Token(w, r, sess)
	// Refill any form fields
	view.Repopulate([]string{"first_name", "last_name", "email"}, r.Form, v.Vars)
	v.Render(w)
}

// EmployeeCreatePOST handles the registration form submission
func EmployeeCreatePOST(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Validate with required fields
	if validate, missingField := view.Validate(r, []string{"first_name", "last_name", "email"}); !validate {
		sess.AddFlash(view.Flash{"Field missing: " + missingField, view.FlashError})
		sess.Save(r, w)
		EmployeeCreateGET(w, r)
		return
	}

	// Validate with Google reCAPTCHA
	if !recaptcha.Verified(r) {
		sess.AddFlash(view.Flash{"reCAPTCHA invalid!", view.FlashError})
		sess.Save(r, w)
		EmployeeCreateGET(w, r)
		return
	}

	// Get form values
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")

	// Get database result
	_, err := model.EmployeeByEmail(email)

	if err == model.ErrNoResult { // If success (no user exists with that email)
		ex := model.EmployeeCreate(firstName, lastName, email)
		// Will only error if there is a problem with the query
		if ex != nil {
			log.Println(ex)
			sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
			sess.Save(r, w)
		} else {
			sess.AddFlash(view.Flash{"Account created successfully for: " + email, view.FlashSuccess})
			sess.Save(r, w)
			http.Redirect(w, r, "/employee", http.StatusFound)
			return
		}
	} else if err != nil { // Catch all other errors
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else { // Else the user already exists
		sess.AddFlash(view.Flash{"Account already exists for: " + email, view.FlashError})
		sess.Save(r, w)
	}

	// Display the page
	EmployeeCreateGET(w, r)
}

func EmployeeViewGET(w http.ResponseWriter, r *http.Request) {
	// Get the session
	sess := session.Instance(r)

	// Get the employee id
	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	employeeID := params.ByName("id")
	// userID := fmt.Sprintf("%s", sess.Values["id"])
	// fmt.Println(userID)

	// Get the employee
	employee, err := model.EmployeeByID(employeeID)
	if err != nil { // If the note doesn't exist
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
		http.Redirect(w, r, "/employee", http.StatusFound)
		return
	}

	// Get all of projects
	projects, err := model.GetAllProjects()
	selected_projects, err := model.ProjectsByEmployeeID(employeeID)
	if err != nil {
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
		http.Redirect(w, r, "/employee", http.StatusFound)
		return
	}

	// Display the view
	v := view.New(r)
	v.Name = "employee/view"
	v.Vars["token"] = csrfbanana.Token(w, r, sess)

	v.Vars["employeeID"] = employee.ObjectID
	v.Vars["first_name"] = employee.FirstName
	v.Vars["last_name"]  = employee.LastName
	v.Vars["email"]      = employee.Email
	v.Vars["projects"]   = projects
	v.Vars["selected_projects"] = selected_projects
	v.Render(w)
}

func EmployeeViewPOST(w http.ResponseWriter, r *http.Request) {
	// Get the session
	sess := session.Instance(r)
	var projectIDs []string

	// Get the project id
	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	employeeID := params.ByName("id")

	for k, v := range r.Form {
		if k == "project_ids"{
		    projectIDs = v
		}
    }

	// Get database result
	err := model.EmployeeUpdateByProjectIDs(employeeID, projectIDs...)
	// Will only error if there is a problem with the query
	if err != nil {
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else {
		sess.AddFlash(view.Flash{"Employee updated!", view.FlashSuccess})
		sess.Save(r, w)
		http.Redirect(w, r, "/employee", http.StatusFound)
		return
	}

	// Display the same page
	EmployeeViewGET(w, r)
}

// EmployeeUpdateGET displays the note update page
func EmployeeUpdateGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Get the employee id
	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	employeeID := params.ByName("id")

	userID := fmt.Sprintf("%s", sess.Values["id"])
	fmt.Println(userID)

	// Get the employee
	employee, err := model.EmployeeByID(employeeID)
	if err != nil { // If the note doesn't exist
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
		http.Redirect(w, r, "/employee", http.StatusFound)
		return
	}

	// Display the view
	v := view.New(r)
	v.Name = "employee/update"
	v.Vars["token"] = csrfbanana.Token(w, r, sess)
	v.Vars["first_name"] = employee.FirstName
	v.Vars["last_name"]  = employee.LastName
	v.Vars["email"] = employee.Email
	v.Render(w)
}

// EmployeeUpdatePOST handles the employee update form submission
func EmployeeUpdatePOST(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Validate with required fields
	if validate, missingField := view.Validate(r, []string{"first_name", "last_name", "email"}); !validate {
		sess.AddFlash(view.Flash{"Field missing: " + missingField, view.FlashError})
		sess.Save(r, w)
		EmployeeUpdateGET(w, r)
		return
	}

	// Get form values
	firstName := r.FormValue("first_name")
	lastName  := r.FormValue("last_name")
	email     := r.FormValue("email")

	userID := fmt.Sprintf("%s", sess.Values["id"])
	fmt.Println("userID: ", userID)

	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	employeeID := params.ByName("id")
	fmt.Println("employeeID: ", employeeID)

	// e, err := model.EmployeeByID(employeeID)
	// if err != nil{
	// 	log.Println(err)
	// 	sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
	// 	sess.Save(r, w)
	// }

	// Get database result
	err := model.EmployeeUpdate(firstName, lastName, email, employeeID)
	// Will only error if there is a problem with the query
	if err != nil {
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else {
		sess.AddFlash(view.Flash{"Employee updated!", view.FlashSuccess})
		sess.Save(r, w)
		http.Redirect(w, r, "/employee", http.StatusFound)
		return
	}

	// Display the same page
	EmployeeUpdateGET(w, r)
}

// EmployeeDeleteGET handles the employee deletion
func EmployeeDeleteGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	userID := fmt.Sprintf("%s", sess.Values["id"])
	fmt.Println("User ID: ", userID)

	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	employeeID := params.ByName("id")

	// Get database result
	err := model.EmployeeDelete(employeeID)
	// Will only error if there is a problem with the query
	if err != nil {
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else {
		sess.AddFlash(view.Flash{"Note deleted!", view.FlashSuccess})
		sess.Save(r, w)
	}

	http.Redirect(w, r, "/employee", http.StatusFound)
	return
}
