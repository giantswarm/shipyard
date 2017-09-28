package shipyard

const (
	etcdImage      = "quay.io/coreos/etcd:v3.2.7"
	hyperkubeImage = "gcr.io/google_containers/hyperkube:v1.7.6"
)

type Options struct {
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
	return Options{
		Kubectl: "kubectl",

		BaseDir: baseDir,
		WorkDir: workDir,

		Docker:         "docker",
		EtcdImage:      etcdImage,
		HyperkubeImage: hyperkubeImage,
		ClusterIpRange: "10.0.0.0/24",
	}
}
