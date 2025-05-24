package models

import "time"

// AuditLogEntry representa uma entrada de log de auditoria.
type AuditLogEntry struct {
	Action    string                 `bson:"action"`
	UserID    int                    `bson:"user_id"`
	AdminID   int                    `bson:"admin_id"` // Assumindo que o MemberID do admin logado será usado aqui
	Timestamp time.Time              `bson:"timestamp"`
	Details   map[string]interface{} `bson:"details,omitempty"`
}

// UserStatusPayload é usado para ativar/desativar um usuário.
type UserStatusPayload struct {
	Enabled bool `json:"enabled"`
}

// UserRegionPayload é usado para forçar a região de um usuário.
type UserRegionPayload struct {
	ForcedCountry string `json:"forced_country" binding:"required,len=2"` // Ex: "US"
}

// SoftDeletePayload é usado para a exclusão lógica de um usuário.
type SoftDeletePayload struct {
	DeleteReason string `json:"delete_reason" binding:"required"`
}

// DeletedUser representa um usuário que foi logicamente excluído.
type DeletedUser struct {
	ID             int        `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	MemberID       int        `json:"member_id"`
	DeletedAt      *time.Time `json:"deleted_at"`
	DeletedBy      *int       `json:"deleted_by"`
	DeleteReason   *string    `json:"delete_reason"`
	ExpDate        *int64     `json:"exp_date"`
	LastLogin      *time.Time `json:"last_login"`
	MaxConnections int        `json:"max_connections"`
}
