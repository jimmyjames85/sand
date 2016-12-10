package main

import (
	"log"
	"fmt"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/kelseyhightower/envconfig"
	"github.com/jimmyjames85/sand/explorexo/models"
	"os"
)

type dbConfig struct {
	Username string `envconfig:"USERNAME"`
	Password string `envconfig:"PASSWORD"`
	Hostname string `envconfig:"HOSTNAME"`
	Port     int    `envconfig:"PORT"`
	Database string `envconfig:"DATABASE"`
}

func (c *dbConfig) SourceName() string {

	//return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", c.Username, c.Password, c.Hostname, c.Port, c.Database)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Hostname, c.Port, c.Database)

}

func main() {

	dbconf := &dbConfig{}
	err := envconfig.Process("DB", dbconf)
	if err != nil {
		log.Fatal("unable to load environment variables: %v", err)
	}

	db, err := sql.Open("mysql", dbconf.SourceName())
	//db, err := durl.Open(dbconf.SourceName())
	if err != nil {
		log.Fatal(err)
	}

	names, err := models.GetEgos(db)
	if err != nil {
		log.Fatalf("GetGetAllNames: %v", err)
	}

	for i, n := range names {
		fmt.Printf("%02d:%02d: [%s]\t%s, %s %s\n", i, n.ID, n.Ego, n.Last, n.First, n.Middle)
	}

	jim, err := models.PersonByID(db, 19)
	if err !=nil{
		log.Fatalf("jim: %v\n", err)
	}

	fmt.Printf("I found him: %v\n", jim)

	if len(os.Args) >1{
		jim.Middle = os.Args[1]
		err = jim.Update(db)
		if err!=nil{
			log.Fatalf("failed to update jim: %v\n", err)
		}
	}
}
