package main

import (
	"log"
	"split/config"
	"split/postgres"
	"split/splitwise/rest"
	"split/splitwise/split"
)

func main() {
	log.Println("------------SPLIT  Starting------------")
	config.LoadConfig()
	if err := postgres.Connect(); err != nil {
		log.Panic("postgreSQL error : ", err)
	} else {
		log.Println("postgreSQL connected")
	}
	split.Init()
	rest.Init()

}
