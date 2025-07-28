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
ome things are only actually, some potentially, some potentially and actually, what they are, viz. in one case a particular 
reality, in another, characterized by a particular quantity, or the like. There is no movement apart from things; for change is 
always according to the categories of being, and there is nothing common to these and in no one category. But each of the 
categories belongs to all its subjects in either of two ways (e.g. 'this-ness'-for one kind of it is 'positive form', and the other is 
'privation'); and as regards quality one kind is 'white' and the other 'black', and as regards quantity one kind is 'complete' and 
the other 'incomplete', and as regards spatial movement one is 'upwards' and the other 'downwards', or one thing is 'light' and 
another 'heavy'); so that there are as many kinds of movement and change as of being. There being a distinction in each 
class of things between the potential and the completely real, I call the actuality of the potential as such, movement.
"""
filename: aristotle-movement-change-categories-being

Example 3:
"""
Shall I compare thee to a summer's day?
Thou art more lovely and more temperate:
Rough winds do shake the darling buds of May,
And summer's lease hath all too short a date;
Sometime too hot the eye of heaven shines,
And often is his gold complexion dimm'd;
And every fair from fair sometime declines,
By chance or nature's changing course untrimm'd;
But thy eternal summer shall not fade,
Nor lose possession of that fair thou ow'st;
Nor shall death brag thou wander'st in his shade,
When in eternal lines to time thou grow'st:
   So long as men can breathe or eyes can see,
   So long lives this, and this gives life to thee.
"""
filename: shakespeare-sonnet-18-summer-day

Example 4:
"""
Hey!
Just wanted to check in — I was thinking about you earlier today for some reason. I think it's 'cause I passed that little café we went to last spring, the one with the weird plant wall and the cinnamon rolls that were, like, offensively good? Anyway, made me laugh.

How's your week going? Things here have been a bit of a blur — work's been nonstop, and I finally finished that project I was ranting about. You would've been proud of the amount of coffee it took.

Also, random: I saw a dog today that looked exactly like yours, but somehow even more dramatic. It barked at a leaf and then looked genuinely offended when it moved.

Anyway, no pressure to reply fast or anything — just wanted to say hey and that I hope everything's going okay on your end. Talk soon?
"""
filename: casual-check-in-funny-dog-story
`
