package openapi

type Settings struct {
	Info    *SettingsInfo
	Servers []*SettingsServer
}

type SettingsInfo struct {
	Title   string
	Version string
}

type SettingsServer struct {
	URL         string
	Description string
}
