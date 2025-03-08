package models

import (
	"database/sql"
	"encoding/json"
)

// 🔹 Converte `sql.NullString` para um tipo JSON serializável
type NullString struct {
	sql.NullString
}

// 🔹 Implementa a conversão para JSON
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

// 🔹 Converte `sql.NullInt64` para um tipo JSON serializável
type NullInt64 struct {
	sql.NullInt64
}

// 🔹 Implementa a conversão para JSON
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	}
	return json.Marshal(nil)
}
