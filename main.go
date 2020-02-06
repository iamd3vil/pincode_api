package main

import (
	"log"

	"github.com/AubSs/fasthttplogger"
	"github.com/buaazp/fasthttprouter"
	"github.com/jasonlvhit/gocron"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	address    = ":8080"
	pincodeURL = "https://data.gov.in/sites/default/files/all_india_PO_list_without_APS_offices_ver2.csv"
)

func main() {

	hub, err := NewHub()
	if err != nil {
		log.Fatalf("error while initializing hub: %v", err)
	}
	log.Println("Saved pincodes in map")

	go func() {
		gocron.Every(2).Hours().DoSafely(func() {
			log.Println("Starting to refresh pincodes")
			hub.RefreshPincodes()
			log.Println("Refreshed pincodes")
		})

		<-gocron.Start()
	}()

	router := fasthttprouter.New()
	router.GET("/", hub.Index)
	router.GET("/api/pincode/:pincode", hub.sendPincode)
	router.GET("/api/pincode", hub.sendPincodeByCityAndDis)

	s := &fasthttp.Server{
		Handler: fasthttplogger.Combined(router.Handler),
		Name:    "pincode_api",
	}

	log.Fatal(s.ListenAndServe(address))
}
