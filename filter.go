package goseq

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	NOKEY   error = errors.New("The key you requested does not exist!")
	BADTYPE error = errors.New("Filter doesn't know how to encode the value you specified. Use an int, string, or bool.")
)

// Filter allows a client to refine the
// results from the MasterServer
type Filter interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Has(key string) bool
	Keys() []string
	Delete(key string)
	GetFilterFormat() []byte
}

// NewFilter creates Filter with the default
// implementation.
func NewFilter() Filter {
	return &msfilter{parameters: make(map[string]interface{})}
}

// msfilter implements Filter
// Default implementation.
type msfilter struct {
	parameters map[string]interface{}
}

func (fil *msfilter) Set(key string, value interface{}) error {
	switch v := value.(type) {
	case int, bool, string:
		fil.parameters[key] = v
		return nil
	}
	return BADTYPE
}

func (fil *msfilter) Get(key string) (interface{}, error) {
	if !fil.Has(key) {
		return nil, NOKEY
	}
	return fil.parameters[key], nil
}

func (fil *msfilter) Has(key string) bool {
	_, ok := fil.parameters[key]
	if !ok {
		return false
	}
	return true
}

func (fil *msfilter) Delete(key string) {
	delete(fil.parameters, key)
}

func (fil *msfilter) Keys() []string {
	keys := make([]string, len(fil.parameters))
	for key, _ := range fil.parameters {
		keys = append(keys, key)
	}
	return keys
}

func (fil *msfilter) GetFilterFormat() []byte {
	buf := bytes.NewBuffer([]byte{})

	for key, val := range fil.parameters {
		// write key
		buf.WriteString("\\")
		buf.WriteString(key)
		buf.WriteString("\\")
		// write val
		switch v := val.(type) {
		case string:
			buf.WriteString(v)
		case int:
			buf.WriteString(fmt.Sprintf("%d", v))
		case bool:
			if v {
				buf.WriteString("1")
			} else {
				buf.WriteString("0")
			}
		default:
			panic("Filter.Set() should have not allowed this value!")
		}
	}

	buf.WriteByte(0x0) // NULL terminated
	return buf.Bytes()
}
