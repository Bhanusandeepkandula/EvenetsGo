package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

type Event struct {
	ID           string `json:"id"`
	EventName    string `json:"eventName"`
	CustomerName string `json:"customerName"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	DataType     string `json:"dataType"`
	CreatedBy    string `json:"createdBy"`
	Paid         int64  `json:"paid"`
	Balance      int64  `json:"balance"`
	TotalCost    int64  `json:"totalCost"`
	Status       string `json:"status"`
	Venue        string `json:"venue"`
	DateTime     string `json:"dateTime"`
	CreatedAt    string `json:"createdAt"`
	GcalId       string `json:"gcalId"`
}

func main() {
	// 1. Read JSON file
	data, err := ioutil.ReadFile("events.json")
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	// 2. Parse as generic map first to handle nested structure
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	fmt.Println("\n=== Analyzing JSON Structure ===")

	// Extract events section
	eventsSection, ok := rawData["events"].(map[string]interface{})
	if !ok {
		log.Fatalf("No 'events' section found in JSON")
	}

	fmt.Printf("Found %d items in events section\n", len(eventsSection))

	// Parse each event manually, skipping non-event entries
	events := make(map[string]Event)
	for key, value := range eventsSection {
		// Skip non-event entries (like "users")
		if key == "users" || key == "auditLogs" {
			fmt.Printf("⊘ Skipping non-event key: %s\n", key)
			continue
		}

		// Convert back to JSON and parse as Event
		eventJSON, err := json.Marshal(value)
		if err != nil {
			log.Printf("⚠ Warning: Failed to marshal key %s: %v", key, err)
			continue
		}

		var event Event
		if err := json.Unmarshal(eventJSON, &event); err != nil {
			log.Printf("⚠ Warning: Failed to parse key %s as Event: %v", key, err)
			continue
		}

		// Validate it's actually an event (has event-like fields)
		if event.EventName == "" && event.TotalCost == 0 {
			fmt.Printf("⊘ Skipping non-event entry: %s\n", key)
			continue
		}

		events[key] = event
	}

	fmt.Printf("\n✓ Found %d valid events to import\n\n", len(events))

	// 3. Connect to Postgres
	connStr := "host=localhost port=5432 user=postgres password=Sandeep@123 dbname=eventplanner sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to Postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging Postgres: %v", err)
	}

	var dbname string
	db.QueryRow("SELECT current_database()").Scan(&dbname)
	fmt.Println("Importer is connected to database:", dbname)
	fmt.Println("✅ Connected to Postgres\n")

	// 4. Prepare INSERT statement
	stmt, err := db.Prepare(`
        INSERT INTO events (
            id,
            event_name,
            customer_name,
            phone,
            address,
            data_type,
            created_by,
            paid,
            balance,
            total_cost,
            status,
            venue,
            date_time,
            created_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        ON CONFLICT (id) DO NOTHING
    `)
	if err != nil {
		log.Fatalf("Error preparing statement: %v", err)
	}
	defer stmt.Close()

	// 5. Insert all events
	count := 0
	skipped := 0
	for key, e := range events {
		// if ID is empty, fall back to the Firebase key
		id := e.ID
		if id == "" {
			id = key
		}

		result, err := stmt.Exec(
			id,
			e.EventName,
			e.CustomerName,
			e.Phone,
			e.Address,
			e.DataType,
			e.CreatedBy,
			e.Paid,
			e.Balance,
			e.TotalCost,
			e.Status,
			e.Venue,
			e.DateTime,
			e.CreatedAt,
		)
		if err != nil {
			log.Printf("❌ Insert error for id=%s: %v", id, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			count++
			customerName := e.CustomerName
			if customerName == "" {
				customerName = "(no name)"
			}
			fmt.Printf("✓ Inserted: %s - %s - %s (%s)\n",
				id, customerName, e.EventName, strings.TrimSpace(e.DateTime))
		} else {
			skipped++
			fmt.Printf("⊘ Skipped (duplicate): %s\n", id)
		}
	}

	fmt.Printf("\n🎉 Import Complete!\n")
	fmt.Printf("   Inserted: %d events\n", count)
	fmt.Printf("   Skipped: %d events (already in database)\n", skipped)
	fmt.Printf("   Total valid events in JSON: %d\n", len(events))
}
