package main

import (
	"encoding/json"
	"fmt"

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
	pincodesByPincode    map[string][]*Pincode
	pincodesByCity       map[string][]*Pincode
	pincodesByCityAndDis map[string][]*Pincode
}

// Index just says welcome
func (h *Hub) Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Welcome/n")
}

func (h *Hub) sendPincode(ctx *fasthttp.RequestCtx) {
	pincode := ctx.UserValue("pincode").(string)

	// Check map
	p, ok := h.pincodesByPincode[pincode]
	if !ok {
		e := &ErrorMessage{}
		ctx.SetStatusCode(404)
		ctx.SetContentType("application/json")
		ctx.WriteString(e.makeError("Pincode doesn't exist.", 404))
		return
	}

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
	resp := Response{Status: fasthttp.StatusOK, Pincode: pincodes}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(&resp)
}
