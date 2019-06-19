package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
)

// Pincode represents a single pincode
type Pincode struct {
	Pincode        string `json:"pincode"`
	OfficeName     string `json:"office_name"`
	DeliveryStatus string `json:"delivery_status"`
	DivisionName   string `json:"division_name"`
	RegionName     string `json:"region_name"`
	CircleName     string `json:"circle_name"`
	District       string `json:"district"`
	StateName      string `json:"state_name"`
	Taluk          string `json:"taluk"`
}

func getPincodes(url string) ([]*Pincode, error) {
	pincodes := []*Pincode{}
	// Download csv file and parse it
	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if req.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with error: %v", req.StatusCode)
	}

	// body, err := ioutil.ReadAll(req.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// log.Println(body)

	// Parse CSV
	r := csv.NewReader(req.Body)
	defer req.Body.Close()

	skipHeader := true

	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if skipHeader {
			skipHeader = false
			continue
		}
		pincode := &Pincode{
			CircleName:     line[6],
			RegionName:     line[5],
			DivisionName:   line[4],
			OfficeName:     line[0],
			Pincode:        line[1],
			DeliveryStatus: line[3],
			District:       line[8],
			StateName:      line[9],
			Taluk:          line[7],
		}
		pincodes = append(pincodes, pincode)
	}

	return pincodes, nil
}
