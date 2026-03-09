package common

type Backend struct {
	Address        string `yaml:"Address"`
	Port           int    `yaml:"Port"`
	MaxConnections int    `yaml:"MaxConnections"`
}
