package db

import (
	"fmt"
	"time"
)

type DateTime time.Time

func (dt *DateTime) Scan(value interface{}) error {
	if value == nil {
		*dt = DateTime{}
		return nil
	}
	switch v := value.(type) {
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return fmt.Errorf("parsing datetime %q: %w", v, err)
		}
		*dt = DateTime(t)
	case []byte:
		t, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return fmt.Errorf("parsing datetime %q: %w", string(v), err)
		}
		*dt = DateTime(t)
	default:
		return fmt.Errorf("unsupported Scan type %T", value)
	}
	return nil
}

func (dt DateTime) String() string {
	return time.Time(dt).Format("2006-01-02 15:04:05")
}

func (dt DateTime) Time() time.Time {
	return time.Time(dt)
}

func (dt DateTime) IsZero() bool {
	return time.Time(dt).IsZero()
}

type NullDateTime struct {
	Time  DateTime
	Valid bool
}

func (n *NullDateTime) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}
	n.Valid = true
	return n.Time.Scan(value)
}