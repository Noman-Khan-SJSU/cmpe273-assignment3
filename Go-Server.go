/*
Author: Noman Khan
SJSU ID: 009801103
Course: CMPE 273
Term: Fall 2015
*/

package main

import (

    "errors"
    "strconv"
    "github.com/jmoiron/jsonq"
    "gopkg.in/mgo.v2/bson"
    "gopkg.in/mgo.v2"
    "fmt"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "bytes"
    "github.com/julienschmidt/httprouter"
    "github.com/r-medina/go-uber"
    "strings"
    "sort"

)

func main()  {
    router  := httprouter.New()
    router .GET("/locations/:uniqueid", get)
    router .POST("/locations", create)
    router .PUT("/locations/:uniqueid", update)
    router .DELETE("/locations/:uniqueid", delete)
    router .POST("/trips",shortroute)
    router .GET("/trips/:tripid", gettrip)
    router .PUT("/trips/:tripid/request", triptrack)
    server := http.Server{
        Addr:        "0.0.0.0:8181",
        Handler: router ,
    }
    server.ListenAndServe()
}

const (

ACCESS_TOKEN = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsImhpc3RvcnkiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiZWY4MjcyNzItMWNkZC00NTUzLWE3NDQtMTlmOTBkY2FmM2NkIiwiaXNzIjoidWJlci11czEiLCJqdGkiOiIyOWE4NjJmNi1lMTczLTRlYTctOGViZS03NjkzZTVjYWZiZGEiLCJleHAiOjE0NTA3NjQzNTcsImlhdCI6MTQ0ODE3MjM1NiwidWFjdCI6IjU0RndzN2Rxc25xcHNKa0xUSXJXWTVwbEplaU1IMiIsIm5iZiI6MTQ0ODE3MjI2NiwiYXVkIjoiT0hQWGZ1NmFPOUcwcFFKc28zYW5rQ29DSnFpYTlIUVEifQ.dNxhW4Dm1woicq1yyMKxyiceXrmaEkBFS3HLqf-k_AVI0zODmQIctEcyl3jFb1-YApRlQXpiLHIvtYr2PBWZ0MtgpxuAreBF1AtYYmzmKoH3jxFbjGzlgsPrgEdcSVepuqL-645ffosE6Q4HmmG0KV7HI46BcbsACJjTDYilvoszj5ChJCpU60BLoYk9dNtZlCJ9EHk5J-09ktkODSuhRN99xH4YCe4FWfTppfe_IIIqlB1QeYwdd5cnHOnwKMVIJGjtxrXQ6UZ8wIb-De_Q04HUF3OyXZtPoUBlpnMkhaHa4PKUo8uC2jQf3Yd3s6-QUeUX_1BDsqrc1ns7fSt45g"
SERVER_TOKEN = "2EOH3qEo3qmVVKJDPy6HaYKiEeyVmkB5-wIKo94_"
CLIENT_ID = "OHPXfu6aO9G0pQJso3ankCoCJqia9HQQ"
CLIENT_SECRET = "98_3czHvHco2MwGIayjMVj9a06K-jumoc_qD6BUA"
REDIRECT_URL = "http://localhost:8080"

)

var Locids []string
var nextid string
var startid string
type dataSlice []Data
type eta struct {

	Eta             int         `json:"eta"`
	RequestID       string      `json:"request_id"`
	Status          string      `json:"status"`
	SurgeMultiplier float64         `json:"surge_multiplier"`

}
type Tdata struct {

Id bson.ObjectId `json:"id" bson:"_id"`
Status string  `json:"status" bson:"status"`
Starting_from_location_id string `json:"starting_from_location_id" bson:"starting_from_location_id"`
 Best_route_location_ids []string `json:"best_route_location_ids" bson:"best_route_location_ids"`
  Total_uber_costs int `json:"total_uber_cost" bson:"total_uber_cost"`
  Total_uber_duration int `json:"total_uber_duration" bson:"total_uber_duration"`
  Total_distance float64 `json:"total_distance" bson:"total_distance"`

}

type Tstatus struct {

Id bson.ObjectId `json:"id" bson:"_id"`
Status string  `json:"status" bson:"status"`
Starting_from_location_id string `json:"starting_from_location_id" bson:"starting_from_location_id"`
 Next_destination_location_id string `json:"next_destination_location_id" bson:"next_destination_location_id"`
 Best_route_location_ids []string `json:"best_route_location_ids" bson:"best_route_location_ids"`
  Total_uber_costs int `json:"total_uber_cost" bson:"total_uber_cost"`
  Total_uber_duration int `json:"total_uber_duration" bson:"total_uber_duration"`
  Total_distance float64 `json:"total_distance" bson:"total_distance"` 
  Uber_wait_time_eta int `json:"uber_wait_time_eta" bson:"uber_wait_time_eta"`

}
type Data struct{
id string
price int
duration int
distance float64
}
type coordinate struct {
    lat float64
    lng float64
}
type request struct {
    LocationIds            []string `json:"location_ids"`
    StartingFromLocationID string   `json:"starting_from_location_id"`
}
type Udata struct {
    Id bson.ObjectId `json:"id" bson:"_id"`
    Name string `json:"name" bson:"name"`
    Address string `json:"address" bson:"address"`
    City string `json:"city" bson:"city"`
    State string `json:"state" bson:"state"`
    Zip string `json:"zip" bson:"zip"`
    Coordinate struct {
        Lat float64 `json:"lat" bson:"lat"`
        Lng float64 `json:"lng" bson:"lng"`
    } `json:"coordinate" bson:"coordinate"`
}


func create(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u Udata
    URL := "http://maps.google.com/maps/api/geocode/json?address="

    json.NewDecoder(req.Body).Decode(&u)

    u.Id = bson.NewObjectId()

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
   if err != nil {
       fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = lat
    u.Coordinate.Lng = lng

    newSession().DB("nomankhan03").C("NewLocations").Insert(u)

    reply, _ := json.Marshal(u)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}

func get(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    uniqueid :=  p.ByName("uniqueid")

    if !bson.IsObjectIdHex(uniqueid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(uniqueid)

    responseObj := Udata{}

    if err := newSession().DB("nomankhan03").C("NewLocations").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", reply)
}

func update(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u Udata
    uniqueid :=  p.ByName("uniqueid")

    URL := "http://maps.google.com/maps/api/geocode/json?address="

    json.NewDecoder(req.Body).Decode(&u)

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
    if err != nil {
        fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = lat
    u.Coordinate.Lng = lng

    dataid := bson.ObjectIdHex(uniqueid)
    var data = Udata{
        Address: u.Address,
        City: u.City,
        State: u.State,
        Zip: u.Zip,
    }

    fmt.Println(data)

    newSession().DB("nomankhan03").C("NewLocations").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "address": u.Address,
        "city": u.City, "state": u.State,"zip": u.Zip, "coordinate.lat":u.Coordinate.Lat, "coordinate.lng":u.Coordinate.Lng}})

    responseObj := Udata{}

    if err := newSession().DB("nomankhan03").C("NewLocations").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}

func delete(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    uniqueid :=  p.ByName("uniqueid")

    if !bson.IsObjectIdHex(uniqueid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(uniqueid)

    if err := newSession().DB("nomankhan03").C("NewLocations").RemoveId(dataid); err != nil {
        rw.WriteHeader(404)
        return
    }

    rw.WriteHeader(200)
}

func newSession() *mgo.Session {

    s, err := mgo.Dial("mongodb://noman:hello123@ds057254.mongolab.com:57254/nomankhan03")

    if err != nil {
        panic(err)
    }
    return s
}

func getdetails(x string) (y coordinate) {
    responseObj := Udata{}

    if err := newSession().DB("nomankhan03").C("NewLocations").Find(bson.M{"_id": bson.ObjectIdHex(x)}).One(&responseObj); err != nil {
        z := coordinate{}
    return z
}
    p := coordinate{
    lat: responseObj.Coordinate.Lat,
    lng: responseObj.Coordinate.Lng,
    }
    return p
    
}

func getprice(x string, z string)(y Data){
response, err := http.Get(x)
    if err != nil {
        return
    }
    defer response.Body.Close()
    var price []int
    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
	panic(err)
        return
    }
    ptr := resp["prices"].([]interface{})
    jq := jsonq.NewQuery(resp)
     for i, _ := range ptr {
     pr,_ := jq.Int("prices",fmt.Sprintf("%d", i),"low_estimate")
     price = append(price,pr)
	}
     min := price[0]
     for j, _ := range price {
     if(price[j]<=min && price[j]!=0){
     min = price[j]
     }
     }
     du,_:=jq.Int("prices","0","duration")
     dist,_:=jq.Float("prices","0","distance")
     data := Data{
     id:z,
     price:min,
     duration:du,
     distance:dist,
     }
    return data     
}

func getpricetostart(x string)(y Data){
var price []int
response, err := http.Get(x)
    if err != nil {
        return
    }
    defer response.Body.Close()
    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }
    ptr := resp["prices"].([]interface{})
    jq := jsonq.NewQuery(resp)
     for i, _ := range ptr {
     pr,_ := jq.Int("prices",fmt.Sprintf("%d", i),"low_estimate")
     price = append(price,pr)
	}
     min := price[0]
     for j, _ := range price {
     if(price[j]<=min && price[j]!=0){
     min = price[j]
     }
     }
     du,_:=jq.Int("prices","0","duration")
     dist,_:=jq.Float("prices","0","distance")
     d := Data{
     id:"",
     price : min,
     duration:du,
     distance:dist,
}
return d
}

func (d dataSlice) Len() int {
	return len(d)
}


func (d dataSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}


func (d dataSlice) Less(i, j int) bool {
	return d[i].price < d[j].price 
}

func sortdata(x map[string]Data)(y Data) {
	m := x
	s := make(dataSlice, 0, len(m))
	for _, d := range m {
		s = append(s, d)
	}		
	sort.Sort(s)
	return s[0]
}


func deleteid(s []string, p string)(x []string) {
    var r []string
    for _, str := range s {
        if str != p {
            r = append(r, str)
        }
    }
    return r
}


func Sumfloat(a []float64) (sum float64) {
    for _, v := range a {
        sum += v
    }
    return
}

func Sumint(a []int) (sum int) {
    for _, v := range a {
        sum += v
    }
    return
}

func shortroute (rw http.ResponseWriter, req *http.Request, p httprouter.Params){
    decoder := json.NewDecoder(req.Body)
    var t request
    err := decoder.Decode(&t)
    if err != nil {
        panic(err)
    }
    Start := t.StartingFromLocationID
    LocIds := t.LocationIds
    var T Tdata
    var z coordinate
    var tp []int
    var td []float64
    var tdu []int

   for arraylength:=len(LocIds); arraylength>0; arraylength--{
    z = getdetails(Start)
    start_lat := z.lat
    start_lng := z.lng
    x := []coordinate{}
    for i := 0; i < len(LocIds); i++ {
       y := getdetails(LocIds[i])
       x = append(x,y)
   }
   tdata := map[string]Data{}
      for i:=0;i<len(x);i++{
      url := fmt.Sprintf("https://sandbox-api.uber.com/v1/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&server_token=2EOH3qEo3qmVVKJDPy6HaYKiEeyVmkB5-wIKo94_",start_lat,start_lng,x[i].lat,x[i].lng)
      d:=getprice(url, LocIds[i])
      tdata[LocIds[i]] = d
      }
   da:=sortdata(tdata)
  T.Best_route_location_ids = append(T.Best_route_location_ids,da.id)
   tp = append(tp,da.price)
   td = append(td,da.distance)
   tdu = append(tdu,da.duration)
   LocIds=deleteid(LocIds,da.id)
   Start=da.id
   }
   if(LocIds==nil){
   z = getdetails(Start)
    start_lat := z.lat
    start_lng := z.lng
    x := coordinate{}
    y := getdetails(t.StartingFromLocationID)
    x.lat=y.lat
    x.lng=y.lng
       tdata := map[string]Data{}
      url := fmt.Sprintf("https://sandbox-api.uber.com/v1/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&server_token=2EOH3qEo3qmVVKJDPy6HaYKiEeyVmkB5-wIKo94_",start_lat,start_lng,x.lat,x.lng)
      d:=getpricetostart(url)
      tdata[Start] = d
   tp = append(tp,d.price)
   td = append(td,d.distance)
   tdu = append(tdu,d.duration)
   }
   

T.Id = bson.NewObjectId()
T.Status = "Planning"
T.Starting_from_location_id= t.StartingFromLocationID
 T.Best_route_location_ids = T.Best_route_location_ids
T.Total_uber_costs = Sumint(tp)
T.Total_uber_duration = Sumint(tdu)
T.Total_distance = Sumfloat(td) 
newSession().DB("nomankhan03").C("OptimizedTrip").Insert(T)

    reply, _ := json.Marshal(T)
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)
        
    }

func gettrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    tripid :=  p.ByName("tripid")

    if !bson.IsObjectIdHex(tripid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(tripid)

    responseObj := Tdata{}

    if err := newSession().DB("nomankhan03").C("OptimizedTrip").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", reply)
}

func geteta(x float64,y float64,z string)(p int){
lat := strconv.FormatFloat(x, 'E', -1, 64)
lng := strconv.FormatFloat(y, 'E', -1, 64)
 url := "https://sandbox-api.uber.com/v1/requests"
    var jsonStr = []byte(`{
"start_latitude":"`+lat+`",
"start_longitude":"`+lng+`",
"product_id":"`+z+`",
}`)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsImhpc3RvcnkiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiZWY4MjcyNzItMWNkZC00NTUzLWE3NDQtMTlmOTBkY2FmM2NkIiwiaXNzIjoidWJlci11czEiLCJqdGkiOiIyOWE4NjJmNi1lMTczLTRlYTctOGViZS03NjkzZTVjYWZiZGEiLCJleHAiOjE0NTA3NjQzNTcsImlhdCI6MTQ0ODE3MjM1NiwidWFjdCI6IjU0RndzN2Rxc25xcHNKa0xUSXJXWTVwbEplaU1IMiIsIm5iZiI6MTQ0ODE3MjI2NiwiYXVkIjoiT0hQWGZ1NmFPOUcwcFFKc28zYW5rQ29DSnFpYTlIUVEifQ.dNxhW4Dm1woicq1yyMKxyiceXrmaEkBFS3HLqf-k_AVI0zODmQIctEcyl3jFb1-YApRlQXpiLHIvtYr2PBWZ0MtgpxuAreBF1AtYYmzmKoH3jxFbjGzlgsPrgEdcSVepuqL-645ffosE6Q4HmmG0KV7HI46BcbsACJjTDYilvoszj5ChJCpU60BLoYk9dNtZlCJ9EHk5J-09ktkODSuhRN99xH4YCe4FWfTppfe_IIIqlB1QeYwdd5cnHOnwKMVIJGjtxrXQ6UZ8wIb-De_Q04HUF3OyXZtPoUBlpnMkhaHa4PKUo8uC2jQf3Yd3s6-QUeUX_1BDsqrc1ns7fSt45g")
     req.Header.Set("Content-Type", "application/json")
var resp1 eta
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    err = json.Unmarshal(body,&resp1)
if err != nil {
  panic(err)
}

rid:= resp1.Eta
return rid

}

func triptrack(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
client := uber.NewClient(SERVER_TOKEN)
tripid :=  p.ByName("tripid")
    if !bson.IsObjectIdHex(tripid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(tripid)

    responseObj := Tdata{}

    if err := newSession().DB("nomankhan03").C("OptimizedTrip").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }
    if(nextid==""){
    startid =responseObj.Starting_from_location_id
     Locids =responseObj. Best_route_location_ids
    z := getdetails(responseObj.Starting_from_location_id)
    start_lat := z.lat
    start_lng := z.lng
    products,_ := client.GetProducts(start_lat,start_lng)
    productid := products[0].ProductID
    eta:=geteta(start_lat,start_lng,productid)
    nextid = Locids[0]
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :startid, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: eta,
     Status : "Requesting",
     Next_destination_location_id: nextid,
  }

  newSession().DB("nomankhan03").C("OptimizedTrip").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Requesting"}})
  startid = nextid
  Locids=deleteid(Locids,nextid)
  if(Locids!=nil){
  nextid = Locids[0]
  }else{
  nextid = "empty"
  }

res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(rw, "%s", res)
    }else if(Locids!=nil){
    if(nextid!="empty"){
    z := getdetails(startid)
    start_lat := z.lat
    start_lng := z.lng
products,_ := client.GetProducts(start_lat,start_lng)
productid := products[0].ProductID
eta:=geteta(start_lat,start_lng,productid)
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :startid, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: eta,
     Status : "Requesting",
     Next_destination_location_id: nextid,
     }
     newSession().DB("nomankhan03").C("OptimizedTrip").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Requesting"}})
     startid = nextid
  Locids=deleteid(Locids,nextid)
  if(Locids!=nil){
  nextid = Locids[0]
  }else{
  nextid = "empty"
  }
  res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(rw, "%s", res)
    }
    }else if(nextid=="empty"){
    z := getdetails(startid)
    start_lat := z.lat
    start_lng := z.lng
products,_ := client.GetProducts(start_lat,start_lng)
productid := products[0].ProductID
eta:=geteta(start_lat,start_lng,productid)
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :startid, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: eta,
     Status : "Requesting",
     Next_destination_location_id: responseObj.Starting_from_location_id,
     }
     newSession().DB("nomankhan03").C("OptimizedTrip").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Requesting"}})
     nextid="complete"
  res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(rw, "%s", res)
    }else{
    reply := Tstatus{
    Id:responseObj.Id,
    Starting_from_location_id :responseObj.Starting_from_location_id, 
    Best_route_location_ids:responseObj. Best_route_location_ids,
    Total_uber_costs:responseObj.Total_uber_costs,
    Total_uber_duration:responseObj.Total_uber_duration,
    Total_distance:responseObj.Total_distance,
    Uber_wait_time_eta: 0 ,
     Status : "Finished",
     Next_destination_location_id: "",
     }
     newSession().DB("nomankhan03").C("OptimizedTrip").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "status": "Finished"}})
     nextid=""
  res, _ := json.Marshal(reply)
    rw.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(rw, "%s", res)
    }
    }
  

