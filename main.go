package main

import (
	"fmt"
	"log"

	"github.com/AubSs/fasthttplogger"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

const pincodeURL = "https://data.gov.in/sites/default/files/all_india_PO_list_without_APS_offices_ver2.csv"

func main() {
	pincodes, err := getPincodes(pincodeURL)
	if err != nil {
		log.Fatalln(err)
	}

	hub := &Hub{
		pincodesByPincode:    map[string][]*Pincode{},
		pincodesByCityAndDis: map[string][]*Pincode{},
		pincodesByCity:       map[string][]*Pincode{},
	}

	// Put in a map
	for _, p := range pincodes {
		_, ok := hub.pincodesByPincode[p.Pincode]
		if !ok {
			hub.pincodesByPincode[p.Pincode] = []*Pincode{p}
			continue
		}
		hub.pincodesByPincode[p.Pincode] = append(hub.pincodesByPincode[p.Pincode], p)
	}

	// Populate cities
	for _, p := range pincodes {
		_, ok := hub.pincodesByCity[p.OfficeName]
		if !ok {
			hub.pincodesByCity[p.OfficeName] = []*Pincode{p}
		}
		hub.pincodesByCity[p.OfficeName] = append(hub.pincodesByCity[p.OfficeName], p)
	}

	// Populate cities and districts
	for _, p := range pincodes {
		name := fmt.Sprintf("%s:%s", p.District, p.OfficeName)
		_, ok := hub.pincodesByCityAndDis[name]
		if !ok {
			hub.pincodesByCityAndDis[name] = []*Pincode{p}
		}
		hub.pincodesByCityAndDis[name] = append(hub.pincodesByCityAndDis[name], p)
	}

	log.Println("Saved pincodes in map")

	router := fasthttprouter.New()
	router.GET("/", hub.Index)
	router.GET("/api/pincode/:pincode", hub.sendPincode)
	router.GET("/api/pincode", hub.sendPincodeByCityAndDis)

	s := &fasthttp.Server{
		Handler: fasthttplogger.Combined(router.Handler),
		Name:    "pincode_api",
	}

	log.Fatal(s.ListenAndServe(":8080"))
}
