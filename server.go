package main

import (
	"controllers"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
)

func main() {
	r := httprouter.New()

	uc := controllers.NewLocationCont(getSession())

	r.GET("/locations/:location_id", uc.GetLoc)

	r.POST("/locations", uc.CreateLoc)

	r.PUT("/locations/:location_id", uc.UpdateLoc)

	r.DELETE("/locations/:location_id", uc.RemoveLoc)

	http.ListenAndServe("localhost:8083", r)
}

func getSession() *mgo.Session {
	// Connect to mongolabs
	s, err := mgo.Dial("mongodb://admin:admin@ds045064.mongolab.com:45064/locations")

	if err != nil {
		panic(err)
	}

	s.SetMode(mgo.Monotonic, true)
	return s
}
