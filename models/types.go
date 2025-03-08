package models

import (
	"database/sql"
	"encoding/json"
)

// NullString representa um valor SQL que pode ser nulo e compatível com JSON.
// swagger:ignore
type NullString struct {
	sql.NullString
}

// MarshalJSON customiza a conversão para JSON de NullString.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON customiza a conversão de JSON para NullString.
func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		ns.Valid = true
		ns.String = *s
	} else {
		ns.Valid = false
	}
	return nil
}

// NullInt64 representa um valor SQL que pode ser nulo e compatível com JSON.
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON customiza a conversão para JSON de NullInt64.
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON customiza a conversão de JSON para NullInt64.
func (ni *NullInt64) UnmarshalJSON(data []byte) error {
	var i *int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	if i != nil {
		ni.Valid = true
		ni.Int64 = *i
	} else {
		ni.Valid = false
	}
	return nil
}
