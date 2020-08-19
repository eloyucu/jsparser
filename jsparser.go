package jsparser

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf16"
)

type JsonParser struct {
	reader        *bufio.Reader
	loopProp      []byte
	resChan       chan *JSON
	isResArr      bool
	skipProps     map[string]bool
	TotalReadSize uint64
	lastReadSize  int
	scratch       *scratch
}

// JSON parsed result
type JSON struct {
	StringVal  string
	BoolVal    bool
	ArrayVals  []interface{}
	ObjectVals map[string]interface{}
	ValueType  ValueType
	Err        error
}

// ValueType of JSON value
type ValueType int8

// JSON types
const (
	Invalid ValueType = iota
	Null
	String
	Number
	Boolean
	Array
	Object
)

func NewJSONParser(reader *bufio.Reader, loopProp string) *JsonParser {

	j := &JsonParser{
		reader:    reader,
		loopProp:  []byte(loopProp),
		resChan:   make(chan *JSON, 256),
		skipProps: map[string]bool{},
		scratch:   &scratch{data: make([]byte, 2048), dataRes: make([]*JSON, 2048)},
	}
	return j
}

func (j *JsonParser) SkipProps(skipProps []string) *JsonParser {

	if len(skipProps) > 0 {
		for _, s := range skipProps {
			j.skipProps[s] = true
		}
	}
	return j

}

func (j *JsonParser) Stream() chan *JSON {

	go j.parse()

	return j.resChan

}

func (j *JsonParser) Parse() []*JSON {

	j.isResArr = true
	j.parse()
	return j.scratch.allRes()

}
func (element *JSON) GetAllNodes(xpath string) map[string]*JSON {
	var path, paths string
	xpaths := strings.SplitN(xpath, ".", 2)
	if len(xpaths) > 1 {
		paths = xpaths[1]
	}
	path = xpaths[0]
	if elementAux, ok := element.ObjectVals[path]; ok {
		if element, ok = elementAux.(*JSON); !ok {
			return map[string]*JSON{}
		}
	} else {
		return map[string]*JSON{}
	}

	if len(xpaths) > 1 {
		return element.GetAllNodes(paths)
	}
	return element.GetObjectVals()
}
func (element *JSON) GetNodes(xpath string) []*JSON {
	var path, paths string
	xpaths := strings.SplitN(xpath, ".", 2)
	if len(xpaths) > 1 {
		paths = xpaths[1]
	}
	path = xpaths[0]
	path, index := element.pathIndex(path)

	elementAux := element.ObjectVals[path]
	if e, ok := elementAux.(*JSON); ok {
		if paths == "" {
			if len(e.ArrayVals) != 0 {
				return e.GetArrayVals(index)
			}
			return []*JSON{e}
		} else {
			for i, eArr := range e.ArrayVals {
				if i == int(index) {
					if eAux, ok := eArr.(*JSON); ok {
						return eAux.GetNodes(paths)
					}
				}
			}
		}
		return e.GetNodes(paths)
	} else if elements, ok := elementAux.([]*JSON); ok {
		if paths == "" && index == 0 {
			return elements
		} else if paths == "" {
			for i, e := range elements {
				if i == int(index) {
					return []*JSON{e}
				}
			}
		}
		return []*JSON{}
	} else if path != "" {
		for _, e := range element.ArrayVals {
			if element, ok = e.(*JSON); ok {
				elementAux = element.ObjectVals[path]
				if eAux, ok := elementAux.(*JSON); ok {
					if paths == "" {
						return []*JSON{eAux}
					}
					return element.GetNodes(paths)
				} else if elements, ok := elementAux.([]*JSON); ok {
					if paths == "" && index == 0 {
						return elements
					} else if paths == "" {
						for i, e := range elements {
							if i == int(index) {
								return []*JSON{e}
							}
						}
					}
				} else if paths == "" {
					return []*JSON{
						{
							StringVal: stringify(elementAux),
							BoolVal:   false,
							ValueType: element.ValueType,
						},
					}
				}
			}
		}
	}
	return []*JSON{
		{
			StringVal: stringify(elementAux),
			BoolVal:   false,
			ValueType: 6,
		},
	}
}
func (element *JSON) GetNode(xpath string) *JSON {
	nodes := element.GetNodes(xpath)
	if len(nodes) > 0 && nodes[0].ValueType != Null {
		return nodes[0]
	}
	return &JSON{}
}
func (element *JSON) GetValueF64(xpath string) float64 {
	v := element.GetValue(xpath)
	f := 0.00
	if t, err := strconv.ParseFloat(v, 64); err == nil {
		f = t
	}
	return f
}
func (element *JSON) GetValueInt(xpath string) int {
	i := element.GetValueF64(xpath)
	return int(i)
}
func (element *JSON) GetValue(xpath string) string {
	if xpath == "." {
		return element.StringVal
	} else if xpath == "" {
		return ""
	}
	node := element.GetNode(xpath)
	if node != nil {
		return node.StringVal
	}
	return ""
}
func (element *JSON) GetObjectVals() map[string]*JSON {
	nodes := map[string]*JSON{}
	for key, value := range element.ObjectVals {
		if node, ok := value.(*JSON); ok {
			nodes[key] = node
		} else {
			nodes[key] = &JSON{
				StringVal: stringify(value),
				BoolVal:   false,
				ValueType: 6,
			}
		}
	}
	return nodes
}
func (element *JSON) GetArrayVals(index int64) []*JSON {
	nodes := []*JSON{}
	for i, a := range element.ArrayVals {
		if index == math.MaxInt64 || int64(i) == index {
			if node, ok := a.(*JSON); ok {
				nodes = append(nodes, node)
			} else {
				nodes = append(nodes, &JSON{
					StringVal: stringify(a),
					BoolVal:   false,
					ValueType: 6,
				})
			}
		}
	}
	return nodes
}
func (element *JSON) IsEmpty() bool {
	if element == nil || (len(element.ArrayVals) == 0 && len(element.ObjectVals) == 0 && element.StringVal == "") {
		return true
	}
	return false
}
func stringify(i interface{}) string {
	if s, ok := i.(string); ok {
		return s
	} else if b, ok := i.(bool); ok {
		return strconv.FormatBool(b)
	} else if n, ok := i.(float64); ok {
		return strconv.FormatFloat(n, 'f', 6, 64)
	} else if n, ok := i.(int); ok {
		return strconv.Itoa(n)
	}
	return ""
}
func (element *JSON) pathIndex(path string) (string, int64) {
	indexes := strings.Split(path, "[")
	path = indexes[0]
	if len(indexes) > 1 {
		indexes := strings.Split(indexes[1], "]")
		index, err := strconv.ParseInt(indexes[0], 10, 64)
		if err != nil {
			return path, math.MaxInt64
		}
		return path, index
	}
	return path, math.MaxInt64
}
func (j *JsonParser) parse() {

	defer close(j.resChan)

	var b byte
	var err error

	for {
		b, err = j.readByte()

		if err != nil {
			return
		}

		if j.isWS(b) {
			continue
		}

		if b == '"' { // begining of possible json property

			isprop, err := j.getPropName()

			if err != nil {
				j.sendError()
				return
			}

			if isprop {

				b, err = j.skipWS()
				if err != nil {
					j.sendError()
					return
				}

				valType, typeErr := j.getValueType(b)

				if typeErr != nil {
					j.sendError()
					return
				}

				if bytes.Equal(j.loopProp, j.scratch.bytes()) {

					switch valType {
					case String:

						err = j.string()

						if err != nil {
							j.sendError()
							return
						}
						j.sendRes(&JSON{StringVal: j.scratch.string(), ValueType: String})

					case Array:

						success := j.loopArray()
						if !success {
							return
						}

					case Object:

						res := &JSON{ObjectVals: map[string]interface{}{}, ValueType: Object}
						j.getObjectTree(res)
						j.sendRes(res)
						if res.Err != nil {
							return
						}

					case Boolean:

						b, err := j.boolean()
						if err != nil {
							j.sendError()
							return
						}
						j.sendRes(&JSON{BoolVal: b, ValueType: Boolean})

					case Number:

						err = j.number(b)

						if err != nil {
							j.sendError()
							return
						}
						j.sendRes(&JSON{StringVal: j.scratch.string(), ValueType: Number})

					case Null:

						err := j.null()

						if err != nil {
							j.sendError()
							return
						}
						j.sendRes(&JSON{ValueType: Null})

					}

				} else {

					if valType == String { // if valtype is string just skip it otherwise continue looking loopProp.
						err = j.skipString()
						if err != nil {
							j.sendError()
							return
						}
					}

				}
			}
		}
	}

}

func (j *JsonParser) sendRes(res *JSON) {
	if j.isResArr {
		j.scratch.addRes(res)
	} else {
		j.resChan <- res
	}
}

func (j *JsonParser) loopArray() bool {

	var b byte
	var err error

	for {

		b, err = j.skipWS()

		if err != nil {
			j.sendError()
			return false
		}

		if b == ']' {
			return true
		}

		if b == ',' {
			continue
		}

		valType, err := j.getValueType(b)

		if err != nil {
			j.sendError()
			return false
		}

		switch valType {
		case String:

			err = j.string()

			if err != nil {
				j.sendRes(&JSON{Err: err, ValueType: Invalid})
				return false
			}
			j.sendRes(&JSON{StringVal: j.scratch.string(), ValueType: String})
		case Array:

			res := &JSON{ObjectVals: map[string]interface{}{}, ValueType: Array}
			j.getArrayTree(res)
			j.sendRes(res)

		case Object:

			res := &JSON{ObjectVals: map[string]interface{}{}, ValueType: Object}
			j.getObjectTree(res)
			j.sendRes(res)

		case Boolean:

			b, err := j.boolean()
			if err != nil {
				j.sendError()
				return false
			}
			j.sendRes(&JSON{BoolVal: b, ValueType: Boolean})

		case Number:

			err = j.number(b)
			if err != nil {
				return false
			}
			j.sendRes(&JSON{StringVal: j.scratch.string(), ValueType: Number})

		case Null:

			err := j.null()

			if err != nil {
				return false
			}
			j.sendRes(&JSON{ValueType: Null})

		}

	}

}

func (j *JsonParser) getObjectTree(res *JSON) {

	if res.Err != nil {
		return
	}

	var b byte
	var err error
	for {

		b, err = j.readByte()

		if err != nil {
			res.Err = err
			return
		}

		if j.isWS(b) {
			continue
		}

		if b == '"' { // begining of json property

			_, err := j.getPropName() // first variable ommited because inside object there can't be string item
			prop := j.scratch.string()

			if err != nil {
				res.Err = err
				return
			}

			b, err = j.skipWS()
			if err != nil {
				res.Err = j.defaultError()
				return
			}

			valType, err := j.getValueType(b)

			if err != nil {
				res.Err = err
				return
			}

			switch valType {
			case String:

				if ok := j.skipProps[prop]; ok {
					err = j.skipString()

					if err != nil {
						res.Err = err
						return
					}
					break
				}

				err = j.string()

				if err != nil {
					res.Err = err
					return
				}

				res.ObjectVals[prop] = j.scratch.string()

			case Array:

				if ok := j.skipProps[prop]; ok {
					err = j.skipArrayOrObject('[', ']')

					if err != nil {
						res.Err = err
						return
					}
					break
				}
				r := &JSON{ValueType: Array}
				j.getArrayTree(r)
				if r.Err != nil {
					res.Err = r.Err
					return
				}
				res.ObjectVals[prop] = r

			case Object:

				if ok := j.skipProps[prop]; ok {
					err = j.skipArrayOrObject('{', '}')

					if err != nil {
						res.Err = err
						return
					}
					break
				}
				r := &JSON{ObjectVals: map[string]interface{}{}, ValueType: Object}
				j.getObjectTree(r)

				if r.Err != nil {
					res.Err = r.Err
					return
				}
				res.ObjectVals[prop] = r

			case Boolean:

				b, err := j.boolean()

				if err != nil {
					res.Err = err
					return
				}

				// rest of the skip since they are small we just don't include in the result
				if ok := j.skipProps[prop]; !ok {
					res.ObjectVals[prop] = b
				}

			case Number:

				err = j.number(b)

				if err != nil {
					res.Err = err
					return
				}

				if ok := j.skipProps[prop]; !ok {
					res.ObjectVals[prop] = j.scratch.string()
				}

			case Null:

				err = j.null()
				if err != nil {
					res.Err = err
					return
				}

				if ok := j.skipProps[prop]; !ok {
					res.ObjectVals[prop] = ""
				}

			}

		} else if b == ',' {

			continue

		} else if b == '}' { // completion of current object

			return

		} else { // invalid end

			res.Err = j.defaultError()
			return

		}

	}

}

func (j *JsonParser) getArrayTree(res *JSON) {

	if res.Err != nil {
		return
	}

	var b byte
	var err error

	for {

		b, err = j.readByte()

		if err != nil {
			res.Err = err
			return
		}

		if j.isWS(b) {
			continue
		}

		if b == ',' {
			continue
		}

		if b == ']' { // means complete of current array
			return
		}

		valType, err := j.getValueType(b)

		if err != nil {
			res.Err = err
			return
		}
		switch valType {
		case String:

			err = j.string()

			if err != nil {
				res.Err = err
				return
			}
			res.ArrayVals = append(res.ArrayVals, j.scratch.string())

		case Array:

			r := &JSON{ValueType: Array}
			j.getArrayTree(r)
			if r.Err != nil {
				res.Err = r.Err
				return
			}
			res.ArrayVals = append(res.ArrayVals, r)

		case Object:

			r := &JSON{ObjectVals: map[string]interface{}{}, ValueType: Object}
			j.getObjectTree(r)
			if r.Err != nil {
				res.Err = r.Err
				return
			}
			res.ArrayVals = append(res.ArrayVals, r)

		case Boolean:

			b, err := j.boolean()
			if err != nil {
				res.Err = err
				return
			}

			res.ArrayVals = append(res.ArrayVals, b)

		case Number:

			err = j.number(b)
			if err != nil {
				res.Err = err
				return
			}
			res.ArrayVals = append(res.ArrayVals, j.scratch.string())

		case Null:

			err = j.null()

			if err != nil {
				res.Err = err
				return
			}

			res.ArrayVals = append(res.ArrayVals, "")

		}

	}

}

func (j *JsonParser) number(first byte) error {

	var c byte
	var err error
	j.scratch.reset()
	j.scratch.add(first)

	for {

		c, err = j.readByte()

		if err != nil {
			return j.defaultError()
		}

		if j.isWS(c) {

			c, err = j.skipWS()

			if err != nil {
				return j.defaultError()
			}

			if !(c == ',' || c == '}' || c == ']') {
				return j.defaultError()
			}
			err := j.unreadByte()
			if err != nil {
				return j.defaultError()
			}

			return nil
		}

		if c == ',' || c == '}' || c == ']' {

			err := j.unreadByte()
			if err != nil {
				return j.defaultError()
			}

			return nil
		}

		j.scratch.add(c)

	}

}

func (j *JsonParser) boolean() (bool, error) {

	var c byte
	var err error

	c, err = j.readByte()

	if err != nil {
		return false, j.defaultError()
	}

	// true
	if c == 'r' {
		c, err = j.readByte()

		if err != nil {
			return false, j.defaultError()
		}
		if c == 'u' {
			c, err = j.readByte()

			if err != nil {
				return false, j.defaultError()
			}
			if c == 'e' {
				// check last
				c, err = j.skipWS()
				if err != nil {
					return false, j.defaultError()
				}
				if !(c == ',' || c == '}' || c == ']') {
					return false, j.defaultError()
				}
				err := j.unreadByte()
				if err != nil {
					return false, j.defaultError()
				}

				return true, nil
			}
		}
	}

	// false
	if c == 'a' {
		c, err = j.readByte()

		if err != nil {
			return false, j.defaultError()
		}
		if c == 'l' {
			c, err = j.readByte()

			if err != nil {
				return false, j.defaultError()
			}
			if c == 's' {
				c, err = j.readByte()

				if err != nil {
					return false, j.defaultError()
				}
				if c == 'e' {
					// check last
					c, err = j.skipWS()
					if err != nil {
						return false, j.defaultError()
					}
					if !(c == ',' || c == '}' || c == ']') {
						return false, j.defaultError()
					}
					err := j.unreadByte()
					if err != nil {
						return false, j.defaultError()
					}

					return false, nil
				}
			}
		}
	}

	return false, j.defaultError()

}

func (j *JsonParser) null() error {

	var c byte
	var err error

	c, err = j.readByte()

	if err != nil {
		return j.defaultError()
	}

	// true
	if c == 'u' {
		c, err = j.readByte()

		if err != nil {
			return j.defaultError()
		}

		if c == 'l' {
			c, err = j.readByte()

			if err != nil {
				return j.defaultError()
			}
			if c == 'l' {
				// check last
				c, err = j.skipWS()
				if err != nil {
					return j.defaultError()
				}

				if !(c == ',' || c == '}' || c == ']') {
					return j.defaultError()
				}

				err := j.unreadByte()
				if err != nil {
					return j.defaultError()
				}

				return nil
			}
		}
	}

	return j.defaultError()
}

func (j *JsonParser) skipString() error {

	var c byte
	var prev byte
	var prevPrev byte
	var err error
	for {

		c, err = j.readByte()

		if err != nil {
			return j.defaultError()
		}

		if c == '"' {

			if !(prev == '\\' && prevPrev != '\\') { // escape check
				return nil
			}

		}

		prevPrev = prev
		prev = c

	}

}

func (j *JsonParser) skipArrayOrObject(start byte, end byte) error {

	var c byte
	var err error
	var depth = 1
	for {

		c, err = j.readByte()

		if err != nil {
			return j.defaultError()
		}

		switch c {
		case '"':
			err = j.skipString() // this is needed because string can contain [ or ]
			if err != nil {
				return err
			}
		case start:
			depth++
		case end:
			depth--
			if depth == 0 {
				return nil
			}

		}

	}

}

func (j *JsonParser) getValueType(c byte) (ValueType, error) {

	switch c {
	case '"':
		return String, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		return Number, nil
	case 'f':
		return Boolean, nil
	case 't':
		return Boolean, nil
	case 'n':
		return Null, nil
	case '[':
		return Array, nil
	case '{':
		return Object, nil
	}

	return Invalid, j.defaultError()

}

// first return type is checking if it is property or just an array item
func (j *JsonParser) getPropName() (bool, error) {

	err := j.string()

	if err != nil {
		return false, err
	}

	b, err := j.skipWS()

	if err != nil {
		return false, err
	}

	if b == ':' { // end of property name
		return true, nil
	}

	err = j.unreadByte()

	return false, err

}

func (j *JsonParser) isWS(in byte) bool {

	if in == ' ' || in == '\n' || in == '\t' || in == '\r' {
		return true
	}

	return false

}

// skips WS and read first non WS
func (j *JsonParser) skipWS() (byte, error) {

	var b byte
	var err error
	for {
		b, err = j.readByte()
		if err != nil {
			return 0, err
		}
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			continue
		} else {
			return b, nil
		}
	}

}

func (j *JsonParser) readByte() (byte, error) {

	by, err := j.reader.ReadByte()

	j.TotalReadSize = j.TotalReadSize + 1

	j.lastReadSize = 1

	if err != nil {
		return 0, err
	}
	return by, nil

}

func (j *JsonParser) unreadByte() error {

	err := j.reader.UnreadByte()
	if err != nil {
		return err
	}
	j.TotalReadSize = j.TotalReadSize - 1
	return nil

}

func (j *JsonParser) sendError() {
	err := fmt.Errorf("Invalid json")
	if j.isResArr {
		j.scratch.addRes(&JSON{Err: err, ValueType: Invalid})
	} else {
		j.resChan <- &JSON{Err: err, ValueType: Invalid}
	}
}

func (j *JsonParser) resultError() *JSON {

	return &JSON{Err: j.defaultError(), ValueType: Invalid}

}

func (j *JsonParser) defaultError() error {
	err := fmt.Errorf("Invalid json")
	return err
}

// based on https://github.com/bcicen/jstream
func (j *JsonParser) string() error {

	j.scratch.reset()

	var err error
	var c byte

	c, err = j.readByte()
	if err != nil {
		if err != nil {
			return j.defaultError()
		}
	}

scan:
	for {
		switch {
		case c == '"':
			return nil
		case c == '\\':
			c, err = j.readByte()
			if err != nil {
				if err != nil {
					return j.defaultError()
				}
			}
			goto scan_esc
		case c < 0x20:
			return j.defaultError()
			// Coerce to well-formed UTF-8.

		}
		j.scratch.add(c)
		c, err = j.readByte()
		if err != nil {
			if err != nil {
				return j.defaultError()
			}
		}
	}

scan_esc:
	switch c {
	case '"', '\\', '/', '\'':
		j.scratch.add(c)
	case 'u':
		goto scan_u
	case 'b':
		j.scratch.add('\b')
	case 'f':
		j.scratch.add('\f')
	case 'n':
		j.scratch.add('\n')
	case 'r':
		j.scratch.add('\r')
	case 't':
		j.scratch.add('\t')
	default:
		//err := fmt.Errorf("error in string escape code")
		return j.defaultError()
	}

	c, err = j.readByte()
	if err != nil {
		if err != nil {
			return j.defaultError()
		}
	}

	goto scan

scan_u:
	r := j.u4()
	if r < 0 {
		//err := fmt.Errorf("in unicode escape sequence")
		return j.defaultError()
	}

	// check for proceeding surrogate pair
	c, err = j.readByte()
	if err != nil {
		if err != nil {
			return j.defaultError()
		}
	}

	if !utf16.IsSurrogate(r) || c != '\\' {
		j.scratch.addRune(r)
		goto scan
	}

	c, err = j.readByte()
	if err != nil {
		if err != nil {
			return j.defaultError()
		}
	}

	if c != 'u' {
		j.scratch.addRune(r)
		goto scan_esc
	}

	r2 := j.u4()
	if r2 < 0 {
		return j.defaultError()
	}

	// write surrogate pair
	j.scratch.addRune(utf16.DecodeRune(r, r2))

	c, err = j.readByte()
	if err != nil {
		if err != nil {
			return j.defaultError()
		}
	}

	goto scan
}

// u4 reads four bytes following a \u escape
func (j *JsonParser) u4() rune {
	// logic taken from:
	// github.com/buger/jsonparser/blob/master/escape.go#L20

	var c byte
	var err error
	var h [4]int
	for i := 0; i < 4; i++ {

		c, err = j.readByte()
		if err != nil {
			if err != nil {
				return -1
			}
		}
		switch {
		case c >= '0' && c <= '9':
			h[i] = int(c - '0')
		case c >= 'A' && c <= 'F':
			h[i] = int(c - 'A' + 10)
		case c >= 'a' && c <= 'f':
			h[i] = int(c - 'a' + 10)
		default:
			return -1
		}
	}
	return rune(h[0]<<12 + h[1]<<8 + h[2]<<4 + h[3])
}
