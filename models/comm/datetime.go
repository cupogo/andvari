package comm

import (
	"encoding/json"
	"time"
)

// copy and optimize from go.mongodb.org/mongo-driver/bson/primitive

// DateTime represents the BSON datetime value.
type DateTime int64

var _ json.Marshaler = DateTime(0)
var _ json.Unmarshaler = (*DateTime)(nil)

// MarshalJSON marshal to time type.
func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time())
}

// UnmarshalJSON creates a primitive.DateTime from a JSON string.
func (d *DateTime) UnmarshalJSON(data []byte) error {
	// Ignore "null" to keep parity with the time.Time type and the standard library. Decoding "null" into a non-pointer
	// DateTime field will leave the field unchanged. For pointer values, the encoding/json will set the pointer to nil
	// and will not defer to the UnmarshalJSON hook.
	if string(data) == "null" {
		return nil
	}

	var tempTime time.Time
	if err := json.Unmarshal(data, &tempTime); err != nil {
		return err
	}

	*d = NewDateTimeFromTime(tempTime)
	return nil
}

// Time returns the date as a time type.
func (d DateTime) Time() time.Time {
	return time.UnixMilli(int64(d))
}

// NewDateTimeFromTime creates a new DateTime from a Time.
func NewDateTimeFromTime(t time.Time) DateTime {
	return DateTime(t.UnixMilli())
}

func (d DateTime) IsZero() bool {
	return d == 0
}

func (d DateTime) Format(layout string) string {
	return d.Time().Format(layout)
}
