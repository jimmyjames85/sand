package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/sand/explorexo/models"
	"github.com/knq/dburl"
)

func main() {
	db, err := dburl.Open("mysql://user:pass@hostname:port/database")
	if err != nil {
		log.Fatal(err)
	}

	names, err := models.GetListNames(db)
	if err != nil {
		log.Fatal(err)
	}

	for i, n := range names {
		log.Printf("%02d: %s\n", i, n.Lastname)
	}

	newname := models.Name{
		Lastname:  sql.NullString{String: "Harvey", Valid:true},
		Firstname: sql.NullString{String: "Dent", Valid:true},
	}

	if err = newname.Save(db); err != nil {
		log.Fatal(err)
	}

}
