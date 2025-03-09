package models

import (
	"apiBackEnd/config"
	"database/sql"
)

// DashboardResponse representa os dados do dashboard retornados pela API
type DashboardResponse struct {
	TotalClientesRevenda int `json:"total_clientes_revenda"`
	TotalTestesAtivos    int `json:"total_testes_ativos"`
	TotalVencido         int `json:"total_vencido"`
	TotalClientes        int `json:"total_clientes"`
}

// ObterDadosDashboard executa as procedures SQL e retorna os dados do dashboard
func ObterDadosDashboard(memberID int) (*DashboardResponse, error) { // ✅ Nome corrigido
	var counts DashboardResponse

	procedures := []struct {
		query string
		dest  *int
	}{
		{"CALL totalClientesRevenda(?);", &counts.TotalClientesRevenda},
		{"CALL totalTestesAtivos(?);", &counts.TotalTestesAtivos},
		{"CALL totalVencido(?);", &counts.TotalVencido},
		{"CALL totalClientes(?);", &counts.TotalClientes},
	}

	for _, proc := range procedures {
		err := config.DB.QueryRow(proc.query, memberID).Scan(proc.dest)
		if err != nil {
			if err == sql.ErrNoRows {
				*proc.dest = 0
			} else {
				return nil, err
			}
		}
	}

	return &counts, nil
}
