package state

import (
	"regexp"
	"sort"
	"unicode/utf16"

	"github.com/sourcegraph/go-lsp"
)

type ReverseOrder []lsp.TextDocumentContentChangeEvent

func (a ReverseOrder) Len() int      { return len(a) }
func (a ReverseOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ReverseOrder) Less(i, j int) bool {
	l1 := a[i].Range.Start.Line
	l2 := a[j].Range.Start.Line
	if l1 > l2 {
		return true
	} else if l2 > l1 {
		return false
	} else {
		return a[i].Range.Start.Character > a[j].Range.Start.Character
	}
}

func insertLines(lines []string, newLines []string, index int) []string {
	if len(newLines) == 0 {
		return lines
	}
	linesLen := len(lines)
	// First put all newLines at the end
	lines = append(lines, newLines...)
	// If we're inserting at the end, we're done!
	if index == linesLen {
		return lines
	}
	copy(lines[index+len(newLines):], lines[index:])
	for i, line := range newLines {
		lines[index+i] = line
	}
	return lines
}

var lineRE = regexp.MustCompile(`\r\n|\r|\n`)

func SplitLines(text string) []string {
	return lineRE.Split(text, -1)
}

// Converts UTF-16 character offsets to byte offsets for easy string manipulation
func CharIndexToByte(line string, character int) int {
	u16 := utf16.Encode([]rune(line))[0:character]
	return len(string(utf16.Decode(u16)))
}

func CharIndexToRune(line string, character int) int {
	u16 := utf16.Encode([]rune(line))[0:character]
	return len(utf16.Decode(u16))
}

func applyTextChanges(text []string, changes []lsp.TextDocumentContentChangeEvent) ([]string, []ChangeTextRange) {
	// NOTE: We know that this function is ONLY called if all of the change.Range fields are non-nil

	// First sort in reverse order so we can make the changes without worrying about changing the indexes
	sort.Sort(ReverseOrder(changes))
	changeText := make([]ChangeTextRange, 0, len(changes))
	for _, change := range changes {
		rng := *change.Range
		newLines := SplitLines(change.Text)
		if change.Text != "" {
			col := rng.Start.Character
			var remainder string
			for i, newLine := range newLines {
				lineNo := rng.Start.Line + i
				line := text[lineNo]
				text[lineNo] = line[:CharIndexToByte(line, col)] + newLine
				if lineNo == rng.End.Line {
					remainder = line[CharIndexToByte(line, rng.End.Character):]
					break
				}
				col = 0
			}
			// we need to insert more lines into the file
			if len(newLines) >= rng.End.Line-rng.Start.Line {
				leftover := newLines[rng.End.Line-rng.Start.Line+1:]
				text = insertLines(text, leftover, rng.End.Line+1)
				if remainder != "" {
					lastLine := rng.Start.Line + len(newLines) - 1
					text[lastLine] = text[lastLine] + remainder
				}
			} else {
				// We need to delete lines from the file
				text = deleteRange(text, lsp.Range{
					Start: lsp.Position{
						Line:      rng.Start.Line + len(newLines),
						Character: 0,
					},
					End: rng.End,
				})
			}
		} else {
			text = deleteRange(text, *change.Range)
		}
		changeText = append(changeText, ChangeTextRange{
			StartLine: change.Range.Start.Line,
			EndLine:   change.Range.End.Line,
			Text:      text[change.Range.Start.Line : change.Range.Start.Line+len(newLines)],
		})
	}
	return text, changeText
}

func deleteRange(text []string, rng lsp.Range) []string {
	startLine := text[rng.Start.Line]
	endLine := text[rng.End.Line]
	text[rng.Start.Line] = startLine[:CharIndexToByte(startLine, rng.Start.Character)] +
		endLine[CharIndexToByte(endLine, rng.End.Character):]
	numCut := rng.End.Line - rng.Start.Line
	if numCut > 0 {
		i := rng.Start.Line + 1
		text = text[:i+copy(text[i:], text[i+numCut:])]
	}
	return text
}
