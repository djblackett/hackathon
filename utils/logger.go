package utils

import (
	"log"
	"os"
)

var lg *log.Logger

func init() {
	f, _ := os.OpenFile("rename.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	lg = log.New(f, "", log.LstdFlags)
}

func LogPrompt(file, prompt, response string) {
	lg.Printf("FILE: %s\nPROMPT:\n%s\nRESPONSE:\n%s\n---\n", file, prompt, response)
}
