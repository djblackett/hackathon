package ai

import "fmt"

func buildPrompt(content string) string {
	return fmt.Sprintf(`You are a file recovery assistant.
A document was recovered from a damaged hard drive, but its filename was lost.

Here is the document:
"""
%s
"""

Generate a single, meaningful filename that summarizes this file.
• 5–10 words
• lowercase letters, numbers, dashes or underscores
• no file extension
Respond with the filename only.`, content)
}
