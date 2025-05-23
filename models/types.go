package models

import (
	"database/sql"
	"encoding/json"
	"time"
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

// Estruturas para controle de confiança, rollback e alteração de vencimento via Redis

type TrustBonus struct {
	DiasAdicionados int       `json:"dias_adicionados"`
	DataLiberacao   time.Time `json:"data_liberacao"`
	AdminID         string    `json:"admin_id"`
	Motivo          string    `json:"motivo"`
}

type RenewBackup struct {
	ExpDateAnterior int64     `json:"exp_date_anterior"`
	CreditosGastos  float64   `json:"creditos_gastos"`
	DataRenovacao   time.Time `json:"data_renovacao"`
	AdminRenovou    string    `json:"admin_renovou"`
}

type ChangeDueDate struct {
	NovaDataVencimento int       `json:"nova_data_vencimento"`
	DataAlteracao      time.Time `json:"data_alteracao"`
	AdminID            string    `json:"admin_id"`
	Motivo             string    `json:"motivo"`
}
