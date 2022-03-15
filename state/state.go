package state

import (
	"log"
	"reflect"
	"sync"

	"github.com/asaskevich/EventBus"
	"github.com/sourcegraph/go-lsp"
)

type WorkspaceState struct {
	files  map[string]*File
	view   *View
	mu     sync.Mutex
	events EventBus.Bus
	logger *log.Logger
	nextID int32
}

type File struct {
	Filename string   `json:"filename"`
	ID       int32    `json:"id"`
	Lines    []string `json:"lines,omitempty"`
	Language string   `json:"language"`
}

type View struct {
	FileID    int32      `json:"file_id"`
	Line      int        `json:"line"`
	Character int        `json:"character"`
	Range     *lsp.Range `json:"range"`
}

type OpenFileEvent struct {
	Filename string `json:"filename"`
	ID       int32  `json:"id"`
	Language string `json:"language"`
}

type CloseFileEvent struct {
	FileID int32 `json:"file_id"`
}

type ReplaceTextEvent struct {
	FileID int32    `json:"file_id"`
	Text   []string `json:"text"`
}

type ChangeTextRange struct {
	StartLine int      `json:"start_line"`
	EndLine   int      `json:"end_line"`
	Text      []string `json:"text"`
}

type UpdateTextEvent struct {
	FileID  int32             `json:"file_id"`
	Changes []ChangeTextRange `json:"changes"`
}

type ChangeViewEvent struct {
	View View `json:"view"`
}

const TOPIC = "all"

func NewState(logger *log.Logger) *WorkspaceState {
	return &WorkspaceState{
		files:  make(map[string]*File),
		events: EventBus.New(),
		logger: logger,
	}
}

func (s *WorkspaceState) publish(value interface{}) {
	s.events.Publish(TOPIC, value)
}

func (s *WorkspaceState) Subscribe(fn interface{}) error {
	return s.events.Subscribe(TOPIC, fn)
}

func (s *WorkspaceState) Unsubscribe(fn interface{}) error {
	return s.events.Unsubscribe(TOPIC, fn)
}

func (s *WorkspaceState) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.files {
		delete(s.files, k)
	}
	s.view = nil
}

func (s *WorkspaceState) CursorMove(filename string, position lsp.Position, rng *lsp.Range) {
	s.mu.Lock()
	defer s.mu.Unlock()
	anyChanges := false
	file := s.files[filename]
	if s.view.FileID != file.ID {
		s.view.FileID = file.ID
		anyChanges = true
	}
	if s.view.Line != position.Line {
		s.view.Line = position.Line
		anyChanges = true
	}
	if s.view.Character != position.Character {
		s.view.Character = CharIndexToRune(file.Lines[position.Line], position.Character)
		anyChanges = true
	}
	if rng != nil {
		if !reflect.DeepEqual(s.view.Range, rng) {
			anyChanges = true
			startLine := file.Lines[rng.Start.Line]
			endLine := file.Lines[rng.End.Line]
			s.view.Range = &lsp.Range{
				Start: lsp.Position{
					Line:      rng.Start.Line,
					Character: CharIndexToRune(startLine, rng.Start.Character),
				},
				End: lsp.Position{
					Line:      rng.End.Line,
					Character: CharIndexToRune(endLine, rng.End.Character),
				},
			}
		}
	} else if s.view.Range != nil {
		s.view.Range = nil
		anyChanges = true
	}
	if anyChanges {
		s.publish(ChangeViewEvent{
			View: *s.view,
		})
	}
}

func (s *WorkspaceState) OpenFile(filename string, text string, language string, updateCursor bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[filename] = &File{
		Filename: filename,
		ID:       s.nextID,
		Lines:    SplitLines(text),
		Language: language,
	}
	s.publish(OpenFileEvent{
		Filename: filename,
		ID:       s.nextID,
		Language: language,
	})

	if updateCursor || s.view == nil {
		s.view = &View{
			FileID:    s.nextID,
			Line:      0,
			Character: 0,
		}
		s.publish(ChangeViewEvent{
			View: *s.view,
		})
	}
	s.nextID++
}

func (s *WorkspaceState) CloseFile(filename string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file := s.files[filename]
	delete(s.files, filename)
	s.publish(CloseFileEvent{
		FileID: file.ID,
	})
}

func (s *WorkspaceState) ReplaceTextRanges(filename string, changes []lsp.TextDocumentContentChangeEvent, updateCursor bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file := s.files[filename]
	var changeText []ChangeTextRange
	newLines, changeText := applyTextChanges(file.Lines, changes)
	s.publish(UpdateTextEvent{
		FileID:  file.ID,
		Changes: changeText,
	})

	if updateCursor {
		lastChange := changeText[len(changeText)-1]
		col := 0
		changeLine := lastChange.Text[len(lastChange.Text)-1]
		if lastChange.EndLine < len(file.Lines) {
			col = longestCommonPrefix(file.Lines[lastChange.EndLine], changeLine)
		} else {
			col = len(changeLine)
		}
		s.view = &View{
			FileID:    file.ID,
			Line:      lastChange.EndLine + len(lastChange.Text) - 1,
			Character: col,
		}
		s.publish(ChangeViewEvent{
			View: *s.view,
		})
	}

	file.Lines = newLines
}

func (s *WorkspaceState) ReplaceText(filename string, text string, updateCursor bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newLines := SplitLines(text)
	prev := s.files[filename]
	s.files[filename] = &File{
		Filename: prev.Filename,
		ID:       prev.ID,
		Language: prev.Language,
		Lines:    newLines,
	}
	s.publish(ReplaceTextEvent{
		FileID: prev.ID,
		Text:   newLines,
	})

	if updateCursor {
		lnum := 0
		col := 0
		lnum = -1
		for i, line := range prev.Lines {
			if i >= len(newLines) {
				lnum = i
				break
			} else if line != newLines[i] {
				lnum = i
				col = longestCommonPrefix(line, newLines[i])
				break
			}
		}
		if lnum == -1 {
			lnum = len(newLines)
		}
		s.view = &View{
			FileID:    prev.ID,
			Line:      lnum,
			Character: col,
		}
		s.publish(ChangeViewEvent{
			View: *s.view,
		})
	}
}

func (s *WorkspaceState) GetFiles() []File {
	s.mu.Lock()
	defer s.mu.Unlock()

	ret := make([]File, 0, len(s.files))
	for _, value := range s.files {
		ret = append(ret, File{
			Filename: value.Filename,
			ID:       value.ID,
			Language: value.Language,
		})
	}
	return ret
}

func (s *WorkspaceState) GetFile(filename string) File {
	s.mu.Lock()
	defer s.mu.Unlock()
	f := s.files[filename]
	lines := make([]string, len(f.Lines))
	copy(lines, f.Lines)
	file := File{
		ID:       f.ID,
		Filename: f.Filename,
		Language: f.Language,
		Lines:    lines,
	}
	return file
}

func (s *WorkspaceState) GetView() *View {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.view
}

func longestCommonPrefix(s1 string, s2 string) int {
	for i := 0; i < len(s1) && i < len(s2); i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	if len(s1) < len(s2) {
		return len(s1)
	} else {
		return len(s2)
	}
}
