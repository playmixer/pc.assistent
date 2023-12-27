package jsonstore

import (
	"encoding/json"
	"log"
	"testing"
)

var (
	db, _                 = New("dataTest")
	testCase []TestStruct = []TestStruct{
		{"test", "nameT", 2},
		{"test1", "name1", 5},
		{"t1", "n1", 1},
		{"t2", "n2", 2},
		{"t1", "n111", 11},
	}
)

type TestStruct struct {
	Val   string `json:"val"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (t TestStruct) Init() {
	db.InitJsonObject(t)
}

func TestInit(t *testing.T) {
	TestStruct{}.Init()
	db.Open(TestStruct{}).Truncate()
	// fmt.Println("cases", testCase)
	for _, v := range testCase {
		db.Open(TestStruct{}).Insert(v)
	}
}

func (t TestStruct) ID() string {
	return t.Val + t.Name
}

func (t TestStruct) Unmarshal(data []byte) (interface{}, error) {
	res := make(map[string]TestStruct)
	err := json.Unmarshal(data, &res)

	return res, err
}

func TestEqual(t *testing.T) {
	db, err := New("dataTest")
	if err != nil {
		t.Fatalf(err.Error())
	}

	res, err := db.Open(TestStruct{}).Where("val", "=", "t1").Get()
	if err != nil {
		log.Fatalf(err.Error())
		t.Fatalf(err.Error())
	}
	if len(res.(map[string]interface{})) != 2 {
		t.Fatalf("TestEqual FAILE")
	}

	// fmt.Println("result", res)
}
func TestWhere1(t *testing.T) {
	db, err := New("dataTest")
	if err != nil {
		t.Fatalf(err.Error())
	}

	res, err := db.Open(TestStruct{}).Where("count", ">", 2).Get()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(res.(map[string]interface{})) != 2 {
		t.Fatalf("TestWhere1 FAILE")
	}

	// fmt.Println("result", res)
}
func TestWhere2(t *testing.T) {
	db, err := New("dataTest")
	if err != nil {
		t.Fatalf(err.Error())
	}

	res, err := db.Open(TestStruct{}).Where("count", "<", 5).Get()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(res.(map[string]interface{})) != 3 {
		t.Fatalf("TestWhere2 FAILE")
	}

	// fmt.Println("result", res)
}
func TestWhereIDEqual(t *testing.T) {
	db, err := New("dataTest")
	if err != nil {
		t.Fatalf(err.Error())
	}

	res, err := db.Open(TestStruct{}).Where("id", "=", "t1n1").Get()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(res.(map[string]interface{})) != 1 {
		t.Fatalf("TestWhereIDEqual FAILE")
	}

	// fmt.Println("result", res)
}
