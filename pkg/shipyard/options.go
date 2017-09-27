package shipyard

const (
	etcdImage      = "quay.io/coreos/etcd:v3.0.14"
	hyperkubeImage = "gcr.io/google_containers/hyperkube:v1.5.1"
)

type Options struct {
	Prefix  string
	Docker  string
	Kubectl string

	BaseDir string
	WorkDir string

	EtcdImage      string
	HyperkubeImage string
	ClusterIpRange string
}

// DefaultOptions to use to run the e2e test.
func DefaultOptions(baseDir string, workDir string) Options {
	ret := Options{
		Prefix:  "xxx",
		Kubectl: "kubectl",

		BaseDir: baseDir,
		WorkDir: workDir,

		Docker:         "docker",
		EtcdImage:      etcdImage,
		HyperkubeImage: hyperkubeImage,
		ClusterIpRange: "10.0.0.0/24",
	}

	return ret
}
