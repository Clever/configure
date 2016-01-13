package main

import (
	"log"

	"github.com/Clever/configure"
)

var config struct {
	DistrictID string `config:"district_id,required"`
	Collection string `config:"collection"`
}

func main() {
	if err := configure.Configure(&config); err != nil {
		log.Fatalf("err: %s", err)
	}
	log.Printf("config: %+v", config)
}
