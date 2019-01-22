package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/pokt-network/pocket-core/config"
)

type Whitelist struct {
	Map map[string]struct{}
	mux sync.Mutex
}

var (
	SNWL   Whitelist
	DWL    Whitelist
	wlOnce sync.Once
)

// "WhiteListInit()" initializes both whitelist structures.
func WhiteListInit() {
	wlOnce.Do(func() {
		SNWL.Map = make(map[string]struct{})
		DWL.Map = make(map[string]struct{})
	})
}

// "GetSWL" returns service node white list.
func GetSWL() Whitelist {
	if SNWL.Map == nil { // just in case
		WhiteListInit()
	}
	return SNWL
}

// "GetDWL" returns developer.
func GetDWL() Whitelist {
	if DWL.Map == nil { // just in case
		WhiteListInit()
	}
	return DWL
}

// "Contains" returns if within whitelist.
func (w Whitelist) Contains(s string) bool {
	w.mux.Lock()
	defer w.mux.Unlock()
	_, ok := w.Map[s]
	return ok
}

// "Delete" removes item from whitelist.
func (w Whitelist) Delete(s string) {
	w.mux.Lock()
	defer w.mux.Unlock()
	delete(w.Map, s)
}

// "Add" appends item to whitelist.
func (w Whitelist) Add(s string) {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.Map[s] = struct{}{}
}

// "AddMulti" appends multiple items to whitelist
func (w Whitelist) AddMulti(list []string) {
	for _, v := range list {
		w.Add(v)
	}
}

// "Size" returns the length of the whitelist.
func (w Whitelist) Size() int {
	w.mux.Lock()
	defer w.mux.Unlock()
	return len(w.Map)
}

// "SWLFile" builds the service white list from a file.
func SWLFile() error {
	return GetSWL().wlFile(config.GetInstance().DWL)
}

// "DWLFile" builds the develoeprs white list from a file.
func DWLFile() error {
	return GetDWL().wlFile(config.GetInstance().SNWL)
}

// "wlFile" builds a whitelist structure from a file.
func (w Whitelist) wlFile(filePath string) error {
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	var data []string
	err = json.Unmarshal(f, &data)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	w.AddMulti(data)
	fmt.Println(w)
	return nil
}

// "EnsureWL" cross checks the whitelist for
func EnsureWL(whiteList Whitelist, query string) bool {
	if !whiteList.Contains(query) {
		fmt.Println("Node: ", query, "rejected because it is not within whitelist")
		return false
	}
	return true
}