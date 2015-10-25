package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"strings"
)

type GCoordinates struct {
	R_obj  []Results `json:"results"`
	status string    `json:"status"`
}

type Results struct {
	AC               []AddressComponents `json:"address_components"`
	FormattedAddress string              `json:"formatted_address"`
	Geo              Geometry            `json:"geometry"`
	PlaceID          string              `json:"place_id"`
	Types            []string            `json:"types"`
}

type AddressComponents struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

type Geometry struct {
	Loc          Location `json:"location"`
	LocationType string   `json:"location_type"`
	Vp           Viewport `json:"viewport"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Viewport struct {
	Ne Northeast `json:"northeast"`
	Sw Southwest `json:"southwest"`
}

type Northeast struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Southwest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Response struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Name    string        `json:"name" bson:"name"`
	Address string        `json:"address" bson:"address" `
	City    string        `json:"city"  bson:"city"`
	State   string        `json:"state"  bson:"state"`
	ZipCode string        `json:"zip"  bson:"zip" `
	Co      Cordinates    `json:"coordinate" bson:"coordinate"`
}

type Cordinates struct {
	Lat float64 `json:"lat"   bson:"lat"`
	Lng float64 `json:"lng"   bson:"lng"`
}

func getSession() *mgo.Session {
	// Connect to our local mongo
	s, err := mgo.Dial("mongodb://user123:pass12345@ds041934.mongolab.com:41934/users")
	//s, err := mgo.Dial("mongodb://<dbuser>:<dbpassword>@ds041934.mongolab.com:41934/users")

	// Check if connection error, is mongo running?
	if err != nil {
		panic(err)
	}
	return s
}

type UserController struct {
	session *mgo.Session
}

func New_User(s *mgo.Session) *UserController {
	return &UserController{s}
}

func create_post(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	uc := New_User(getSession())
	var V Response

	json.NewDecoder(r.Body).Decode(&V)

	ans := giveurl(&V)

	ans.Id = bson.NewObjectId()
	uc.session.DB("users").C("hello").Insert(ans)
	UJ, _ := json.Marshal(ans)
	fmt.Fprintf(rw, "%s", UJ)
}

func read_get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	uc := New_User(getSession())
	id := p.ByName("id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	v := Response{}

	if err := uc.session.DB("users").C("hello").FindId(oid).One(&v); err != nil {
		w.WriteHeader(404)
		return
	}

	uj, _ := json.Marshal(v)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)

}

func updatedb(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	uc := New_User(getSession())
	id := p.ByName("id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)

	var obj1 Response
	var obj2 Response

	if err := uc.session.DB("users").C("hello").FindId(oid).One(&obj2); err != nil {
		w.WriteHeader(404)
		return
	}

	json.NewDecoder(r.Body).Decode(&obj1)

	ans := giveurl(&obj1)

	ans.Id = oid
	if err := uc.session.DB("users").C("hello").Update(obj2, ans); err != nil {
		w.WriteHeader(404)
		return
	}

	uj, _ := json.Marshal(ans)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)

}

func D_delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	uc := New_User(getSession())

	id := p.ByName("id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)

	if err := uc.session.DB("users").C("hello").RemoveId(oid); err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
}

func giveurl(res *Response) Response {

	var Gcord GCoordinates

	Url_address := res.Address
	Url_city := res.City
	Url_state := res.State
	Url_zip := res.City

	Final_url_strng := Url_address + "," + Url_city + "," + Url_state + "," + Url_zip
	Final_url_strng = strings.Replace(Final_url_strng, " ", "+", -1)
	Final_url_strng = "https://maps.google.com/maps/api/geocode/json?address=" + Final_url_strng + "&sensor=false"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", Final_url_strng, nil)
	resp, _ := client.Do(req)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &Gcord)
	}

	for _, Sample := range Gcord.R_obj {

		res.Co.Lat = Sample.Geo.Loc.Lat

		res.Co.Lng = Sample.Geo.Loc.Lng
	}

	return *res
}

func main() {

	a := httprouter.New()
	a.POST("/locations", create_post)
	a.GET("/locations/:id", read_get)
	a.PUT("/locations/:id", updatedb)
	a.DELETE("/locations/:id", D_delete)
	http.ListenAndServe("Localhost:6666", a)
}
