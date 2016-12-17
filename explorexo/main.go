package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jimmyjames85/sand/explorexo/models"
	"github.com/kelseyhightower/envconfig"
	"os/exec"
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

func getJim(db *sql.DB) (*models.Person, error) {
	jim, err := models.PersonByFirstMiddleLast(db, "James", "Worthington", "Gordon")
	if err != nil {
		fmt.Printf("Jim does not exist: %v\n", err)
		fmt.Printf("Creating him now\n")
		jim = &models.Person{First: "James", Middle: "Worthington", Last: "Gordon"}
		fmt.Printf("Before save: Struct.Exists() = %v\n", jim.Exists())
		err = jim.Save(db)
		if err != nil {
			return nil, err
		}
		fmt.Printf("After save: Struct.Exists() = %v\n", jim.Exists())
	}
	return jim, nil
}

func displayGotham(db *sql.DB){
	names, err := models.GetEgos(db)
	if err != nil {
		log.Fatalf("GetGetAllNames: %v", err)
	}
	for i, n := range names {
		fmt.Printf("%02d:%02d: [%s]\t%s, %s %s\n", i, n.ID, n.Ego, n.Last, n.First, n.Middle)
	}
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
	defer db.Close()

	displayGotham(db)
	jim, err := getJim(db)
	if err != nil {
		log.Fatal("failed to get James Worthington Gordon: %v\n")
	}

	fmt.Printf("Deleting Jim manually (id=%d)\n", jim.ID)

	r, err := db.Exec("DELETE from person where id=?", jim.ID)
	if err != nil {
		log.Fatal("failed to delete: %v", err)
	}

	rc, err := r.RowsAffected()
	if err != nil {
		log.Fatal("failed to get rowsaffected: %v", err)
	}
	fmt.Printf("removed %d row\n", rc)
	fmt.Printf("After delete: Struct.Exists() = %v\n", jim.Exists())
	fmt.Printf("Attempting to update jim with struct\n")
	err  = jim.Update(db)
	if err!=nil{
		log.Fatalf("failed to update: %v\n", err)
	}
	displayGotham(db)


	//if len(os.Args) > 1 {
	//	jim.Middle = os.Args[1]
	//	err = jim.Update(db)
	//	if err != nil {
	//		log.Fatalf("failed to update jim: %v\n", err)
	//	}
	//}
}
