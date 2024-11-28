package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() {
	var err error
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbPass == "" || dbHost == "" || dbName == "" {
		log.Fatal("Database credentials are not set in environment variables")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", dbUser, dbPass, dbHost, dbName)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
}

func getTorontoTime() (time.Time, error) {
	location, err := time.LoadLocation("America/Toronto")
	if err != nil {
		return time.Time{}, err
	}
	return time.Now().In(location), nil
}

func currentTimeHandler(w http.ResponseWriter, r *http.Request) {
	torontoTime, err := getTorontoTime()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting Toronto time: %v", err)
		return
	}

	// Log the current time to the database
	_, err = db.Exec("INSERT INTO time_log (timestamp) VALUES (?)", torontoTime)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error logging time to database: %v", err)
		return
	}

	// Return the current time in JSON format
	response := map[string]string{
		"current_time": torontoTime.Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func timeLogsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, timestamp FROM time_log")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error querying time logs: %v", err)
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id int
		var timestampStr string // Retrieve the `timestamp` as a string
		if err := rows.Scan(&id, &timestampStr); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Parse the string timestamp into a Go `time.Time`
		timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			continue
		}

		// Append the result to the logs slice
		logs = append(logs, map[string]interface{}{
			"id":        id,
			"timestamp": timestamp.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func setupLogging() {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}

func main() {
	setupLogging()
	initDB()
	defer db.Close()

	http.HandleFunc("/current-time", currentTimeHandler)
	http.HandleFunc("/time-logs", timeLogsHandler)

	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
