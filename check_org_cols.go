package main
import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)
func main() {
	dbpool, _ := pgxpool.New(context.Background(), "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	rows, _ := dbpool.Query(context.Background(), "SELECT column_name FROM information_schema.columns WHERE table_name='organizations'")
	defer rows.Close()
	for rows.Next() {
		var col string
		rows.Scan(&col)
		fmt.Println(col)
	}
}
