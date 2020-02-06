package main

import (
	"fmt"
	"log"
	"sync"

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

// Hub contains all global context
type Hub struct {
	sync.RWMutex
	pincodesByPincode    map[string][]*Pincode
	pincodesByCity       map[string][]*Pincode
	pincodesByCityAndDis map[string][]*Pincode
}

// NewHub initializes and returns a new instance of hub
func NewHub() (*Hub, error) {
	hub := &Hub{
		pincodesByPincode:    map[string][]*Pincode{},
		pincodesByCityAndDis: map[string][]*Pincode{},
		pincodesByCity:       map[string][]*Pincode{},
	}
	err := hub.RefreshPincodes()
	if err != nil {
		return nil, err
	}

	return hub, nil
}

// RefreshPincodes refreshes pincodes
func (h *Hub) RefreshPincodes() error {
	log.Println("started refreshing pincodes...")
	pincodes, err := getPincodes(pincodeURL)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("pincodes downloaded...")

	h.Lock()
	defer h.Unlock()

	// Put in a map
	for _, p := range pincodes {
		_, ok := h.pincodesByPincode[p.Pincode]
		if !ok {
			h.pincodesByPincode[p.Pincode] = []*Pincode{p}
			continue
		}
		h.pincodesByPincode[p.Pincode] = append(h.pincodesByPincode[p.Pincode], p)
	}

	// Populate cities
	for _, p := range pincodes {
		_, ok := h.pincodesByCity[p.OfficeName]
		if !ok {
			h.pincodesByCity[p.OfficeName] = []*Pincode{p}
		}
		h.pincodesByCity[p.OfficeName] = append(h.pincodesByCity[p.OfficeName], p)
	}

	// Populate cities and districts
	for _, p := range pincodes {
		name := fmt.Sprintf("%s:%s", p.District, p.OfficeName)
		_, ok := h.pincodesByCityAndDis[name]
		if !ok {
			h.pincodesByCityAndDis[name] = []*Pincode{p}
		}
		h.pincodesByCityAndDis[name] = append(h.pincodesByCityAndDis[name], p)
	}

	return nil
}

// Index just says welcome
func (h *Hub) Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Welcome/n")
}

func (h *Hub) sendPincode(ctx *fasthttp.RequestCtx) {
	pincode := ctx.UserValue("pincode").(string)

	// Check map
	h.RLock()
	p, ok := h.pincodesByPincode[pincode]
	if !ok {
		e := &ErrorMessage{}
		ctx.SetStatusCode(404)
		ctx.SetContentType("application/json")
		ctx.WriteString(e.makeError("Pincode doesn't exist.", 404))
		return
	}
	h.RUnlock()

	resp := Response{Status: fasthttp.StatusOK, Pincode: p}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(&resp)
}

func (h *Hub) sendPincodeByCityAndDis(ctx *fasthttp.RequestCtx) {
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

	citySuffixes := []string{"H.O", "S.O", "B.O"}

	h.RLock()

	if district == "" {
		for _, cs := range citySuffixes {
			name := fmt.Sprintf("%s %s", city, cs)
			p, ok := h.pincodesByCity[name]
			if !ok {
				continue
			}
			pincodes = append(pincodes, p...)
		}
		goto final
	}

	for _, cs := range citySuffixes {
		p, ok := h.pincodesByCity[fmt.Sprintf("%s:%s %s", district, city, cs)]
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
	h.Unlock()
	resp := Response{Status: fasthttp.StatusOK, Pincode: pincodes}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(&resp)
}
