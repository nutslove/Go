package config

type Metrics struct {
	NameSpace  string
	MetricName string
	Query      string
	MetricHelp string
	Labels     []string
}
