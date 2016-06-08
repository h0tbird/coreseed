package udata

//---------------------------------------------------------------------------
// CoreOS master user data:
//---------------------------------------------------------------------------

const templMaster = `#cloud-config

hostname: "master-{{.HostID}}.{{.Domain}}"

write_files:

 - path: "/etc/hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 master-{{.HostID}}.{{.Domain}} master-{{.HostID}}
    $private_ipv4 master-{{.HostID}}.int.{{.Domain}} master-{{.HostID}}.int

 - path: "/etc/.hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 master-{{.HostID}}.{{.Domain}} master-{{.HostID}}
    $private_ipv4 master-{{.HostID}}.int.{{.Domain}} master-{{.HostID}}.int

 - path: "/etc/resolv.conf"
   content: |
    search {{.Domain}}
    nameserver 8.8.8.8

 - path: "/etc/kato.env"
   content: |
    KATO_CLUSTER_ID={{.ClusterID}}
    KATO_MASTER_COUNT={{.MasterCount}}
    KATO_ROLE={{.Role}}
    KATO_HOST_ID={{.HostID}}
    KATO_ZK={{.ZkServers}}

 {{if .CaCert}}- path: "/etc/ssl/certs/{{.ClusterID}}.pem"
   content: |
    {{.CaCert}}
 {{- end}}

 - path: "/etc/systemd/system/docker.service.d/50-docker-opts.conf"
   content: |
    [Service]
    Environment='DOCKER_OPTS=--registry-mirror=http://external-registry-sys.marathon:5000'

 - path: "/etc/rexray/rexray.env"

 - path: "/etc/rexray/config.yml"
{{- if .RexrayStorageDriver }}
   content: |
    rexray:
      storageDrivers:
      - {{.RexrayStorageDriver}}

    {{.RexrayConfigSnippet}}
{{- end}}

 - path: "/home/core/.bashrc"
   owner: "core:core"
   content: |
    [[ $- != *i* ]] && return
    alias ls='ls -hF --color=auto --group-directories-first'
    alias l='ls -l'
    alias ll='ls -la'
    alias grep='grep --color=auto'
    alias dim='docker images'
    alias dps='docker ps'
    alias drm='docker rm -v $(docker ps -qaf status=exited)'
    alias drmi='docker rmi $(docker images -qf dangling=true)'
    alias drmv='docker volume rm $(docker volume ls -qf dangling=true)'

 - path: "/home/core/.aws/config"
   owner: "core:core"
   permissions: "0644"
   content: |
    [default]
    region = {{.Ec2Region}}

 - path: "/etc/ssh/sshd_config"
   permissions: "0600"
   content: |
    UsePrivilegeSeparation sandbox
    Subsystem sftp internal-sftp
    ClientAliveInterval 180
    UseDNS no
    PermitRootLogin no
    AllowUsers core
    PasswordAuthentication no
    ChallengeResponseAuthentication no

 - path: "/opt/bin/ns1dns"
   permissions: "0755"
   content: |
    #!/bin/bash

    readonly HOST="$(hostname -s)"
    readonly DOMAIN="$(hostname -d)"
    readonly APIURL='https://api.nsone.net/v1'
    readonly APIKEY='{{.Ns1ApiKey}}'
    readonly IP_PUB="$(dig +short myip.opendns.com @resolver1.opendns.com)"
    readonly IP_PRI="$(hostname -i)"
    declare -A IP=(['ext']="${IP_PUB}" ['int']="${IP_PRI}")

    for i in ext int; do

      curl -sX GET -H "X-NSONE-Key: ${APIKEY}" \
      ${APIURL}/zones/${i}.${DOMAIN}/${HOST}.${i}.${DOMAIN}/A | \
      grep -q 'record not found' && METHOD='PUT' || METHOD='POST'

      curl -sX ${METHOD} -H "X-NSONE-Key: ${APIKEY}" \
      ${APIURL}/zones/${i}.${DOMAIN}/${HOST}.${i}.${DOMAIN}/A -d "{
        \"zone\":\"${i}.${DOMAIN}\",
        \"domain\":\"${HOST}.${i}.${DOMAIN}\",
        \"type\":\"A\",
        \"answers\":[{\"answer\":[\"${IP[${i}]}\"]}]}"

    done

 - path: "/opt/bin/etchost"
   permissions: "0755"
   content: |
    #!/bin/bash

    PUSH+=$(echo $(hostname -i) $(hostname -f) $(hostname -s))$'\n'
    PUSH+=$(echo $(hostname -i) $(hostname -s).int.$(hostname -d) $(hostname -s).int)
    etcdctl set /hosts/$(hostname -f) "${PUSH}"

    for i in $(etcdctl ls /hosts 2>/dev/null | grep -v $(hostname -f) | sort); do
      PULL+=$(etcdctl get ${i})$'\n'
    done

    cat /etc/.hosts > /etc/hosts
    echo "${PULL}" >> /etc/hosts

 - path: "/opt/bin/loopssh"
   permissions: "0755"
   content: |
    #!/bin/bash
    A=$(fleetctl list-machines -fields=ip -no-legend)
    for i in $A; do ssh -o UserKnownHostsFile=/dev/null \
    -o StrictHostKeyChecking=no $i -C "$*"; done

 - path: "/opt/bin/awscli"
   permissions: "0755"
   content: |
    #!/bin/bash
    docker run -i --rm \
    --volume /home/core/.aws:/root/.aws:ro \
    --volume ${PWD}:/aws \
    h0tbird/awscli "${@}"

 - path: "/etc/prometheus/prometheus.yml"
   permissions: "0600"
   content: |
    global:
      scrape_interval: 15s
      external_labels:
        monitor: 'codelab-monitor'

    scrape_configs:
      - job_name: 'prometheus'
        scrape_interval: 5s
        target_groups:
          - targets: ['localhost:9191']

 - path: "/etc/fleet/zookeeper.service"
   content: |
    [Unit]
    Description=Zookeeper
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    EnvironmentFile=/etc/kato.env
    ExecStartPre=-/usr/bin/docker kill zookeeper
    ExecStartPre=-/usr/bin/docker rm zookeeper
    ExecStartPre=-/usr/bin/docker pull h0tbird/zookeeper:v3.4.8-2
    ExecStart=/usr/bin/sh -c "docker run \
      --net host \
      --name zookeeper \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --env ZK_SERVER_ID=${KATO_HOST_ID} \
      --env ZK_TICK_TIME=2000 \
      --env ZK_INIT_LIMIT=5 \
      --env ZK_SYNC_LIMIT=2 \
      --env ZK_SERVERS=$${KATO_ZK//:2181/} \
      --env ZK_DATA_DIR=/var/lib/zookeeper \
      --env ZK_CLIENT_PORT=2181 \
      --env ZK_CLIENT_PORT_ADDRESS=$(hostname -i) \
      --env JMXDISABLE=true \
      h0tbird/zookeeper:v3.4.8-2"
    ExecStop=/usr/bin/docker stop -t 5 zookeeper

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=master

 - path: "/etc/fleet/prometheus.service"
   content: |
    [Unit]
    Description=Prometheus Service
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    EnvironmentFile=/etc/kato.env
    ExecStartPre=-/usr/bin/docker kill prometheus
    ExecStartPre=-/usr/bin/docker rm -f prometheus
    ExecStartPre=-/usr/bin/docker pull prom/prometheus:0.19.2
    ExecStartPre=-/usr/bin/docker volume create --name ${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID} -d rexray
    ExecStart=/usr/bin/sh -c "docker run \
      --net host \
      --name prometheus \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --volume /etc/prometheus:/etc/prometheus:ro \
      --volume ${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID}:/prometheus:rw \
      prom/prometheus:0.19.2 \
      -config.file=/etc/prometheus/prometheus.yml \
      -storage.local.path=/prometheus \
      -web.console.libraries=/etc/prometheus/console_libraries \
      -web.console.templates=/etc/prometheus/consoles \
      -web.listen-address=:9191"
    ExecStop=/usr/bin/docker stop -t 5 prometheus

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=master

 - path: "/etc/fleet/mesos-master.service"
   content: |
    [Unit]
    Description=Mesos Master
    After=docker.service zookeeper.service
    Requires=docker.service zookeeper.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    EnvironmentFile=/etc/kato.env
    ExecStartPre=-/usr/bin/docker kill mesos-master
    ExecStartPre=-/usr/bin/docker rm mesos-master
    ExecStartPre=-/usr/bin/docker pull mesosphere/mesos-master:0.28.1
    ExecStart=/usr/bin/sh -c "docker run \
      --privileged \
      --name mesos-master \
      --net host \
      --volume /var/lib/mesos:/var/lib/mesos:rw \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      mesosphere/mesos-master:0.28.1 \
      --ip=$(hostname -i) \
      --zk=zk://${KATO_ZK}/mesos \
      --work_dir=/var/lib/mesos/master \
      --log_dir=/var/log/mesos \
      --quorum=$(($KATO_MASTER_COUNT/2 + 1))"
    ExecStop=/usr/bin/docker stop -t 5 mesos-master

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=master

 - path: "/etc/fleet/mesos-node.service"
   content: |
    [Unit]
    Description=Mesos Node
    After=docker.service dnsmasq.service
    Wants=dnsmasq.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    EnvironmentFile=/etc/kato.env
    ExecStartPre=-/usr/bin/docker kill mesos-node
    ExecStartPre=-/usr/bin/docker rm mesos-node
    ExecStartPre=-/usr/bin/docker pull mesosphere/mesos-slave:0.28.1
    ExecStart=/usr/bin/sh -c "docker run \
      --privileged \
      --net host \
      --pid host \
      --name mesos-node \
      --volume /sys:/sys \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --volume /usr/bin/docker:/usr/bin/docker:ro \
      --volume /var/run/docker.sock:/var/run/docker.sock:rw \
      --volume /lib64/libdevmapper.so.1.02:/lib/libdevmapper.so.1.02:ro \
      --volume /lib64/libsystemd.so.0:/lib/libsystemd.so.0:ro \
      --volume /lib64/libgcrypt.so.20:/lib/libgcrypt.so.20:ro \
      --volume /var/lib/mesos:/var/lib/mesos:rw \
      --volume /etc/certs:/etc/certs:ro \
      mesosphere/mesos-slave:0.28.1 \
      --ip=$(hostname -i) \
      --containerizers=docker \
      --executor_registration_timeout=2mins \
      --master=zk://${KATO_ZK}/mesos \
      --work_dir=/var/lib/mesos/node \
      --log_dir=/var/log/mesos/node"
    ExecStop=/usr/bin/docker stop -t 5 mesos-node

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=worker

 - path: "/etc/fleet/mesos-dns.service"
   content: |
    [Unit]
    Description=Mesos DNS
    After=docker.service zookeeper.service mesos-master.service
    Requires=docker.service zookeeper.service mesos-master.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    EnvironmentFile=/etc/kato.env
    ExecStartPre=-/usr/bin/docker kill mesos-dns
    ExecStartPre=-/usr/bin/docker rm mesos-dns
    ExecStartPre=-/usr/bin/docker pull h0tbird/mesos-dns:v0.5.2-1
    ExecStart=/usr/bin/sh -c "docker run \
      --name mesos-dns \
      --net host \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --env MDNS_ZK=zk://${KATO_ZK}/mesos \
      --env MDNS_REFRESHSECONDS=45 \
      --env MDNS_LISTENER=$(hostname -i) \
      --env MDNS_HTTPON=false \
      --env MDNS_TTL=45 \
      --env MDNS_RESOLVERS=8.8.8.8 \
      --env MDNS_DOMAIN=$(hostname -d | cut -d. -f-2).mesos \
      --env MDNS_IPSOURCE=netinfo \
      h0tbird/mesos-dns:v0.5.2-1"
    ExecStartPost=/usr/bin/sh -c ' \
      echo search $(hostname -d | cut -d. -f-2).mesos $(hostname -d) > /etc/resolv.conf && \
      echo "nameserver $(hostname -i)" >> /etc/resolv.conf'
    ExecStop=/usr/bin/sh -c ' \
      echo search $(hostname -d) > /etc/resolv.conf && \
      echo "nameserver 8.8.8.8" >> /etc/resolv.conf'
    ExecStop=/usr/bin/docker stop -t 5 mesos-dns

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=master

 - path: "/etc/fleet/marathon.service"
   content: |
    [Unit]
    Description=Marathon
    After=docker.service zookeeper.service mesos-master.service
    Requires=docker.service zookeeper.service mesos-master.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    EnvironmentFile=/etc/kato.env
    ExecStartPre=-/usr/bin/docker kill marathon
    ExecStartPre=-/usr/bin/docker rm marathon
    ExecStartPre=-/usr/bin/docker pull mesosphere/marathon:v1.1.1
    ExecStart=/usr/bin/sh -c "docker run \
      --name marathon \
      --net host \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --env LIBPROCESS_IP=$(hostname -i) \
      --env LIBPROCESS_PORT=9090 \
      mesosphere/marathon:v1.1.1 \
      --http_address $(hostname -i) \
      --master zk://${KATO_ZK}/mesos \
      --zk zk://${KATO_ZK}/marathon \
      --task_launch_timeout 240000 \
      --checkpoint"
    ExecStop=/usr/bin/docker stop -t 5 marathon

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=master

 - path: "/etc/fleet/marathon-lb.service"
   content: |
    [Unit]
    Description=marathon-lb
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill marathon-lb
    ExecStartPre=-/usr/bin/docker rm marathon-lb
    ExecStartPre=-/usr/bin/docker pull mesosphere/marathon-lb:v1.2.2
    ExecStart=/usr/bin/sh -c "docker run \
      --name marathon-lb \
      --net host \
      --privileged \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --env PORTS=9090,9091 \
      mesosphere/marathon-lb:v1.2.2 sse \
      --marathon http://marathon:8080 \
      --health-check \
      --group external \
      --group internal"
    ExecStop=/usr/bin/docker stop -t 5 marathon-lb

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=worker

 - path: "/etc/fleet/cadvisor.service"
   content: |
    [Unit]
    Description=cAdvisor Service
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill cadvisor
    ExecStartPre=-/usr/bin/docker rm -f cadvisor
    ExecStartPre=-/usr/bin/docker pull google/cadvisor:v0.22.0
    ExecStart=/usr/bin/sh -c "docker run \
      --net host \
      --name cadvisor \
      --volume /:/rootfs:ro \
      --volume /var/run:/var/run:rw \
      --volume /sys:/sys:ro \
      --volume /var/lib/docker/:/var/lib/docker:ro \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      google/cadvisor:v0.22.0 \
      --listen_ip $(hostname -i) \
      --logtostderr \
      --port=4194"
    ExecStop=/usr/bin/docker stop -t 5 cadvisor

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true

 - path: "/etc/fleet/dnsmasq.service"
   content: |
    [Unit]
    Description=Lightweight caching DNS proxy
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill dnsmasq
    ExecStartPre=-/usr/bin/docker rm -f dnsmasq
    ExecStartPre=-/usr/bin/docker pull janeczku/go-dnsmasq:release-1.0.5
    ExecStartPre=/usr/bin/sh -c " \
      etcdctl member list 2>1 | awk -F [/:] '{print $9}' | tr '\n' ',' > /tmp/ns && \
      awk '/^nameserver/ {print $2; exit}' /run/systemd/resolve/resolv.conf >> /tmp/ns"
    ExecStart=/usr/bin/sh -c "docker run \
      --name dnsmasq \
      --net host \
      --volume /etc/resolv.conf:/etc/resolv.conf:rw \
      --volume /etc/hosts:/etc/hosts:ro \
      janeczku/go-dnsmasq:release-1.0.5 \
      --listen $(hostname -i) \
      --nameservers $(cat /tmp/ns) \
      --hostsfile /etc/hosts \
      --hostsfile-poll 60 \
      --default-resolver \
      --search-domains $(hostname -d | cut -d. -f-2).mesos,$(hostname -d) \
      --append-search-domains"
    ExecStop=/usr/bin/docker stop -t 5 dnsmasq

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=worker

 - path: "/etc/fleet/mongodb.service"
   content: |
    [Unit]
    Description=MongoDB
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    EnvironmentFile=/etc/kato.env
    ExecStartPre=-/usr/bin/docker kill mongodb
    ExecStartPre=-/usr/bin/docker rm mongodb
    ExecStartPre=-/usr/bin/docker pull mongo:3.2
    ExecStartPre=-/usr/bin/docker volume create --name ${KATO_CLUSTER_ID}-pritunl-mongo -d rexray
    ExecStart=/usr/bin/sh -c "docker run \
      --name mongodb \
      --net host \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --volume ${KATO_CLUSTER_ID}-pritunl-mongo:/data/db:rw \
      mongo:3.2 \
      --bind_ip 127.0.0.1"
    ExecStop=/usr/bin/docker stop -t 5 mongodb

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=edge

 - path: "/etc/fleet/pritunl.service"
   content: |
    [Unit]
    Description=Pritunl
    After=docker.service mongodb.service
    Requires=docker.service mongodb.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill pritunl
    ExecStartPre=-/usr/bin/docker rm pritunl
    ExecStartPre=-/usr/bin/docker pull h0tbird/pritunl:v1.21.954.48-3
    ExecStart=/usr/bin/sh -c "docker run \
      --privileged \
      --name pritunl \
      --net host \
      --volume /etc/resolv.conf:/etc/resolv.conf:ro \
      --volume /etc/hosts:/etc/hosts:ro \
      --env MONGODB_URI=mongodb://127.0.0.1:27017/pritunl \
      h0tbird/pritunl:v1.21.954.48-3"
    ExecStop=/usr/bin/docker stop -t 5 pritunl

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=edge

coreos:

 units:

  - name: "etcd2.service"
    command: "start"

  - name: "fleet.service"
    command: "start"

  - name: flanneld.service
    command: "start"
    drop-ins:
     - name: 50-network-config.conf
       content: |
        [Service]
        ExecStartPre=/usr/bin/etcdctl set /coreos.com/network/config '{ "Network": "{{.FlannelNetwork}}","SubnetLen":{{.FlannelSubnetLen}} ,"SubnetMin": "{{.FlannelSubnetMin}}","SubnetMax": "{{.FlannelSubnetMax}}","Backend": {"Type": "{{.FlannelBackend}}"} }'

  - name: "ns1dns.service"
    command: "start"
    content: |
     [Unit]
     Description=Publish DNS records to nsone
     Before=etcd2.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/ns1dns

  - name: "etchost.service"
    command: "start"
    content: |
     [Unit]
     Description=Stores IP and hostname in etcd
     Requires=etcd2.service
     After=etcd2.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/etchost

  - name: "etchost.timer"
    command: "start"
    content: |
     [Unit]
     Description=Run etchost.service every 5 minutes

     [Timer]
     OnBootSec=2min
     OnUnitActiveSec=5min

  - name: "rexray.service"
    command: "start"
    content: |
     [Unit]
     Description=REX-Ray volume plugin
     Before=docker.service

     [Service]
     EnvironmentFile=/etc/rexray/rexray.env
     ExecStartPre=-/bin/bash -c '\
       REXRAY_URL=https://dl.bintray.com/emccode/rexray/stable/latest/rexray-Linux-x86_64.tar.gz; \
       [ -f /opt/bin/rexray ] || { curl -sL $${REXRAY_URL} | tar -xz -C /opt/bin; }; \
       [ -x /opt/bin/rexray ] || { chmod +x /opt/bin/rexray; }'
     ExecStart=/opt/bin/rexray start -f
     ExecReload=/bin/kill -HUP $MAINPID
     KillMode=process

     [Install]
     WantedBy=docker.service

 flannel:
  interface: $private_ipv4

 fleet:
  public-ip: "$private_ipv4"
  metadata: "role=master,id={{.HostID}}"

 etcd2:
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} name: "master-{{.HostID}}"
  initial-cluster: "master-1=http://master-1:2380,master-2=http://master-2:2380,master-3=http://master-3:2380"
  initial-cluster-state: "new"{{end}}
  advertise-client-urls: "http://$private_ipv4:2379"
  initial-advertise-peer-urls: "http://$private_ipv4:2380"
  listen-client-urls: "http://127.0.0.1:2379,http://$private_ipv4:2379"
  listen-peer-urls: "http://$private_ipv4:2380"
`
