# where
--
    import "github.com/autom8ter/graphik/functions/where"


## Usage

#### func  All

```go
func All(wheres ...graphik.WhereFunc) graphik.WhereFunc
```
All returns true if the attributes passes all of the provided Where functions

#### func  Any

```go
func Any(wheres ...graphik.WhereFunc) graphik.WhereFunc
```
Any returns true if the attributes passes any of the provided Where functions

#### func  DeepEqual

```go
func DeepEqual(key string, val interface{}) graphik.WhereFunc
```
DeepEqual returns true if the key-val is exactly the same within the attributes
using reflection

#### func  Equals

```go
func Equals(key string, val interface{}) graphik.WhereFunc
```
Equals returns true if the key-val is exactly the same within the attributes

#### func  Exists

```go
func Exists(key string) graphik.WhereFunc
```
Exists returns true if the key exists within the attributes
