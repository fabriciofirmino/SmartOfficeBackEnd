package models

import (
	"apiBackEnd/config"
	"database/sql"
)

// DashboardResponse representa os dados do dashboard retornados pela API
type DashboardResponse struct {
	TotalClientesRevenda int    `json:"total_clientes_revenda"`
	TotalTestesAtivos    int    `json:"total_testes_ativos"`
	TotalVencido         int    `json:"total_vencido"`
	TotalClientes        int    `json:"total_clientes"`
	TotalFilmes          int    `json:"total_filmes"`
	TotalEpisodiosSeries int    `json:"total_episodios_series"`
	TotalCanais          int    `json:"total_canais"`
	TotalSeries          int    `json:"total_series"`
	CanaisOff            int    `json:"canais_off"`
	LastUpdated          string `json:"last_updated"` // Ou sql.NullString se puder ser nulo e precisar de tratamento especial
}

// ObterDadosDashboard executa as procedures SQL e a query de streams, retornando os dados do dashboard
func ObterDadosDashboard(memberID int) (*DashboardResponse, error) {
	var counts DashboardResponse

	// Obter dados das procedures
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
				*proc.dest = 0 // Define como 0 se não houver linhas
			} else {
				return nil, err
			}
		}
	}

	// Obter dados da tabela aggregated_streams_counts
	streamsQuery := `SELECT TotalFilmes, TotalEpisodiosSeries, TotalCanais, TotalSeries, CanaisOff, last_updated FROM aggregated_streams_counts LIMIT 1;`
	// Usar sql.NullString para last_updated se puder ser NULL no banco, para evitar erro no Scan
	var lastUpdated sql.NullString
	err := config.DB.QueryRow(streamsQuery).Scan(
		&counts.TotalFilmes,
		&counts.TotalEpisodiosSeries,
		&counts.TotalCanais,
		&counts.TotalSeries,
		&counts.CanaisOff,
		&lastUpdated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Se não houver dados na tabela aggregated_streams_counts, defina os campos como 0 ou string vazia
			counts.TotalFilmes = 0
			counts.TotalEpisodiosSeries = 0
			counts.TotalCanais = 0
			counts.TotalSeries = 0
			counts.CanaisOff = 0
			counts.LastUpdated = ""
		} else {
			return nil, err // Retorna erro se for algo diferente de ErrNoRows
		}
	} else {
		if lastUpdated.Valid {
			counts.LastUpdated = lastUpdated.String
		} else {
			counts.LastUpdated = "" // Ou algum valor padrão se for NULL
		}
	}

	return &counts, nil
}
