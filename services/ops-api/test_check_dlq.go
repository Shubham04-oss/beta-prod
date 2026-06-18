package main
import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil { log.Fatal(err) }
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT status, error_message FROM sync_failures_dlq ORDER BY created_at DESC LIMIT 1")
	if err != nil { log.Fatal(err) }
	defer rows.Close()

	if rows.Next() {
		var status, errMsg string
		rows.Scan(&status, &errMsg)
		fmt.Printf("DLQ Entry Found - Status: %s, Error: %s\n", status, errMsg)
	} else {
		fmt.Println("No DLQ entries found")
	}
}
