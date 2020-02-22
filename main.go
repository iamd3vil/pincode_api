package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jasonlvhit/gocron"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	defaultAddress = ":8080"
	pincodeURL     = "https://data.gov.in/sites/default/files/all_india_PO_list_without_APS_offices_ver2.csv"
)

func main() {
	addr := flag.String("addr", defaultAddress, "HTTP Address")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigs
		cancel()
	}()

	var (
		err error
		srv *server
	)
	for {
		srv, err = newServer()
		if err != nil {
			log.Fatalf("error while initializing hub: %v. Retrying....", err)
			continue
		}
		log.Println("Saved pincodes in map")
		break
	}

	go func() {
		gocron.Every(2).Hours().DoSafely(func() {
			log.Println("Starting to refresh pincodes")
			srv.RefreshPincodes()
			log.Println("Refreshed pincodes")
		})

		<-gocron.Start()
	}()

	err = srv.ListenAndServe(ctx, *addr)
	if err != nil {
		log.Fatalln(err)
	}
}
