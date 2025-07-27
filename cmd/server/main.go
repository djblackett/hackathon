package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/djblackett/bootdev-hackathon/internal/ai"
	"github.com/djblackett/bootdev-hackathon/internal/config"
	"github.com/joho/godotenv"
)

type FilenameRequest struct {
	Content string `json:"content"`
	Model   string `json:"model"`
}

type FilenameResponse struct {
	Filename string `json:"filename"`
	Error    string `json:"error,omitempty"`
}

func main() {
	log.Println("Starting AI filename server...")

	// Load environment variables (API keys stay on server)
	_ = godotenv.Load()
	log.Println("Environment variables loaded")

	cfg := config.FromEnv()
	log.Printf("Configuration loaded: %+v", cfg)

	http.HandleFunc("/suggest-filename", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request to /suggest-filename from %s", r.Method, r.RemoteAddr)

		if r.Method != http.MethodPost {
			log.Printf("Method not allowed: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req FilenameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode JSON request: %v", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("Processing request with model: %s, content length: %d", req.Model, len(req.Content))

		// Create AI client (always use OpenAI on server)
		client, err := ai.NewClient(cfg, false, req.Model)
		if err != nil {
			log.Printf("Failed to create AI client: %v", err)
			resp := FilenameResponse{Error: err.Error()}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		log.Println("AI client created successfully")

		suggested, err := client.SuggestFilename(req.Content)
		resp := FilenameResponse{Filename: suggested}
		if err != nil {
			log.Printf("AI suggestion failed: %v", err)
			resp.Error = err.Error()
		} else {
			log.Printf("AI suggested filename: %s", suggested)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		log.Println("Response sent successfully")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Using port: %s", port)

	log.Printf("AI filename server starting on port %s", port)
	log.Printf("Server binary: relay-server (from cmd/server/main.go)")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
