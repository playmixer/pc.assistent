# JsonStore

```golang

var (
    db, _= New("data")
)


type TestStruct struct {
	Val   string `json:"val"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (t TestStruct) Init() {
	db.InitJsonObject(t)
}

func (t TestStruct) ID() string {
	return t.Val + t.Name
}

func (t TestStruct) Unmarshal(data []byte) (interface{}, error) {
	res := make(map[string]TestStruct)
	err := json.Unmarshal(data, &res)

	return res, err
}

func main() {
	err := db.Open(TestStruct{}).Insert(TestStruct{"test", "nameT", 2})
	if err != nil {
		log.Fatalf(err.Error())
	}


    //res = map[string]interface{}
	res, err := db.Open(TestStruct{}).Where("val", "=", "test").Get()
	if err != nil {
		log.Fatalf(err.Error())
	}
    ...

    // удалить объект
	err := db.Open(TestStruct{}).Remove(TestStruct{"test", "nameT", 2})
    ...


    // получить все данные
    //res = map[string]interface{}
	res, err := db.Open(TestStruct{}).All()
    ...
    
    //обновить, после обновление всех даннх так же обновится id
    //res = map[string]interface{}
	res, err := db.Open(TestStruct{}).Update(id, TestStruct{"test", "nameT", 2})
    ...

    //выбрать данные
    //res = map[string]interface{}
    res, err := db.Open(TestStruct{}).Where("id", "=", "testnameT").Get()
    ...
    res, err := db.Open(TestStruct{}).Where("count", ">", 1).Get()
    ...
}

```