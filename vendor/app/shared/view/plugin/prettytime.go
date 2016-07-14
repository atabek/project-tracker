package plugin

import (
	"html/template"
	"time"
	// "fmt"
	// "strings"
	"gopkg.in/mgo.v2/bson"
)

// PrettyTime returns a template.FuncMap
// * PRETTYTIME outputs a nice time format
func PrettyTime() template.FuncMap {
	f := make(template.FuncMap)

	f["PRETTYTIME"] = func(t time.Time) string {
		return t.Format("3:04 PM 01/02/2006")
	}

	return f
}

func PrettyObjectId() template.FuncMap {
	f := make(template.FuncMap)

	f["PRETTYOBJECTID"] = func(obj bson.ObjectId) string {
		//fmt.Println("This is a test string for Id: ", id)
		return obj.Hex()
	}

	return f
}
