# The Káto system

**Káto** (from Greek *κάτω*: 'down', 'below', 'underneath') is an opinionated system which governs diverse computing workloads and work-flows.
Like in catabolism (from Greek *κάτω* káto, 'downward' and *βάλλειν* ballein, 'to throw'), the *Káto* system is used to breakdown monolithic platforms into microservices.

</br>

<img src="https://www.lucidchart.com/publicSegments/view/36b62f8a-cb78-4807-850f-d0df68e94bd7/image.png"
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

### Deploy on Packet.net
```bash
#!/bin/bash

case $1 in

  "masters")

    ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')

    for i in 1 2 3; do

      coreseed udata \
      --hostname core-${i} \
      --domain cell-1.dc-1.demo.com \
      --role master \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --etcd-token ${ETCD_TOKEN} |

      coreseed run-packet \
      --api-key xxx \
      --hostname core-${i} \
      --project-id xxx \
      --plan baremetal_0 \
      --os coreos_alpha \
      --facility ewr1 \
      --billing hourly

    done ;;

  "nodes")

    for i in 4 5 6; do

      coreseed udata \
      --hostname core-${i} \
      --domain cell-1.dc-1.demo.com \
      --role node \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --flannel-network 10.128.0.0/21 \
      --flannel-subnet-len 27 \
      --flannel-subnet-min 10.128.0.192 \
      --flannel-subnet-max 10.128.7.224 \
      --flannel-backend vxlan |

      coreseed run-packet \
      --api-key xxx \
      --hostname core-${i} \
      --project-id xxx \
      --plan baremetal_0 \
      --os coreos_alpha \
      --facility ewr1 \
      --billing hourly

    done ;;

esac
```

#### Deploy on Amazon EC2
```bash
#!/bin/bash

case $1 in

  "masters")

    ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')

    for i in 1 2 3; do

      coreseed udata \
      --hostname core-${i} \
      --domain cell-1.dc-1.demo.com \
      --role master \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --etcd-token ${ETCD_TOKEN} |

      gzip --best | coreseed run-ec2 \
      --region eu-west-1 \
      --image-id ami-95bb00e6 \
      --instance-type t2.medium \
      --key-pair xxx \
      --vpc-id vpc-xxx \
      --subnet-id subnet-xxx \
      --sec-group-ids sg-xxx,sg-xxx

    done ;;

  "nodes")

    for i in 4 5 6; do

      coreseed udata \
      --hostname core-${i} \
      --domain cell-1.dc-1.demo.com \
      --role node \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --flannel-network 10.128.0.0/21 \
      --flannel-subnet-len 27 \
      --flannel-subnet-min 10.128.0.192 \
      --flannel-subnet-max 10.128.7.224 \
      --flannel-backend vxlan |

      gzip --best | coreseed run-ec2 \
      --region eu-west-1 \
      --image-id ami-95bb00e6 \
      --instance-type t2.medium \
      --key-pair xxx \
      --vpc-id vpc-xxx \
      --subnet-id subnet-xxx \
      --sec-group-ids sg-xxx,sg-xxx

    done ;;

esac
```

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
