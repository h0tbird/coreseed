# The Káto system

**Káto** (from Greek *κάτω*: 'down', 'below', 'underneath') is an opinionated system which governs diverse computing workloads and work-flows.
Like in catabolism (from Greek *κάτω* káto, 'downward' and *βάλλειν* ballein, 'to throw'), the *Káto* system is used to breakdown monolithic platforms into microservices.

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

### Deploy Káto on IaaS providers

- [x] [Amazon EC2]()
- [x] [Packet.net]()

#### Deploy the full stack:
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
