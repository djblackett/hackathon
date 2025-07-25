package ai

import "fmt"

func buildPrompt(content string) string {
	return fmt.Sprintf(`
	%s
	You are a file recovery assistant.
A document was recovered from a damaged hard drive, but its filename was lost.

Here is the document:
"""
%s
"""

Generate a single, meaningful filename that summarizes this file.
• 5-10 words
• lowercase letters, numbers, dashes or underscores
• no file extension
Respond with the filename only.

Respond **with one line only** - a single filename and nothing else. Do **not** return multiple suggestions or bullet points.
Return one filename **without generic words such as _document, file, draft, text, note, notes_.
Return one filename only, avoid repeating the same word, 5-8 words max…
`, fewShot, content)
}

const fewShot = `
Example 1:
"""
prepare slides for jazz history class; focus on modal jazz; miles davis kind of blue 1959 sessions
"""
filename: miles-davis-kind-of-blue-jazz-history

Example 2:
"""
hello there
general kenobi
roger roger
"""
filename: hello-general-kenobi-star-wars-reference
`
