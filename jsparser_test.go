package jsparser

import (
	"bufio"
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var minify bool
var parseall bool

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	flag.BoolVar(&minify, "minify", false, "Minify")

	flag.BoolVar(&parseall, "parseall", false, "ParseAll")

	flag.Parse()

	os.Exit(m.Run())
}

func getparser(prop string) *JsonParser {

	if minify {
		// todo add some space after some values
		const minijson string = `{"nu":null,"b":true,"b1":false,"n":2323,"n1":23.23,"n2":23.23e-6 ,"s":"sstring","s1":"s1tring","s2":"s2tr\\ing\"蒜","o":{"o1":"o1string","o2":"o2string","o3":true,"o4":["o4string",{"o41":"o41string"},["o4nestedarray item 1","o4nestedarray item 1 item 2",true,99,null,90.98]],"o5":98.21,"o6":null,"o7":{"o71":"o71string","o72":["o72string",null,false,98,{}],"o73":true,"o74":98}},"a":[{"a11":"o71string\\","a12":["o72string",null,false,98,{}],"a13":true,"a14":98},{"a11":"o71string","a12":["o72string",null,false,98,{}],"a13":true,"a14":98},"astringinside",false,99,null,0.00043333]}`

		br := bufio.NewReader(strings.NewReader(minijson))

		p := NewJSONParser(br, prop)

		return p
	}

	file, _ := os.Open("sample.json")

	br := bufio.NewReader(file)

	p := NewJSONParser(br, prop)

	return p

}

func allResult(p *JsonParser) []*JSON {

	if parseall {
		return p.Parse()

	}
	var res []*JSON
	for json := range p.Stream() {
		res = append(res, json)
	}
	return res

}
func TestString(t *testing.T) {

	var js JSON

	p := getparser("s")
	resultCount := 0

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		js = *json
		resultCount++

	}

	if resultCount != 1 {
		panic("result count must 1")
	}

	if js.StringVal != "sstring" {
		panic("invalid result string")
	}

	if js.ValueType != String {
		panic("Value type must be string")
	}

	p = getparser("s2")

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		js = *json

	}

	if js.StringVal != "s2tr\\ing\"蒜" {
		panic("invalid result string")
	}

	// Skip

}

func TestBoolean(t *testing.T) {

	p := getparser("b")

	resultCount := 0
	var js JSON

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		js = *json
		resultCount++

	}

	if resultCount != 1 {
		panic("result count must 1")
	}

	if !js.BoolVal {
		panic("invalid result boolean")
	}

	if js.ValueType != Boolean {
		panic("Value type must be boolean")
	}

}

func TestNumber(t *testing.T) {

	p := getparser("n2")

	resultCount := 0
	var js JSON

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		js = *json
		resultCount++

	}

	if resultCount != 1 {
		panic("result count must 1")
	}

	if js.StringVal != "23.23e-6" {
		panic("invalid result")
	}

	if js.ValueType != Number {
		panic("Value type must be boolean")
	}

}

func TestNull(t *testing.T) {

	p := getparser("nu")

	resultCount := 0
	var js JSON

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		js = *json
		resultCount++

	}

	if resultCount != 1 {
		panic("result count must 1")
	}

	if js.StringVal != "" {
		panic("invalid result")
	}

	if js.ValueType != Null {
		panic("Value type must be null")
	}

}

func TestObject(t *testing.T) {

	p := getparser("o")

	resultCount := 0
	var js JSON

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		js = *json
		resultCount++

	}

	if resultCount != 1 {
		panic("result count must 1")
	}

	if js.ValueType != Object {
		panic("Value type must be object")
	}

	if val, ok := js.ObjectVals["o1"]; !ok || val.(string) != "o1string" {
		panic("Test failed")
	}

	if val, ok := js.ObjectVals["o2"]; !ok || val.(string) != "o2string" {
		panic("Test failed")
	}

	if val, ok := js.ObjectVals["o3"]; !ok || !val.(bool) {
		panic("Test failed")
	}

	if val, ok := js.ObjectVals["o4"]; !ok || len(val.(*JSON).ArrayVals) != 3 {
		panic("Test failed")
	}

	if val, ok := js.ObjectVals["o4"]; !ok || len(val.(*JSON).ArrayVals[2].(*JSON).ArrayVals) != 6 {
		panic("Test failed")
	}

	// Skip test
	p = getparser("o").SkipProps([]string{"o1", "o2", "o4", "o5", "o6", "o7"})

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		js = *json
		resultCount++
	}

	if _, ok := js.ObjectVals["o1"]; ok {
		panic("Test failed")
	}

	if _, ok := js.ObjectVals["o2"]; ok {
		panic("Test failed")
	}

	if _, ok := js.ObjectVals["o4"]; ok {
		panic("Test failed")
	}

	if _, ok := js.ObjectVals["o5"]; ok {
		panic("Test failed")
	}

	if _, ok := js.ObjectVals["o6"]; ok {
		panic("Test failed")
	}

	if _, ok := js.ObjectVals["o7"]; ok {
		panic("Test failed")
	}

	if val, ok := js.ObjectVals["o3"]; !ok || !val.(bool) {
		panic("Test failed")
	}

}

func TestArray(t *testing.T) {

	p := getparser("a")

	var results []*JSON

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}
		results = append(results, json)
	}

	if len(results) != 7 {
		panic("result count must 7")
	}

	if results[0].ValueType != Object {
		panic("Value type must be object")
	}
	if results[1].ValueType != Object {
		panic("Value type must be object")
	}

	if results[2].ValueType != String {
		panic("Value type must be string")
	}

	if results[3].ValueType != Boolean {
		panic("Value type must be bool")
	}

	if results[4].ValueType != Number {
		panic("Value type must be bool")
	}

	if results[5].ValueType != Null {
		panic("Value type must be null")
	}

	if results[6].ValueType != Number {
		panic("Value type must be bool")
	}

	// Skip test
	p = getparser("a").SkipProps([]string{"a11", "a12", "a13"})

	for _, json := range allResult(p) {

		if json.Err != nil {
			panic(json.Err)
		}

		if json.ValueType == Object {

			if _, ok := json.ObjectVals["a11"]; ok {
				panic("Test failed")
			}

			if _, ok := json.ObjectVals["a12"]; ok {
				panic("Test failed")
			}

			if _, ok := json.ObjectVals["a13"]; ok {
				panic("Test failed")
			}

		}

	}

}

func TestArrayOnly(t *testing.T) {

	jsonArrays := [2]string{}
	jsonArrays[0] = `
		{"list":[
											{"Name": "Ed", "Text": "Knock knock."},
											{"Name": "Sam", "Text": "Who's there?"},
											{"Name": "Ed", "Text": "Go fmt."},
											{"Name": "Sam", "Text": "Go fmt ?"},
											{"Name": "Ed", "Text": "Go fmt !"}
									]}
		`
	jsonArrays[1] = "[" + jsonArrays[0] + "]"

	for _, jsarray := range jsonArrays {
		br := bufio.NewReader(bytes.NewReader([]byte(jsarray)))
		p := NewJSONParser(br, "list")
		var results []*JSON
		for _, json := range allResult(p) {

			if json.Err != nil {
				t.Fatal(" Test failed")
			}
			results = append(results, json)
		}
		if results[0].ObjectVals["Text"].(string) != "Knock knock." {
			t.Fatal("results[0] Test failed ")
		}

		if results[1].ObjectVals["Name"].(string) != "Sam" {
			t.Fatal("results[0] Test failed ")
		}

		if results[4].ObjectVals["Name"].(string) != "Ed" {
			t.Fatal("results[0] Test failed ")
		}
	}
}

func TestInvalid(t *testing.T) {

	invalidStart := `{{"Name": "Ed", "Text": "Go fmt."},"s":"valid","s2":in"valid"}`

	br := bufio.NewReader(bytes.NewReader([]byte(invalidStart)))
	p := NewJSONParser(br, "s2")

	for _, json := range allResult(p) {

		if json.Err == nil {
			t.Fatal("Invalid error expected")
		}

	}

	invalidStart2 := `{{"Name": "Ed", "Text": "Go fmt."},"s":in"valid","s2":"valid"}` // invalid in non loop property

	br = bufio.NewReader(bytes.NewReader([]byte(invalidStart2)))
	p = NewJSONParser(br, "s2")

	for _, json := range allResult(p) {

		if json.Err == nil {
			t.Fatal("Invalid error expected")
		}

	}

	invalidEnd := `{"list":[{"Name": "Ed" , "Text": "Go fmt."} , {"Name": "Sam" , "Text": "Go fm"t who?"}]}`

	br = bufio.NewReader(bytes.NewReader([]byte(invalidEnd)))
	p = NewJSONParser(br, "list")
	index := 0
	for _, json := range allResult(p) {

		if index == 1 && json.Err == nil {
			t.Fatal("Invalid error expected")
		}
		index++
	}

}

func TestGetAllNodes(t *testing.T) {
	file, _ := os.Open("sample.json")
	br := bufio.NewReader(file)
	data, _ := ioutil.ReadAll(br)
	dataStr := `{"data":` + string(data) + "}"
	br = bufio.NewReaderSize(bytes.NewReader([]byte(dataStr)), 65536)
	p := NewJSONParser(br, "data")

	for json := range p.Stream() {
		nodes := json.GetAllNodes("o.o7")
		if len(nodes) != 4 {
			t.Errorf("Lenght of json.GetAllNodes is not the expected \n\t Expected: %d \n\t Found: %d", 4, len(nodes))
		} else {
			values := map[string]string{"o71": "o71string", "o73": "true", "o74": "98"}
			for key, expected := range values {
				if nodes[key].GetValue(".") != expected {
					t.Errorf("The value of the key %s doesn´t match with the expected \n\t Expected: %s \n\t Found: %s", key, expected, nodes[key].GetValue("."))
				}
			}
		}
	}
}

func TestGetNode(t *testing.T) {
	var expected, found, path string

	file, _ := os.Open("sample.json")
	br := bufio.NewReader(file)
	data, _ := ioutil.ReadAll(br)
	dataStr := `{"data":` + string(data) + "}"
	br = bufio.NewReaderSize(bytes.NewReader([]byte(dataStr)), 65536)
	p := NewJSONParser(br, "data")

	for json := range p.Stream() {
		path = "f.f1"
		node := json.GetNode(path)
		if node == nil {
			t.Errorf("GetNode Node for path %s is nul", path)
		} else {
			path = "f11"
			expected = "f11value"
			found = node.GetValue(path)
			if found != expected {
				t.Errorf("GetNode %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)

			}
		}

		path = "o.o1"
		node = json.GetNode(path)
		if node == nil {
			t.Errorf("GetNode Node for path %s is nul", path)
		} else {
			path = "."
			expected = "o1string"
			found = node.GetValue(path)
			if found != expected {
				t.Errorf("GetNode %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)

			}
		}

		path = "nu"
		node = json.GetNode(path)
		if !node.IsEmpty() {
			t.Errorf("GetNode Node for path %s is not Empty", path)
		}

		path = "not_exist"
		node = json.GetNode(path)
		if !node.IsEmpty() {
			t.Errorf("GetNode Node for path %s is not Empty", path)
		}
	}
}

func TestGetNodes(t *testing.T) {
	var expected, found, path string
	var index int
	var node *JSON
	var nodes []*JSON

	file, _ := os.Open("sample.json")
	br := bufio.NewReader(file)
	data, _ := ioutil.ReadAll(br)
	dataStr := `{"data":` + string(data) + "}"
	br = bufio.NewReaderSize(bytes.NewReader([]byte(dataStr)), 65536)
	p := NewJSONParser(br, "data")

	for json := range p.Stream() {
		path = "a"
		nodes = json.GetNodes(path)
		if len(nodes) != 7 {
			t.Errorf("GetNodes %s doesn´t match with expected \n\t Expected: %d \n\t Found: %d", path, 7, len(nodes))
		}

		index = 0
		path = "a11"
		expected = "o71string\\"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 0
		path = "a12"
		expected = "o72string"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 0
		path = "a13"
		expected = "true"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 1
		path = "a12[1]"
		expected = ""
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 1
		path = "a12[2]"
		expected = "false"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 2
		path = "."
		expected = "astringinside"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 3
		path = "."
		expected = "false"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 4
		path = "."
		expected = "99"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		index = 6
		path = "."
		expected = "433.33e-6"
		node = nodes[index]
		found = node.GetValue(path)
		if found != expected {
			t.Errorf("Node index %d Path: %s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", index, path, expected, found)
		}

		path = "f.f2"
		nodes = json.GetNodes(path)
		if len(nodes) != 0 {
			t.Errorf("GetNodes %s doesn´t match with expected \n\t Expected: %d \n\t Found: %d", path, 0, len(nodes))
		}
	}
}

func TestGetValue(t *testing.T) {
	var found, expected, path string

	file, _ := os.Open("sample.json")
	br := bufio.NewReader(file)
	data, _ := ioutil.ReadAll(br)
	dataStr := `{"data":` + string(data) + "}"
	br = bufio.NewReaderSize(bytes.NewReader([]byte(dataStr)), 65536)
	p := NewJSONParser(br, "data")

	for json := range p.Stream() {
		expected = ""
		path = "nu"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "o1string"
		path = "o.o1"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "true"
		path = "o.o3"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "98.21"
		path = "o.o5"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "98"
		path = "o.o7.o74"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "o71string"
		path = "o.o7.o71"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "false"
		path = "o.o7.o72[2]"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "o4string"
		path = "o.o4[0]"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "o41string"
		path = "o.o4.o41"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "o71string\\"
		path = "a.a11"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "o71string"
		path = "a[1].a11"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "o72string"
		path = "a[1].a12[0]"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = ""
		path = "a[1].a12[1]"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "false"
		path = "a[1].a12[2]"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

		expected = "98"
		path = "a[1].a12[3]"
		found = json.GetValue(path)
		if found != expected {
			t.Errorf("%s doesn´t match with expected \n\t Expected: %s \n\t Found: %s", path, expected, found)
		}

	}

}

func TestGetValueNumeric(t *testing.T) {
	var i, expectedI int
	var f, expectedF float64
	var path string

	file, _ := os.Open("sample.json")
	br := bufio.NewReader(file)
	data, _ := ioutil.ReadAll(br)
	dataStr := `{"data":` + string(data) + "}"
	br = bufio.NewReaderSize(bytes.NewReader([]byte(dataStr)), 65536)
	p := NewJSONParser(br, "data")

	for json := range p.Stream() {
		expectedI = 2323
		path = "n"
		i = json.GetValueInt(path)
		if i != expectedI {
			t.Errorf("numeric.int %s doesn´t match with expected \n\t Expected: %d \n\t Found: %d", path, expectedI, i)
		}

		expectedI = 98
		path = "o.o7.o74"
		i = json.GetValueInt(path)
		if i != expectedI {
			t.Errorf("numeric.int %s doesn´t match with expected \n\t Expected: %d \n\t Found: %d", path, expectedI, i)
		}

		expectedI = 98
		path = "a[1].a12[3]"
		i = json.GetValueInt(path)
		if i != expectedI {
			t.Errorf("numeric.int %s doesn´t match with expected \n\t Expected: %d \n\t Found: %d", path, expectedI, i)
		}

		expectedF = 23.23
		path = "n1"
		f = json.GetValueF64(path)
		if f != expectedF {
			t.Errorf("numeric.float %s doesn´t match with expected \n\t Expected: %f \n\t Found: %f", path, expectedF, f)
		}
	}
}

func Benchmark1(b *testing.B) {

	for n := 0; n < b.N; n++ {
		p := getparser("a").SkipProps([]string{"a11"})
		for json := range p.Stream() {
			nothing(json)
		}
	}
}

func Benchmark2(b *testing.B) {

	for n := 0; n < b.N; n++ {
		p := getparser("a").SkipProps([]string{"a11"})
		for _, json := range p.Parse() {
			nothing(json)
		}
	}
}

func nothing(j *JSON) {

}
