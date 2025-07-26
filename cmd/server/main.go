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
	// Load environment variables (API keys stay on server)
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	cfg := config.FromEnv()

	http.HandleFunc("/suggest-filename", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req FilenameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Create AI client (always use OpenAI on server)
		client, err := ai.NewClient(cfg, false, req.Model)
		if err != nil {
			resp := FilenameResponse{Error: err.Error()}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		suggested, err := client.SuggestFilename(req.Content)
		resp := FilenameResponse{Filename: suggested}
		if err != nil {
			resp.Error = err.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("AI filename server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}




// make changes to client and AI code to support server-side AI filename suggestions