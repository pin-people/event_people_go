package EventPeople

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func StructToJsonString(object any) string {
	switch object.(type) {
	case string:
		return fmt.Sprintf("%v", object)
	default:
		jsonBody, err := json.Marshal(object)
		FailOnError(err, "Error Marshing object")
		return bytes.NewBuffer(jsonBody).String()
	}
}
