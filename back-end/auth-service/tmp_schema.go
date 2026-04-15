package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "desarrollo:test_24*@tcp(192.168.90.32:3306)/bdsaocomco_inventario_carrito"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("DESCRIBE registros_inventario")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, typ, null, key string
		var def sql.NullString
		var ext sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &def, &ext); err != nil {
			log.Fatal(err)
		}
		if field == "novedad" || field == "accion_faltante" {
			fmt.Printf("Field: %s, Type: %s, Null: %s\n", field, typ, null)
		}
	}
}
