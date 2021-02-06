package strutil

import "container/list"

type Keyer interface {
	GetKey() string
}

type MapList struct {
	DataMap  map[string]*list.Element
	DataList *list.List
}

func NewMapList() *MapList {
	return &MapList{
		DataMap:  make(map[string]*list.Element),
		DataList: list.New(),
	}
}

func (mapList *MapList) Exists(data Keyer) bool {
	_, exists := mapList.DataMap[string(data.GetKey())]
	return exists
}

func (mapList *MapList) Push(data Keyer) bool {
	if mapList.Exists(data) {
		return false
	}
	elem := mapList.DataList.PushBack(data)
	mapList.DataMap[data.GetKey()] = elem
	return true
}

func (mapList *MapList) Remove(data Keyer) {
	if !mapList.Exists(data) {
		return
	}
	mapList.DataList.Remove(mapList.DataMap[data.GetKey()])
	delete(mapList.DataMap, data.GetKey())
}

func (mapList *MapList) Size() int {
	return mapList.DataList.Len()
}

func (mapList *MapList) Walk(cb func(data Keyer)) {
	for elem := mapList.DataList.Front(); elem != nil; elem = elem.Next() {
		cb(elem.Value.(Keyer))
	}
}

type Elements struct {
	Key   string
	Value string
}

func (e Elements) GetKey() string {
	return e.Key
}
