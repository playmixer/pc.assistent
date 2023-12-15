package jsonstore

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

type ErrJsonStore string

func (e ErrJsonStore) Error() string {
	return string(e)
}

const (
	ErrNotFoundRow ErrJsonStore = "Not found row"
)

type Driver struct {
	Dir   string
	sufix string
	mut   *sync.RWMutex
}

type jObject interface {
	Init()
	ID() string
	Unmarshal(data []byte) (interface{}, error)
}

func New(dir string) (*Driver, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return &Driver{
		Dir:   dir,
		sufix: "json",
		mut:   &sync.RWMutex{},
	}, nil
}

func (d *Driver) InitJsonObject(obj interface{}) error {
	name := d.nameOfObject(obj)
	path := d.pathJsonObject(name)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		f.Write([]byte("{}"))
	} else if err != nil {
		return err
	}
	return nil
}

func (d *Driver) pathJsonObject(name string) string {
	return filepath.Join(d.Dir, name+"."+d.sufix)
}

func (d *Driver) nameOfObject(obj interface{}) string {
	if t := reflect.TypeOf(obj); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

type DriverWhere struct {
	Field string
	Comp  string
	Val   interface{}
}

type DriverOpen struct {
	driver *Driver
	model  jObject
	where  []DriverWhere
}

func (d *Driver) Open(model jObject) *DriverOpen {
	return &DriverOpen{
		d,
		model,
		[]DriverWhere{},
	}
}

func (d *DriverOpen) Truncate() error {
	name := d.driver.nameOfObject(d.model)
	path := d.driver.pathJsonObject(name)
	d.driver.mut.RLock()
	defer d.driver.mut.RUnlock()
	f, err := os.OpenFile(path, os.O_TRUNC, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	return nil
}

func (d *DriverOpen) Where(field, comp string, val interface{}) *DriverOpen {
	d.where = append(d.where, DriverWhere{
		field,
		comp,
		val,
	})

	return d
}

func (d *DriverOpen) filtering(data interface{}) (interface{}, error) {

	if len(d.where) > 0 {
		res := make(map[string]interface{})
		_b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		_obj := make(map[string]map[string]interface{})
		err = json.Unmarshal(_b, &_obj)
		if err != nil {
			return nil, err
		}
		for k, v := range _obj {
			v["id"] = k
			isSelected := true
			for _, w := range d.where {
				// fmt.Println(v[w.Field], w.Comp, w.Val)
				if w.Comp == "=" {
					isSelected = isSelected && reflect.DeepEqual(v[w.Field], w.Val)
				} else {
					switch _v := v[w.Field].(type) {
					case int:
						v[w.Field] = float64(_v)
					}
					switch _v := w.Val.(type) {
					case int:
						w.Val = float64(_v)
					}
					switch w.Val.(type) {
					case float64:
						if w.Comp == ">" {
							isSelected = isSelected && v[w.Field].(float64) > w.Val.(float64)
						} else if w.Comp == "<" {
							isSelected = isSelected && v[w.Field].(float64) < w.Val.(float64)
						}
					}

				}

			}
			if isSelected {
				res[k] = v
			}
		}
		return res, nil
	}
	return data, nil
}

// return map[string]object
func (d *DriverOpen) Get() (interface{}, error) {
	obj, err := d.All()
	if err != nil {
		return nil, err
	}

	return d.filtering(obj)
}

// return map[string]object
func (d *DriverOpen) All() (interface{}, error) {
	name := d.driver.nameOfObject(d.model)
	path := d.driver.pathJsonObject(name)
	d.driver.mut.RLock()
	defer d.driver.mut.RUnlock()
	f, err := os.OpenFile(path, os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	obj, err := d.model.Unmarshal(b)
	// obj := make(map[string]interface{})
	// err = json.Unmarshal(b, &obj)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return obj, nil
}

func (d *DriverOpen) Insert(data jObject) error {
	name := d.driver.nameOfObject(d.model)
	path := d.driver.pathJsonObject(name)
	d.driver.mut.Lock()
	defer d.driver.mut.Unlock()
	f, err := os.OpenFile(path, os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	var obj map[string]interface{}

	b, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return err
	}
	f.Close()

	err = json.Unmarshal(b, &obj)
	if err != nil {
		log.Println(err)
		return err
	}
	_map := obj
	_map[data.ID()] = data
	res, err := json.MarshalIndent(_map, "", "  ")
	if err != nil {
		log.Println(err)
		return err
	}

	err = os.WriteFile(path, res, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (d *DriverOpen) Remove(data jObject) error {
	name := d.driver.nameOfObject(d.model)
	path := d.driver.pathJsonObject(name)
	d.driver.mut.Lock()
	defer d.driver.mut.Unlock()
	f, err := os.OpenFile(path, os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	var obj map[string]interface{}

	b, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return err
	}
	f.Close()

	err = json.Unmarshal(b, &obj)
	if err != nil {
		log.Println(err)
		return err
	}
	_map := obj
	if _, ok := _map[data.ID()]; !ok {
		return ErrNotFoundRow
	}
	delete(_map, data.ID())
	res, err := json.MarshalIndent(_map, "", "  ")
	if err != nil {
		log.Println(err)
		return err
	}

	err = os.WriteFile(path, res, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (d *DriverOpen) Delete() error {
	name := d.driver.nameOfObject(d.model)
	path := d.driver.pathJsonObject(name)
	d.driver.mut.Lock()
	defer d.driver.mut.Unlock()
	f, err := os.OpenFile(path, os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	var obj map[string]interface{}

	b, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return err
	}
	f.Close()

	err = json.Unmarshal(b, &obj)
	if err != nil {
		log.Println(err)
		return err
	}
	//находим записи для удаления
	_del, err := d.filtering(obj)
	if err != nil {
		log.Println(err)
		return err
	}
	//очищаем мапу
	for k := range _del.(map[string]interface{}) {
		delete(obj, k)
	}
	res, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Println(err)
		return err
	}

	err = os.WriteFile(path, res, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (d *DriverOpen) Update(id string, data jObject) (interface{}, error) {
	err := d.Where("id", "=", id).Delete()
	// err := d.Remove(data)
	if err != nil {
		return nil, err
	}
	err = d.Insert(data)
	if err != nil {
		return nil, err
	}

	return d.Where("id", "=", data.ID()).Get()
}
