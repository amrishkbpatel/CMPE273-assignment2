package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type LocationCont struct {
	session *mgo.Session
}

type Input struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type Output struct {
	Id      bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name    string        `json:"name"`
	Address string        `json:"address"`
	City    string        `json:"city" `
	State   string        `json:"state"`
	Zip     string        `json:"zip"`

	Coordinate struct {
		Lat  string `json:"lat"`
		Lang string `json:"lang"`
	}
}

type mapsResponse struct {
	Results []mapsResult
}

type mapsResult struct {
	Address      string        `json:"formatted_address"`
	AddressParts []AddressPart `json:"address_components"`
	Geometry     Geometry
	Types        []string
}

type AddressPart struct {
	Name      string `json:"long_name"`
	ShortName string `json:"short_name"`
	Types     []string
}

func NewLocationCont(s *mgo.Session) *LocationCont {
	return &LocationCont{s}
}

//google maps api response
func getGoogLocation(address string) Output {
	client := &http.Client{}

	reqURL := "http://maps.google.com/maps/api/geocode/json?address="
	reqURL += url.QueryEscape(address)
	reqURL += "&sensor=false"
	fmt.Println("URL: " + reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error: ", err)
	}

	var res mapsResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("error in unmashalling: ", err)
	}

	var ret Output
	ret.Coordinate.Lat = strconv.FormatFloat(res.Results[0].Geometry.Location.Lat, 'f', 7, 64)
	ret.Coordinate.Lang = strconv.FormatFloat(res.Results[0].Geometry.Location.Lng, 'f', 7, 64)

	return ret
}

type Geometry struct {
	Bounds   Bounds
	Location Point
	Type     string
	Viewport Bounds
}
type Bounds struct {
	NorthEast, SouthWest Point
}

type Point struct {
	Lat float64
	Lng float64
}

// CreateLocation
func (uc LocationCont) CreateLoc(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var u Input
	var oA Output

	json.NewDecoder(r.Body).Decode(&u)
	googResCoor := getGoogLocation(u.Address + "+" + u.City + "+" + u.State + "+" + u.Zip)
	fmt.Println("resp is: ", googResCoor.Coordinate.Lat, googResCoor.Coordinate.Lang)

	oA.Id = bson.NewObjectId()
	oA.Name = u.Name
	oA.Address = u.Address
	oA.City = u.City
	oA.State = u.State
	oA.Zip = u.Zip
	oA.Coordinate.Lat = googResCoor.Coordinate.Lat
	oA.Coordinate.Lang = googResCoor.Coordinate.Lang

	// Write the user to mongo
	uc.session.DB("locations").C("locationA").Insert(oA)

	uj, _ := json.Marshal(oA)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

// GetLocation
func (uc LocationCont) GetLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("location_id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	var o Output
	if err := uc.session.DB("locations").C("locationA").FindId(oid).One(&o); err != nil {
		w.WriteHeader(404)
		return
	}
	uj, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

// RemoveLocation
func (uc LocationCont) RemoveLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("location_id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	// get id
	oid := bson.ObjectIdHex(id)

	// Remove user
	if err := uc.session.DB("locations").C("locationA").RemoveId(oid); err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
}

//UpdateLocation
func (uc LocationCont) UpdateLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var i Input
	var o Output

	id := p.ByName("location_id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	oid := bson.ObjectIdHex(id)

	if err := uc.session.DB("locations").C("locationA").FindId(oid).One(&o); err != nil {
		w.WriteHeader(404)
		return
	}

	json.NewDecoder(r.Body).Decode(&i)
	googResCoor := getGoogLocation(i.Address + "+" + i.City + "+" + i.State + "+" + i.Zip)
	fmt.Println("resp is: ", googResCoor.Coordinate.Lat, googResCoor.Coordinate.Lang)

	o.Address = i.Address
	o.City = i.City
	o.State = i.State
	o.Zip = i.Zip
	o.Coordinate.Lat = googResCoor.Coordinate.Lat
	o.Coordinate.Lang = googResCoor.Coordinate.Lang

	// Write the user to mongo
	c := uc.session.DB("locations").C("locationA")

	id2 := bson.M{"_id": oid}
	err := c.Update(id2, o)
	if err != nil {
		panic(err)
	}

	uj, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}
