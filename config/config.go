// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period time.Duration `config:"period"`
	Path   string        `config:"mqsiscriptpath"` // This is the path for mqsi scripts.
}

var DefaultConfig = Config{
	Period: 1 * time.Second,
	Path:   ".",
}

//added IIB beat related structs
type NodeCollection struct {
    Nodes []Node
}
type Node struct {
    Name    string
    Status string
	IntegrationServers []IntegrationServer
}

type IntegrationServer struct {
	Name	string
	Status	string
	Components	[]Component
}
type Component struct {
	Name	string
	Status	string
	Type	string
	DeployedTime string
	BarFileName	string
	BarFileLastModifiedTime	string
}