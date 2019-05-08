# Shipyard

[![CircleCI](https://circleci.com/gh/giantswarm/shipyard.svg?style=shield)](https://circleci.com/gh/giantswarm/shipyard)

Creates thin kubernetes environments for automated testing.

## Getting Project

Clone the git repository: https://github.com/giantswarm/shipyard.git

## Running Shipyard

Shipyard let's you create a remote minikube instance and configures the local
environment for connecting to it. It currently supports AWS as the remote
provider, so in order to making it work you need first to do some preparations
on the cloud side:

### AWS setup

These are the requirements for Shipyard to work:

* AWS credentials loaded into the environment (`AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY` defined).

* AMI image suitable to be used by shipyard. It can be created using [packer](https://www.packer.io/)
executing this command from the root of the project:

```bash
$ packer build ./image/minikube.json
```
This will create a private AMI ready to be used by shipyard, using the same AWS
credentials for packer and shipyard you are good to go.

### Running as a command line utility

You can download a prebuilt binary from [here](https://github.com/giantswarm/shipyard/releases/). Having that in your path you
can create a shipyard instance with:

```bash
$ shipyard
```
After setting `KUBECONFIG` to `./.shipyard/config`  kubectl access the remote cluster:

```bash
$ export KUBECONFIG=./.shipyard/config
$ kubectl get cs
NAME                 STATUS    MESSAGE              ERROR
scheduler            Healthy   ok
controller-manager   Healthy   ok
etcd-0               Healthy   {"health": "true"}
```

Remember to delete the remote cluster with:

```bash
$ shipyard -action stop
```

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/shipyard/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING.md](/giantswarm/shipyard/blob/master/CONTRIBUTING.md) for details on submitting patches, the contribution workflow as well as reporting bugs.

## License

Shipyard is under the Apache 2.0 license. See the [LICENSE](/giantswarm/shipyard/blob/master/LICENSE) file for details.
