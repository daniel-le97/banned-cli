package main

import (
	"fmt"
	"os"

	"github.com/daniel-le97/banned-cli/cmd"
)

func testDatabase() {
	// Test database connection and query
	fmt.Println("Testing database connection...")

	db, err := cmd.GetDB()
	if err != nil {
		fmt.Printf("Failed to get database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Testing settings query...")
	rows, err := db.Query("SELECT key, value, updated_at FROM settings ORDER BY key")
	if err != nil {
		fmt.Printf("Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var key, value, updatedAt string
		if err := rows.Scan(&key, &value, &updatedAt); err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}
		fmt.Printf("Row %d: %s = %s (updated: %s)\n", count+1, key, value, updatedAt)
		count++
	}

	fmt.Printf("Total rows: %d\n", count)
}
