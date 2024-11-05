package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type PageData struct {
	Visits int64
}

func main() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisURL,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
	})

	// Check Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Cannot connect to Redis: %v", err)
	}

	tmpl := template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Hello World</title>
    <link rel="icon" href="data:,"> <!-- Prevents favicon request -->
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 40px auto;
            max-width: 650px;
            line-height: 1.6;
            font-size: 18px;
            padding: 0 10px;
        }
        .counter {
            color: #333;
            font-size: 24px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <h1>Hello World</h1>
    <div class="counter">Visit count: {{.Visits}}</div>
</body>
</html>
`))

	// Handle favicon.ico requests
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // Return 204 No Content
	})

	// Handle main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only count visits for the main page
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		visits, err := rdb.Incr(ctx, "visits").Result()
		if err != nil {
			log.Printf("Error incrementing visit counter: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := PageData{
			Visits: visits,
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	addr := ":7001"
	fmt.Printf("Server is running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
