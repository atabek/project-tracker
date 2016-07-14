package controller

import (
	"log"
	"net/http"
	"fmt"
	"strconv"

	"app/model"
	"app/shared/recaptcha"
	"app/shared/session"
	"app/shared/view"

	"github.com/gorilla/context"
	"github.com/josephspurrier/csrfbanana"
	"github.com/julienschmidt/httprouter"
)

// ProjectReadGET displays the entries in the project
func ProjectReadGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// userID := fmt.Sprintf("%s", sess.Values["id"])
	// fmt.Println(userID)

	projects, err := model.GetAllProjects()
	if err != nil {
		log.Println(err)
		projects = []model.Project{}
	}

	// Display the view
	v := view.New(r)
	v.Name = "project/read"
	v.Vars["first_name"] = sess.Values["first_name"]
	v.Vars["projects"] = projects
	v.Render(w)
}

// ProjectCreateGET displays the register page
func ProjectCreateGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Display the view
	v := view.New(r)
	v.Name = "project/create"
	v.Vars["token"] = csrfbanana.Token(w, r, sess)
	// Refill any form fields
	view.Repopulate([]string{"project_name", "customer_company", "employee_company",
		"supervisor", "priority", "start_date", "end_date", "done"}, r.Form, v.Vars)
	v.Render(w)
}

// ProjectCreatePOST handles the registration form submission
func ProjectCreatePOST(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Validate with required fields
	if validate, missingField := view.Validate(r, []string{"project_name",
		"customer_company", "employee_company", "supervisor", "priority",
		"start_date", "end_date"}); !validate {
		sess.AddFlash(view.Flash{"Field missing: " + missingField, view.FlashError})
		sess.Save(r, w)
		ProjectCreateGET(w, r)
		return
	}

	// Validate with Google reCAPTCHA
	if !recaptcha.Verified(r) {
		sess.AddFlash(view.Flash{"reCAPTCHA invalid!", view.FlashError})
		sess.Save(r, w)
		ProjectCreateGET(w, r)
		return
	}

	// Get form values
	projectName     := r.FormValue("project_name")
	customerCompany := r.FormValue("customer_company")
	employeeCompany := r.FormValue("employee_company")
	supervisor      := r.FormValue("supervisor")
	priority        := r.FormValue("priority")
	startDate       := r.FormValue("start_date")
	endDate         := r.FormValue("end_date")
	done            := r.FormValue("done")

	priorityInt, _ := strconv.ParseInt(priority, 10, 64)
	boolDone, _ := strconv.ParseBool(done)
	fmt.Println("PriotiryInt: ", priorityInt)
	fmt.Println("boolDone: ", boolDone)

	// Get database result
	_, err := model.ProjectByName(projectName)

	if err == model.ErrNoResult { // If success (no user exists with that email)
		ex := model.ProjectCreate(projectName, customerCompany, employeeCompany,
			supervisor, priorityInt, startDate, endDate, boolDone)
		// Will only error if there is a problem with the query
		if ex != nil {
			log.Println(ex)
			sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
			sess.Save(r, w)
		} else {
			sess.AddFlash(view.Flash{"Account created successfully for: " + projectName, view.FlashSuccess})
			sess.Save(r, w)
			http.Redirect(w, r, "/project", http.StatusFound)
			return
		}
	} else if err != nil { // Catch all other errors
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else { // Else the user already exists
		sess.AddFlash(view.Flash{"Project already exists for: " + projectName, view.FlashError})
		sess.Save(r, w)
	}

	// Display the page
	ProjectCreateGET(w, r)
}

// ProjectUpdateGET displays the project update page
func ProjectUpdateGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Get the employee id
	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	projectID := params.ByName("id")

	// userID := fmt.Sprintf("%s", sess.Values["id"])
	// fmt.Println(userID)

	// Get the project
	project, err := model.ProjectByID(projectID)
	if err != nil { // If the note doesn't exist
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
		http.Redirect(w, r, "/project", http.StatusFound)
		return
	}

	// Display the view
	v := view.New(r)
	v.Name = "project/update"
	v.Vars["token"] = csrfbanana.Token(w, r, sess)

	v.Vars["project_name"]      = project.ProjectName
	v.Vars["customer_company"]  = project.CustomerCompany
	v.Vars["employee_company"]  = project.EmployeeCompany
	v.Vars["supervisor"]        = project.Supervisor
	v.Vars["priority"]          = project.Priority
	v.Vars["start_date"]        = project.StartDate
	v.Vars["end_date"]          = project.EndDate
	v.Vars["done"]              = project.Done
	v.Render(w)
}

// ProjectUpdatePOST handles the project update form submission
func ProjectUpdatePOST(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Validate with required fields
	if validate, missingField := view.Validate(r, []string{"project_name",
		"customer_company", "employee_company", "supervisor", "priority",
		"start_date", "end_date"}); !validate {
		sess.AddFlash(view.Flash{"Field missing: " + missingField, view.FlashError})
		sess.Save(r, w)
		EmployeeUpdateGET(w, r)
		return
	}

	// Get form values
	projectName     := r.FormValue("project_name")
	customerCompany := r.FormValue("customer_company")
	employeeCompany := r.FormValue("employee_company")
	supervisor      := r.FormValue("supervisor")
	priority        := r.FormValue("priority")
	startDate       := r.FormValue("start_date")
	endDate         := r.FormValue("end_date")
	done            := r.FormValue("done")

	priorityInt, _ := strconv.ParseInt(priority, 10, 64)
	boolDone, _ := strconv.ParseBool(done)
	fmt.Println("PriotiryInt: ", priorityInt)
	fmt.Println("boolDone: ", boolDone)

	// userID := fmt.Sprintf("%s", sess.Values["id"])
	// fmt.Println("userID: ", userID)

	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	projectID := params.ByName("id")

	// Get database result
	err := model.ProjectUpdate(projectName, customerCompany, employeeCompany,
		supervisor, priorityInt, startDate, endDate, boolDone, projectID)
	// Will only error if there is a problem with the query
	if err != nil {
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else {
		sess.AddFlash(view.Flash{"Project updated!", view.FlashSuccess})
		sess.Save(r, w)
		http.Redirect(w, r, "/project", http.StatusFound)
		return
	}

	// Display the same page
	ProjectUpdateGET(w, r)
}

// ProjectDeleteGET handles the project deletion
func ProjectDeleteGET(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	userID := fmt.Sprintf("%s", sess.Values["id"])
	fmt.Println("User ID: ", userID)

	var params httprouter.Params
	params = context.Get(r, "params").(httprouter.Params)
	projectID := params.ByName("id")

	// Get database result
	err := model.ProjectDelete(projectID)
	// Will only error if there is a problem with the query
	if err != nil {
		log.Println(err)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
	} else {
		sess.AddFlash(view.Flash{"Project deleted!", view.FlashSuccess})
		sess.Save(r, w)
	}

	http.Redirect(w, r, "/project", http.StatusFound)
	return
}
