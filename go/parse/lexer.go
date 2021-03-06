package parse

// import (
// 	"fmt"
// 	"strings"
// 	"unicode/utf8"
// )

// type stateFn func(*lexer) stateFn

// // item represents a token returned from the scanner
// type item struct {
// 	typ itemType // Type such as itemNumber
// 	val string   // Value, such as "23.2"
// }

// // itemType identifies the type of lex items
// type itemType int

// const (
// 	itemError itemType = iota // error occurred;
// 	itemDot                   // the cursor, spelled '.'
// 	itemEOF                   // marks the end

// 	itemElse      // else keyword
// 	itemEnd       // end keyword
// 	itemField     // identifier, starting with '.'
// 	itemIf        // if keyword
// 	itemLeftMeta  // left meta-string
// 	itemNumber    // number
// 	itemPipe      // pipe symbol
// 	itemRange     // range keword
// 	itemRawString // raw quoted string (includes quotes)
// 	itemRightMeta // right meta-string
// 	itemString    // quoted string (includes quotes)
// 	itemText      // plain text
// )

// func (i item) String() string {
// 	switch i.typ {
// 	case itemEOF:
// 		return "EOF"
// 	case itemError:
// 		return i.val
// 	}
// 	if len(i.val) > 10 {
// 		return fmt.Sprintf("%.10q...", i.val)
// 	}
// 	return fmt.Sprintf("%q", i.val)
// }

// type lexer struct {
// 	name  string    // used only for error reports.
// 	input string    // the string being scanned.
// 	state stateFn   // current state
// 	start int       // start position of this item.
// 	pos   int       // current position in the input.
// 	width int       // width of last rune read.
// 	items chan item // channel of scanned items
// }

// func lex(name, input string) (*lexer, chan item) {
// 	l := &lexer{
// 		name:  name,
// 		input: input,
// 		items: make(chan item),
// 	}
// 	go l.run()
// 	return l, l.items
// }

// // Lex the input until the state is nil
// func (l *lexer) run() {
// 	for state := lexText; state != nil; {
// 		state = state(l)
// 	}
// 	close(l.items) // No more tokens will be delivered
// }

// func lexText(l *lexer) stateFn {
// 	for {
// 		if strings.HasPrefix(l.input[l.pos:], leftMeta) {
// 			if l.pos > l.start {
// 				l.emit(itemText)
// 			}
// 			return lexLeftMeta // Next state
// 		}
// 		if l.next() == eof {
// 			break
// 		}
// 	}
// 	// Correctly reached EOF.
// 	if l.pos > l.start {
// 		l.emit(itemText)
// 	}
// 	l.emit(itemEOF) // Useful to make EOF a token.
// 	return nil      // Stop the run loop.
// }

// func lexLeftMeta(l *lexer) stateFn {
// 	l.pos += len(leftMeta)
// 	l.emit(itemLeftMeta)
// 	return lexInsideAction // Now inside {{ }}.
// }

// func lexInsideAction(l *lexer) stateFn {
// 	// Either number, quoted string, or identifier.
// 	// Spaces separate and are ignore.
// 	// Pipe symbols separate and are emitted.
// 	for {
// 		if strings.HasPrefix(l.input[l.pos:], rightMeta) {
// 			return lexRightMeta
// 		}
// 		switch r := l.next(); {
// 		case r == eof || r == '\n':
// 			return l.errorf("unclosed action")
// 		case isSpace(r):
// 			l.ignore()
// 		case r == '|':
// 			l.emit(itemPipe)
// 		case r == '"':
// 			return lexQuote
// 		case r == '`':
// 			return lexRawQuote
// 		case r == '+' || r == '-' || '0' <= r && r <= '9':
// 			l.backup()
// 			return lexNumber
// 		case isAlphaNumeric(r):
// 			l.backup()
// 			return lexIdentifier
// 		}
// 	}
// }

// func (l *lexer) next() rune {
// 	if l.pos >= len(l.input) {
// 		l.width = 0
// 		return eof
// 	}

// 	var r rune
// 	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
// 	l.pos += l.width
// 	return r
// }

// func (l *lexer) ignore() {
// 	l.start = l.pos
// }

// func (l *lexer) backup() {
// 	l.pos -= l.width
// }

// func (l *lexer) peek() int {
// 	r := l.next()
// 	l.backup()
// 	return r
// }

// // accept consumes the next rune
// // if it's from the valid set
// func (l *lexer) accept(valid string) bool {
// 	if strings.IndexRune(valid, l.next()) >= 0 {
// 		return true
// 	}
// 	l.backup()
// 	return false
// }

// // acceptRun consumes a run of runes from the valid set
// func (l *lexer) acceptRun(valid string) {
// 	for strings.IndexRune(valid, l.next()) >= 0 {
// 	}
// 	l.backup()
// }

// func lexNumber(l *lexer) stateFn {
// 	// Optional leading sign.
// 	l.accept("+-")
// 	// Is it a hex?
// 	digits := "0123456789"
// 	if l.accept("0") && l.accept("xX") {
// 		digits = "0123456789abcdefABCDEF"
// 	}
// 	l.acceptRun(digits)
// 	if l.accept(".") {
// 		l.acceptRun(digits)
// 	}
// 	if l.accept("eE") {
// 		l.accept("+-")
// 		l.acceptRun("0123456789")
// 	}
// 	// Is it imaginary?
// 	l.accept("i")
// 	// Next thing mustn't be alphanumeric
// 	if isAlphaNumeric(l.peek()) {
// 		l.next()
// 		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
// 	}
// 	l.emit(itemNumber)
// 	return lexInsideAction
// }

// const leftMeta = "{{"
// const rightMeta = "}}"

// // emit passes an item back to the client.
// func (l *lexer) emit(t itemType) {
// 	l.items <- item{t, l.input[l.start:l.pos]}
// 	l.start = l.pos
// }

// func (l *lexer) errorf(format string, args ...interface{}) stateFn {
// 	l.items <- item{
// 		itemError,
// 		fmt.Sprintf(format, args...),
// 	}
// 	return nil
// }

// func (l *lexer) nexrtItem() item {
// 	for {
// 		select {
// 		case item := <-l.items:
// 		default:
// 			l.state = l.state(l)
// 		}
// 	}
// }
