package presenter

import "encoding/json"

// jsonMarshal is a variable wrapping json.Marshal to allow replacement in tests.
var jsonMarshal = json.Marshal
