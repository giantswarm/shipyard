package engine

// Result encapsulates information about a SetUp execution
type Result struct {
	ClusterID         string `yaml:"clusterID"`
	InstanceIP        string `yaml:"instanceIP"`
	CaCrtContent      string `yaml:"-"`
	ClientCrtContent  string `yaml:"-"`
	ClientKeyContent  string `yaml:"-"`
	KubeconfigContent string `yaml:"-"`
}

// Engine abstracts the required functionality of a test cluster
type Engine interface {
	SetUp() (*Result, error)
	TearDown(res *Result) (*Result, error)
}
