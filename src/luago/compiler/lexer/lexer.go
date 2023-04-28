package lexer

import "bytes"
import "fmt"
import "regexp"
import "strconv"
import "strings"

var reNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")				//正则表达式，处理换行符序列
var reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)			//正则表达式，用于探测左长方括号
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)		//正则表达式，短字符串
var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)	//正则表达式，数字字面量
var reIdentifier = regexp.MustCompile(`^[_\d\w]+`)					//正则表达式，标识符/关键字

var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

//词法规则概念：字符可以任意组合，词法规则定义了怎么样的组合可以构成合法的token。
//词法分析阶段：根据词法规则将字符序列分解为token序列。

//词法分析器
type Lexer struct {
	chunk         string // source code
	chunkName     string // source name
	line          int    // current line number
	//下面三个字段缓存下一个token信息
	nextToken     string
	nextTokenKind int
	nextTokenLine int
}

func NewLexer(chunk, chunkName string) *Lexer {
	return &Lexer{chunk, chunkName, 1, "", 0, 0}
}

//返回当前行号
func (self *Lexer) Line() int {
	return self.line
}

//查看下一个token，并且做缓存
func (self *Lexer) LookAhead() int {
	if self.nextTokenLine > 0 {
		return self.nextTokenKind
	}
	currentLine := self.line
	line, kind, token := self.NextToken()
	self.line = currentLine
	self.nextTokenLine = line
	self.nextTokenKind = kind
	self.nextToken = token
	return kind
}

//提取指定类型的token
func (self *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, _kind, token := self.NextToken()
	if kind != _kind {
		self.error("syntax error near '%s'", token)
	}
	return line, token
}

//提取标识符
func (self *Lexer) NextIdentifier() (line int, token string) {
	return self.NextTokenOfKind(TOKEN_IDENTIFIER)
}

//提取下一个token
func (self *Lexer) NextToken() (line, kind int, token string) {
	if self.nextTokenLine > 0 {
		//如果有下一个token的缓存，直接取缓存的值即可
		line = self.nextTokenLine
		kind = self.nextTokenKind
		token = self.nextToken
		self.line = self.nextTokenLine
		self.nextTokenLine = 0
		return
	}

	//跳过空白字符、换行和注释，以及更新行号
	self.skipWhiteSpaces()
	if len(self.chunk) == 0 {
		return self.line, TOKEN_EOF, "EOF"
	}

	switch self.chunk[0] {
	case ';':
		self.next(1)
		return self.line, TOKEN_SEP_SEMI, ";"
	case ',':
		self.next(1)
		return self.line, TOKEN_SEP_COMMA, ","
	case '(':
		self.next(1)
		return self.line, TOKEN_SEP_LPAREN, "("
	case ')':
		self.next(1)
		return self.line, TOKEN_SEP_RPAREN, ")"
	case ']':
		self.next(1)
		return self.line, TOKEN_SEP_RBRACK, "]"
	case '{':
		self.next(1)
		return self.line, TOKEN_SEP_LCURLY, "{"
	case '}':
		self.next(1)
		return self.line, TOKEN_SEP_RCURLY, "}"
	case '+':
		self.next(1)
		return self.line, TOKEN_OP_ADD, "+"
	case '-':
		self.next(1)
		return self.line, TOKEN_OP_MINUS, "-"
	case '*':
		self.next(1)
		return self.line, TOKEN_OP_MUL, "*"
	case '^':
		self.next(1)
		return self.line, TOKEN_OP_POW, "^"
	case '%':
		self.next(1)
		return self.line, TOKEN_OP_MOD, "%"
	case '&':
		self.next(1)
		return self.line, TOKEN_OP_BAND, "&"
	case '|':
		self.next(1)
		return self.line, TOKEN_OP_BOR, "|"
	case '#':
		self.next(1)
		return self.line, TOKEN_OP_LEN, "#"
	case ':':
		if self.test("::") {
			self.next(2)
			return self.line, TOKEN_SEP_LABEL, "::"
		} else {
			self.next(1)
			return self.line, TOKEN_SEP_COLON, ":"
		}
	case '/':
		if self.test("//") {
			self.next(2)
			return self.line, TOKEN_OP_IDIV, "//"
		} else {
			self.next(1)
			return self.line, TOKEN_OP_DIV, "/"
		}
	case '~':
		if self.test("~=") {
			self.next(2)
			return self.line, TOKEN_OP_NE, "~="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_WAVE, "~"
		}
	case '=':
		if self.test("==") {
			self.next(2)
			return self.line, TOKEN_OP_EQ, "=="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_ASSIGN, "="
		}
	case '<':
		if self.test("<<") {
			self.next(2)
			return self.line, TOKEN_OP_SHL, "<<"
		} else if self.test("<=") {
			self.next(2)
			return self.line, TOKEN_OP_LE, "<="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_LT, "<"
		}
	case '>':
		if self.test(">>") {
			self.next(2)
			return self.line, TOKEN_OP_SHR, ">>"
		} else if self.test(">=") {
			self.next(2)
			return self.line, TOKEN_OP_GE, ">="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_GT, ">"
		}
	case '.':
		if self.test("...") {
			self.next(3)
			return self.line, TOKEN_VARARG, "..."
		} else if self.test("..") {
			self.next(2)
			return self.line, TOKEN_OP_CONCAT, ".."
		} else if len(self.chunk) == 1 || !isDigit(self.chunk[1]) {
			self.next(1)
			return self.line, TOKEN_SEP_DOT, "."
		}
	case '[':
		if self.test("[[") || self.test("[=") {
			return self.line, TOKEN_STRING, self.scanLongString()
		} else {
			self.next(1)
			return self.line, TOKEN_SEP_LBRACK, "["
		}
	case '\'', '"':
		return self.line, TOKEN_STRING, self.scanShortString()
	}

	c := self.chunk[0]
	//数字字面量
	if c == '.' || isDigit(c) {
		token := self.scanNumber()
		return self.line, TOKEN_NUMBER, token
	}
	//标识符和关键字
	if c == '_' || isLetter(c) {
		token := self.scanIdentifier()
		if kind, found := keywords[token]; found {
			return self.line, kind, token // keyword 关键字
		} else { 
			return self.line, TOKEN_IDENTIFIER, token	//普通标识符
		}
	}

	self.error("unexpected symbol near %q", c)
	return
}

//跳过空白字符、换行和注释，以及更新行号
func (self *Lexer) skipWhiteSpaces() {
	for len(self.chunk) > 0 {
		if self.test("--") {
			self.skipComment()
		} else if self.test("\r\n") || self.test("\n\r") {
			self.next(2)
			self.line += 1
		} else if isNewLine(self.chunk[0]) {
			self.next(1)
			self.line += 1
		} else if isWhiteSpace(self.chunk[0]) {
			self.next(1)
		} else {
			break
		}
	}
}

//判断剩余的源代码是否以某种字符串开头
func (self *Lexer) test(s string) bool {
	return strings.HasPrefix(self.chunk, s)
}

//跳过n个字符
func (self *Lexer) next(n int) {
	self.chunk = self.chunk[n:]
}

func (self *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	err = fmt.Sprintf("%s:%d: %s", self.chunkName, self.line, err)
	panic(err)
}

//判断字符是否空白字符
func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

//判断字符是否回车或者换行符
func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

//跳过注释
func (self *Lexer) skipComment() {
	self.next(2) // skip --

	// long comment ?
	// 提取一个长字符扔掉
	if self.test("[") {
		if reOpeningLongBracket.FindString(self.chunk) != "" {
			self.scanLongString()
			return
		}
	}

	// short comment 
	// 跳过后续所有字符，直到遇到换行字符为止
	for len(self.chunk) > 0 && !isNewLine(self.chunk[0]) {	
		self.next(1)
	}
}

//长字符串字面量
func (self *Lexer) scanLongString() string {
	//先寻找左右长方括号，如果任何一个找不到，则说明源代码有语法错误
	openingLongBracket := reOpeningLongBracket.FindString(self.chunk)
	if openingLongBracket == "" {
		self.error("invalid long string delimiter near '%s'",
			self.chunk[0:2])
	}

	//提取字符串字面量，把左右长方括号去掉
	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIdx := strings.Index(self.chunk, closingLongBracket)
	if closingLongBracketIdx < 0 {
		self.error("unfinished long string or comment")
	}

	str := self.chunk[len(openingLongBracket):closingLongBracketIdx]
	self.next(closingLongBracketIdx + len(closingLongBracket))

	//把换行符序列统一转换成换行符\n，再把开头第一个换行符（如果有的话）去掉，就是最终的字符串
	str = reNewLine.ReplaceAllString(str, "\n")
	self.line += strings.Count(str, "\n")
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}

	return str
}

//短字符串字面量
func (self *Lexer) scanShortString() string {
	//使用正则表达式提取短字符串
	if str := reShortStr.FindString(self.chunk); str != "" {
		self.next(len(str))
		//去掉字面量两端的引号
		str = str[1 : len(str)-1]
		if strings.Index(str, `\`) >= 0 {
			self.line += len(reNewLine.FindAllString(str, -1))
			//escape()方法对转义序列进行处理，得到最终字符串
			str = self.escape(str)
		}
		return str
	}
	self.error("unfinished string")
	return ""
}

//提取数字字面量
func (self *Lexer) scanNumber() string {
	return self.scan(reNumber)
}

//提取标识符/关键字
func (self *Lexer) scanIdentifier() string {
	return self.scan(reIdentifier)
}

func (self *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(self.chunk); token != "" {
		self.next(len(token))
		return token
	}
	panic("unreachable!")
}

//对转义序列进行处理
func (self *Lexer) escape(str string) string {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			self.error("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				self.error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				self.error("UTF-8 value too large near '%s'", found)
			}
		case 'z':
			//跳过\z这个转义序列，然后跳过紧随其后的空白字符
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { // todo
				str = str[1:]
			}
			continue
		}
		self.error("invalid escape sequence near '\\%c'", str[1])
	}

	return buf.String()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
