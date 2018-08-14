package bytesobj

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"

	"github.com/tidwall/gjson"
)

// BytesObj is usefult to represent a JSON field,
// whose value may be an object, bytes
type BytesObj struct {
	tp uint8

	data []byte
}

const (
	isObj uint8 = 0x1
	isBts uint8 = 0x2
)

// common errors
var (
	ErrInvaildObject    = errors.New("bytesobj: object is nil")
	ErrNotSupportedJSON = errors.New("bytesobj: not supported json type")
)

// New creates a object from raw bytes
func New(raw []byte) (*BytesObj, error) {
	bo := new(BytesObj)
	if err := bo.UnmarshalJSON(raw); err != nil {
		return nil, err
	}
	return bo, nil
}

// MustFromObject creates new from a object
func MustFromObject(obj interface{}) *BytesObj {
	bo, err := NewFromObject(obj)
	if err != nil {
		panic(err)
	}
	return bo
}

// NewFromObject creates new from a object
func NewFromObject(obj interface{}) (*BytesObj, error) {
	bts, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return New(bts)
}

// NewFromBytes creates new from bytes
func NewFromBytes(bts []byte) *BytesObj {
	return &BytesObj{
		tp:   isBts,
		data: bts,
	}
}

// NewFromString creates new from string
func NewFromString(str string) *BytesObj {
	return &BytesObj{
		data: []byte(str),
	}
}

// IsObject returns true if the underlying data is a JSON object
func (bo *BytesObj) IsObject() bool {
	if bo == nil {
		return false
	}
	return bo.tp&isObj == isObj
}

// UnmarshalObject unmarshals the raw bytes to given object
func (bo *BytesObj) UnmarshalObject(v interface{}) error {
	if bo == nil {
		return ErrInvaildObject
	}
	if !bo.IsObject() {
		return ErrNotSupportedJSON
	}
	return json.Unmarshal(bo.data, v)
}

// ObjectGetInline gets fields using gjson inline
func (bo *BytesObj) ObjectGetInline(path string) (res gjson.Result) {
	if bo != nil && bo.IsObject() && len(bo.data) > 0 {
		res = gjson.GetBytes(bo.data, path)
	}
	return
}

// GJSONParse parse underlying data using gjson
func (bo *BytesObj) GJSONParse() (res gjson.Result) {
	if bo != nil && bo.IsObject() && len(bo.data) > 0 {
		unsafeStr := *(*string)(unsafe.Pointer(&bo.data))
		res = gjson.Parse(unsafeStr)
	}
	return
}

// Bytes retuns raw bytes of the object
func (bo *BytesObj) Bytes() []byte {
	if bo == nil {
		return nil
	}
	return bo.data
}

// UnmarshalJSON implements JSON's Unmarshaler interface
func (bo *BytesObj) UnmarshalJSON(data []byte) error {
	if bo == nil {
		return ErrInvaildObject
	}

	first, last := data[0], data[len(data)-1]
	if (first == '{' && last == '}') || (first == '[' && last == ']') {
		bo.tp = isObj
		bo.data = data
		return nil
	} else if data[0] == '"' {
		data = data[1 : len(data)-1]

		if len(data) == 0 {
			// empty
			bo.data = nil
			return nil
		}

		// try to decode as base64
		unsafeStr := *(*string)(unsafe.Pointer(&data))
		if bts, err := base64.StdEncoding.DecodeString(unsafeStr); err == nil {
			bo.tp = isBts
			data = bts
		}

		bo.data = data // the underlying data is string
		return nil
	}
	return ErrNotSupportedJSON
}

// MarshalJSON implements JSON's Marshaler interface
func (bo *BytesObj) MarshalJSON() (bts []byte, err error) {
	if bo == nil {
		return nil, ErrInvaildObject
	}

	if bo.IsObject() {
		if bo.data != nil {
			bts = bo.data
			return
		}
		bts = []byte{'{', '}'}
		return
	}
	if len(bo.data) == 0 {
		bts = []byte{'"', '"'}
		return
	}
	if bo.tp != 0x0 {
		return json.Marshal(bo.data)
	}

	unsafeStr := *(*string)(unsafe.Pointer(&bo.data))
	return json.Marshal(unsafeStr)
}

func (bo *BytesObj) String() string {
	if bo.IsObject() {
		return string(bo.Bytes())
	}
	return fmt.Sprintf("bytes<%d>", len(bo.Bytes()))
}

// DeepCopy returns a deep copy of underlying object
func (bo *BytesObj) DeepCopy() *BytesObj {
	if bo == nil {
		return nil
	}
	out := new(BytesObj)
	bo.DeepCopyInto(out)
	return out
}

// DeepCopyInto creates a deep copy of underlying object to the given
func (bo *BytesObj) DeepCopyInto(out *BytesObj) {
	out.tp = bo.tp
	if len(bo.data) > 0 {
		out.data = make([]byte, len(bo.data))
		copy(out.data, bo.data)
	}
	return
}
