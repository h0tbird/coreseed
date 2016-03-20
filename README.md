# The Káto system

**Káto** (from Greek *κάτω*: 'down', 'below', 'underneath') is an opinionated system which governs diverse computing workloads and work-flows.
Like in catabolism (from Greek *κάτω* káto, 'downward' and *βάλλειν* ballein, 'to throw'), the *Káto* system is used to breakdown complex monolithic platforms into simpler microservices.

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

</br>

# Overview

[CoreOS]() is the foundation on which *Káto* is built. It provides the fundamental components used to assemble container-based distributed systems: [etcd]() is used for consensus and discovery, [fleet]() is a distributed init system, [flannel]() is used for virtual networking and [rkt]() and [docker]() are container engines.

All this *CoreOS* goodies are used to bootstrap a [Mesos]() cluster. *Mesos* is a distributed systems kernel which abstracts compute resources away from machines. Accordingly, it provides schedulers (or frameworks in Mesos parlance) which can run on top in order to utilise the exposed compute resources.

## 1. Deploy Káto's infrastructure

- [x] [Amazon EC2](https://github.com/h0tbird/coreseed/blob/master/docs/ec2.md)
- [x] [Packet.net](https://github.com/h0tbird/coreseed/blob/master/docs/packet.md)

## 2. Start the stack
```
cd /etc/fleet
fleetctl submit zookeeper\@.service
fleetctl start zookeeper@1 zookeeper@2 zookeeper@3
fleetctl start mesos-master.service mesos-dns.service
fleetctl start marathon.service cadvisor.service
fleetctl start ceph-mon.service dnsmasq.service
fleetctl start ceph-osd.service mesos-node.service
```

## License
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
