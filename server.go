package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/AubSs/fasthttplogger"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// ErrorMessage is the struct for error messages
type ErrorMessage struct {
	Status  int32  `json:"status"`
	Message string `json:"message"`
}

func (error *ErrorMessage) makeError(message string, status int32) string {
	error.Status = status
	error.Message = message
	resp, _ := json.Marshal(error)
	return string(resp)
}

// Response descibes the response for all api responses
type Response struct {
	Status  int32      `json:"status"`
	Pincode []*Pincode `json:"data"`
}

// server contains all global context
type server struct {
	sync.RWMutex
	pincodesByPincode    map[string][]*Pincode
	pincodesByCity       map[string][]*Pincode
	pincodesByCityAndDis map[string][]*Pincode
	router               *router.Router
}

// newServer initializes and returns a new instance of hub
func newServer() (*server, error) {
	router := router.New()
	hub := &server{
		pincodesByPincode:    map[string][]*Pincode{},
		pincodesByCityAndDis: map[string][]*Pincode{},
		pincodesByCity:       map[string][]*Pincode{},
		router:               router,
	}
	err := hub.RefreshPincodes()
	if err != nil {
		return nil, err
	}

	return hub, nil
}

// ListenAndServer starts the HTTP server
func (srv *server) ListenAndServe(ctx context.Context, address string) error {
	srv.routes()
	s := &fasthttp.Server{
		Handler: fasthttplogger.Combined(srv.router.Handler),
		Name:    "pincode_api",
	}

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			log.Println("exiting...")
			s.Shutdown()
			os.Exit(0)
		}
	}(ctx)
	return s.ListenAndServe(address)
}

// RefreshPincodes refreshes pincodes
func (srv *server) RefreshPincodes() error {
	log.Println("started refreshing pincodes...")
	pincodes, err := getPincodes(pincodeURL)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("pincodes downloaded...")

	srv.Lock()
	defer srv.Unlock()

	// Put in a map
	for _, p := range pincodes {
		_, ok := srv.pincodesByPincode[p.Pincode]
		if !ok {
			srv.pincodesByPincode[p.Pincode] = []*Pincode{p}
			continue
		}
		srv.pincodesByPincode[p.Pincode] = append(srv.pincodesByPincode[p.Pincode], p)
	}

	// Populate cities
	for _, p := range pincodes {
		_, ok := srv.pincodesByCity[p.OfficeName]
		if !ok {
			srv.pincodesByCity[p.OfficeName] = []*Pincode{p}
		}
		srv.pincodesByCity[p.OfficeName] = append(srv.pincodesByCity[p.OfficeName], p)
	}

	// Populate cities and districts
	for _, p := range pincodes {
		name := fmt.Sprintf("%s:%s", p.District, p.OfficeName)
		_, ok := srv.pincodesByCityAndDis[name]
		if !ok {
			srv.pincodesByCityAndDis[name] = []*Pincode{p}
		}
		srv.pincodesByCityAndDis[name] = append(srv.pincodesByCityAndDis[name], p)
	}

	return nil
}

// Index just says welcome
func (srv *server) Index() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		fmt.Fprintf(ctx, "Welcome\n")
	}
}

func (srv *server) handleHealth() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("OK")
	}
}

func (srv *server) sendPincode() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		pincode := ctx.UserValue("pincode").(string)

		// Check map
		srv.RLock()
		p, ok := srv.pincodesByPincode[pincode]
		if !ok {
			e := &ErrorMessage{}
			ctx.SetStatusCode(404)
			ctx.SetContentType("application/json")
			ctx.WriteString(e.makeError("Pincode doesn't exist.", 404))
			return
		}
		srv.RUnlock()

		resp := Response{Status: fasthttp.StatusOK, Pincode: p}
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(&resp)
	}
}

func (srv *server) sendPincodeByCityAndDis() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		city := fmt.Sprintf("%s", ctx.QueryArgs().Peek("city"))
		district := fmt.Sprintf("%s", ctx.QueryArgs().Peek("district"))
		if city == "" {
			e := &ErrorMessage{}
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetContentType("application/json")
			ctx.WriteString(e.makeError("City can't be blank.", fasthttp.StatusBadRequest))
			return
		}

		pincodes := []*Pincode{}

		citySuffixes := []string{"srv.O", "S.O", "B.O"}

		srv.RLock()

		if district == "" {
			for _, cs := range citySuffixes {
				name := fmt.Sprintf("%s %s", city, cs)
				p, ok := srv.pincodesByCity[name]
				if !ok {
					continue
				}
				pincodes = append(pincodes, p...)
			}
			goto final
		}

		for _, cs := range citySuffixes {
			p, ok := srv.pincodesByCity[fmt.Sprintf("%s:%s %s", district, city, cs)]
			if !ok {
				continue
			}
			pincodes = append(pincodes, p...)
		}

	final:
		if len(pincodes) == 0 {
			e := &ErrorMessage{}
			ctx.SetStatusCode(404)
			ctx.SetContentType("application/json")
			ctx.WriteString(e.makeError("Pincode doesn't exist.", 404))
			return
		}
		srv.RUnlock()
		resp := Response{Status: fasthttp.StatusOK, Pincode: pincodes}
		ctx.SetContentType("application/json")
		json.NewEncoder(ctx).Encode(&resp)
	}
}
