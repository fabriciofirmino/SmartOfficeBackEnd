package models

// OnlineStatusData representa os dados de conexão do usuário quando ele está online
type OnlineStatusData struct {
	Id                int    `json:"Id"`
	Username          string `json:"username"`
	StreamDisplayName string `json:"stream_display_name"`
	DateStart         string `json:"date_start"`
	TempoOnline       string `json:"tempo_online"`
	UserAgent         string `json:"user_agent"`
	UserIP            string `json:"user_ip"`
	Container         string `json:"container"`
	GeoIPCountryCode  string `json:"geoip_country_code"`
	ISP               string `json:"isp"`
	City              string `json:"city"`
	Divergence        int    `json:"divergence"`
	StreamIcon        string `json:"stream_icon"`
}
