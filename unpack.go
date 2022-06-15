package conveyor

//
type Unpack struct {
	Data []interface{}
}

//
func UnpackData[T any](data []T) Unpack {
	result := make([]interface{}, 0)
	for _, content := range data {
		result = append(result, content)
	}

	return Unpack{Data: result}
}
