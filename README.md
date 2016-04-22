# The Káto system

[![License Widget]][License] [![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis] [![Coverage Status Widget]][Coverage Status]

[License]: http://www.apache.org/licenses/LICENSE-2.0.txt
[License Widget]: https://img.shields.io/badge/license-APACHE2-1eb0fc.svg
[GoReportCard]: https://goreportcard.com/report/h0tbird/kato
[GoReportCard Widget]: https://goreportcard.com/badge/h0tbird/kato
[Travis]: https://travis-ci.org/h0tbird/kato
[Travis Widget]: https://travis-ci.org/h0tbird/kato.svg?branch=master
[Coverage Status]: https://coveralls.io/github/h0tbird/kato?branch=master
[Coverage Status Widget]: https://coveralls.io/repos/github/h0tbird/kato/badge.svg?branch=master

**Káto** (from Greek *κάτω*: 'down', 'below', 'underneath') is an opinionated software-defined infrastructure (*SDI*) which governs diverse computing workloads and work-flows.
Like in catabolism (from Greek *κάτω* káto, 'downward' and *βάλλειν* ballein, 'to throw'), the *Káto* system is the catalyst used to breakdown complex monolithic platforms into its fundamental microservices.

</br>

<img src="https://raw.githubusercontent.com/h0tbird/coreseed/master/imgs/kato.png"
 alt="Booddies logo" title="Booddies" align="right" width="69%" height="69%"/>

**Distinctive attributes**

- Geolocation
- Multidatacenter
- Cloud agnostic
- Variable costs
- Hardware abstraction
- Endo/exo-elasticity
- Microservices
- Containerization
- Task scheduling
- CI/CD pipelines
- Service discovery
- Load balancing
- High availability
- Self-healing

</br>

# Overview

[CoreOS](https://coreos.com/) is the foundation on which *Káto* is built. It provides the fundamental components used to assemble container-based distributed systems: [etcd](https://github.com/coreos/etcd) is used for consensus and discovery, [fleet](https://github.com/coreos/etcd) is a distributed init system, [flannel](https://github.com/coreos/flannel) is used for virtual networking and [rkt](https://github.com/coreos/rkt) and [docker](https://github.com/docker/docker) are container engines.

All this *CoreOS* goodies are used to bootstrap a [Mesos](https://github.com/apache/mesos) cluster. *Mesos* is a distributed systems kernel which abstracts compute resources away from machines. Accordingly, it provides schedulers (or frameworks in *Mesos* parlance) which can run on top in order to utilise the exposed compute resources.

[Marathon](https://github.com/mesosphere/marathon) is one of such frameworks. It is a cluster-wide init and control system for long-running applications. Other frameworks like [Jenkins](https://github.com/jenkinsci/mesos-plugin) and [Elasticsearch ](https://github.com/mesos/elasticsearch) might share the same cluster resources.

[Ceph](https://github.com/ceph/ceph-docker) is a distributed object store and file system designed to provide excellent performance, reliability and scalability. It provides a multi-host persistent data storage for containers.

# Components

|Component|Current Version|Container|
|---|:---:|:---:|
|[CoreOS](https://coreos.com)|[alpha](https://coreos.com/releases/)|-|
|[Mesos](http://mesos.apache.org)|[0.27.2](https://git-wip-us.apache.org/repos/asf?p=mesos.git;a=blob_plain;f=CHANGELOG;hb=0.27.2)|[![Docker Pulls](https://img.shields.io/docker/pulls/mesosphere/mesos-master.svg)](https://hub.docker.com/r/mesosphere/mesos-master/)|
|[Mesos-DNS](http://mesosphere.github.io/mesos-dns)|[0.5.2](https://github.com/mesosphere/mesos-dns/releases/tag/v0.5.2)|[![Docker Pulls](https://img.shields.io/docker/pulls/mesosphere/mesos-dns.svg)](https://hub.docker.com/r/mesosphere/mesos-dns/)|
|[Marathon](https://mesosphere.github.io/marathon)|[0.15.3](https://github.com/mesosphere/marathon/releases/tag/v0.15.3)|[![Docker Pulls](https://img.shields.io/docker/pulls/mesosphere/marathon.svg)](https://hub.docker.com/r/mesosphere/marathon/)|
|[Zookeeper](https://zookeeper.apache.org)|[3.4.8](https://zookeeper.apache.org/doc/r3.4.8/)|[![Docker Pulls](https://img.shields.io/docker/pulls/h0tbird/zookeeper.svg)](https://hub.docker.com/r/h0tbird/zookeeper/)|
|[go-dnsmasq](https://github.com/janeczku/go-dnsmasq)|[1.0.0](https://github.com/janeczku/go-dnsmasq/releases/tag/1.0.0)|[![Docker Pulls](https://img.shields.io/docker/pulls/janeczku/go-dnsmasq.svg)](https://hub.docker.com/r/janeczku/go-dnsmasq/)|
|[cAdvisor](https://github.com/google/cadvisor)|[0.22.0](https://github.com/google/cadvisor/releases/tag/v0.22.0)|[![Docker Pulls](https://img.shields.io/docker/pulls/google/cadvisor.svg)](https://hub.docker.com/r/google/cadvisor/)|
|[Ceph](http://ceph.com)|[9.2.0](https://github.com/h0tbird/docker-ceph/releases/tag/v9.2.0-2)|[![Docker Pulls](https://img.shields.io/docker/pulls/h0tbird/ceph.svg)](https://hub.docker.com/r/h0tbird/ceph/)|

## 1. Install katoctl

##### From the source (for *Káto* developers)
```bash
marc@desk-1 ~ $ go get github.com/h0tbird/kato/cmd/katoctl
marc@desk-1 ~ $ go install github.com/h0tbird/kato/cmd/katoctl
marc@desk-1 ~ $ eval "$(katoctl --completion-script-${0#-})"
```

##### From the latest release (for *Káto* end users)
```bash
not yet
```

## 2. Deploy Káto's infrastructure

*Káto* can be deployed on a few *IaaS* providers. More providers are planned but feel free to send a pull request if your prefered provider is not supported yet. Find below deployment guides for each supported provider:

|:white_check_mark:|:white_check_mark:|:white_check_mark:|:x:|:x:|
|---|---|---|---|---|
|[Vagrant](https://github.com/h0tbird/coreseed/blob/master/docs/vagrant.md)|[Packet.net](https://github.com/h0tbird/coreseed/blob/master/docs/packet.md)|[Amazon EC2](https://github.com/h0tbird/coreseed/blob/master/docs/ec2.md)|[Google GCE]()|[Digital Ocean]()|

## 3. Pre-flight checklist
Once you have deployed the infrastructure, run sanity checks to evaluate whether the cluster is ready for normal operation. This can be done on any node type `master`, `node` or `edge`:
```bash
marc@desk-1 ~ $ ssh -A core@edge-1.ext.<your.domain>
core@edge-1 ~ $ etcdctl cluster-health
core@edge-1 ~ $ fleetctl list-machines
core@edge-1 ~ $ loopssh sudo etchost
core@edge-1 ~ $ loopssh ip r | grep docker
core@edge-1 ~ $ watch "fleetctl list-units"
```

## 4. Start the stack
Open a second terminal to `edge-1` (bastion host) and jump to `master-1` from there. If you are using vagrant you can ssh directly to `master-1` instead. Also, enable forwarding of the authentication agent (`ssh -A`):
```bash
marc@desk-1 ~ $ ssh -A core@edge-1.ext.<your.domain>
core@edge-1 ~ $ ssh master-1
```

Use `fleetctl` to start the service units while you check the status on the first terminal:
```bash
core@master-1 ~ $ cd /etc/fleet
core@master-1 ~ $ fleetctl start zookeeper.service
core@master-1 ~ $ fleetctl start mesos-master.service mesos-dns.service
core@master-1 ~ $ fleetctl start marathon.service cadvisor.service
core@master-1 ~ $ fleetctl start dnsmasq.service mesos-node.service
core@master-1 ~ $ fleetctl start marathon-lb-int.service marathon-lb-ext.service
core@master-1 ~ $ fleetctl start mongodb.service pritunl.service
```

## 5. Setup pritunl
*Pritunl* is an OpenVPN server that provides secure access to *Káto*'s private network.
Access your pritunl WebGUI at `http://edge-1.ext.<your.domain>`
