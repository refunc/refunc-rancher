# bytesobj
--
    import "git.v87.us/pkg/bytesobj"


## Usage

```go
var (
	ErrInvaildObject    = errors.New("bytesobj: object is nil")
	ErrNotSupportedJSON = errors.New("bytesobj: not supported json type")
)
```
common errors

#### type BytesObj

```go
type BytesObj struct {
}
```

BytesObj is usefult to represent a JSON field, whose value may be an object,
bytes

#### func  MustFromObject

```go
func MustFromObject(obj interface{}) *BytesObj
```
MustFromObject creates new from a object

#### func  New

```go
func New(raw []byte) (*BytesObj, error)
```
New creates a object from raw bytes

#### func  NewFromBytes

```go
func NewFromBytes(bts []byte) *BytesObj
```
NewFromBytes creates new from bytes

#### func  NewFromObject

```go
func NewFromObject(obj interface{}) (*BytesObj, error)
```
NewFromObject creates new from a object

#### func  NewFromString

```go
func NewFromString(str string) *BytesObj
```
NewFromString creates new from string

#### func (*BytesObj) Bytes

```go
func (bo *BytesObj) Bytes() []byte
```
Bytes retuns raw bytes of the object

#### func (*BytesObj) GJSONParse

```go
func (bo *BytesObj) GJSONParse() (res gjson.Result)
```
GJSONParse parse underlying data using gjson

#### func (*BytesObj) IsObject

```go
func (bo *BytesObj) IsObject() bool
```
IsObject returns true if the underlying data is a JSON object

#### func (*BytesObj) MarshalJSON

```go
func (bo *BytesObj) MarshalJSON() (bts []byte, err error)
```
MarshalJSON implements JSON's Marshaler interface

#### func (*BytesObj) ObjectGetInline

```go
func (bo *BytesObj) ObjectGetInline(path string) (res gjson.Result)
```
ObjectGetInline gets fields using gjson inline

#### func (*BytesObj) String

```go
func (bo *BytesObj) String() string
```

#### func (*BytesObj) UnmarshalJSON

```go
func (bo *BytesObj) UnmarshalJSON(data []byte) error
```
UnmarshalJSON implements JSON's Unmarshaler interface

#### func (*BytesObj) UnmarshalObject

```go
func (bo *BytesObj) UnmarshalObject(v interface{}) error
```
UnmarshalObject unmarshals the raw bytes to given object
