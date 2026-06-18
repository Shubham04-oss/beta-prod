package main
import (
	"context"
	"fmt"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
)
func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db")
	if err != nil { log.Fatal(err) }
	rows, err := dbpool.Query(context.Background(), "SELECT enumlabel FROM pg_enum JOIN pg_type ON pg_enum.enumtypid = pg_type.oid WHERE typname = 'order_status';")
	if err != nil { log.Fatal(err) }
	defer rows.Close()
	for rows.Next() {
		var label string
		rows.Scan(&label)
		fmt.Println(label)
	}
}
