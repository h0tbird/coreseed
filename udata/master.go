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
    source /etc/kato.env
    PUSH+=$(echo $(hostname -i) $(hostname -f) $(hostname -s))$'\n'
    PUSH+=$(echo $(hostname -i) $(hostname -s).int.$(hostname -d) $(hostname -s).int)
    etcdctl set /hosts/${KATO_ROLE}/$(hostname -f) "${PUSH}"
    KEYS=$(etcdctl ls --recursive /hosts | grep $(hostname -d) | grep -v $(hostname -f) | sort)
    for i in $KEYS; do PULL+=$(etcdctl get ${i})$'\n'; done
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

 - path: "/etc/prometheus/targets/prometheus.yml"

 - path: "/etc/prometheus/prometheus.yml"
   permissions: "0600"
   content: |
    global:
     scrape_interval: 1m
     scrape_timeout: 10s
     evaluation_interval: 10s

    rule_files:
     - /etc/prometheus/prometheus.rules

    scrape_configs:

     - job_name: 'prometheus'
       scrape_interval: 10s
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/prometheus.yml

     - job_name: 'cadvisor'
       scrape_interval: 10s
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/cadvisor.yml

     - job_name: 'etcd'
       scrape_interval: 10s
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/etcd.yml

     - job_name: 'node'
       scrape_interval: 10s
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/node.yml

     - job_name: 'mesos'
       scrape_interval: 10s
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/mesos.yml

     - job_name: 'haproxy'
       scrape_interval: 10s
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/haproxy.yml

     - job_name: 'zookeeper'
       scrape_interval: 10s
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/zookeeper.yml

 - path: "/etc/confd/conf.d/prom-prometheus.toml"
   content: |
    [template]
    src = "prom-prometheus.tmpl"
    dest = "/etc/prometheus/targets/prometheus.yml"
    keys = [ "/hosts/master" ]

 - path: "/etc/confd/templates/prom-prometheus.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:9191{{"{{"}}end{{"}}"}}
      labels:
        role: master

 - path: "/etc/confd/conf.d/prom-cadvisor.toml"
   content: |
    [template]
    src = "prom-cadvisor.tmpl"
    dest = "/etc/prometheus/targets/cadvisor.yml"
    keys = [
      "/hosts/master",
      "/hosts/worker",
    ]

 - path: "/etc/confd/templates/prom-cadvisor.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:4194{{"{{"}}end{{"}}"}}
      labels:
        role: master
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:4194{{"{{"}}end{{"}}"}}
      labels:
        role: worker

 - path: "/etc/confd/conf.d/prom-etcd.toml"
   content: |
    [template]
    src = "prom-etcd.tmpl"
    dest = "/etc/prometheus/targets/etcd.yml"
    keys = [
      "/hosts/master",
      "/hosts/worker",
    ]

 - path: "/etc/confd/templates/prom-etcd.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:2379{{"{{"}}end{{"}}"}}
      labels:
        role: master
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:2379{{"{{"}}end{{"}}"}}
      labels:
        role: worker

 - path: "/etc/confd/conf.d/prom-node.toml"
   content: |
    [template]
    src = "prom-node.tmpl"
    dest = "/etc/prometheus/targets/node.yml"
    keys = [
      "/hosts/master",
      "/hosts/worker",
    ]

 - path: "/etc/confd/templates/prom-node.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:9101{{"{{"}}end{{"}}"}}
      labels:
        role: master
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:9101{{"{{"}}end{{"}}"}}
      labels:
        role: worker

 - path: "/etc/confd/conf.d/prom-mesos.toml"
   content: |
    [template]
    src = "prom-mesos.tmpl"
    dest = "/etc/prometheus/targets/mesos.yml"
    keys = [
      "/hosts/master",
      "/hosts/worker",
    ]

 - path: "/etc/confd/templates/prom-mesos.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:9104{{"{{"}}end{{"}}"}}
      labels:
        role: master
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:9104{{"{{"}}end{{"}}"}}
      labels:
        role: worker

 - path: "/etc/confd/conf.d/prom-haproxy.toml"
   content: |
    [template]
    src = "prom-haproxy.tmpl"
    dest = "/etc/prometheus/targets/haproxy.yml"
    keys = [ "/hosts/worker" ]

 - path: "/etc/confd/templates/prom-haproxy.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:9102{{"{{"}}end{{"}}"}}
      labels:
        role: worker

 - path: "/etc/confd/conf.d/prom-zookeeper.toml"
   content: |
    [template]
    src = "prom-zookeeper.tmpl"
    dest = "/etc/prometheus/targets/zookeeper.yml"
    keys = [ "/hosts/master" ]

 - path: "/etc/confd/templates/prom-zookeeper.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      - {{"{{"}}base .Key{{"}}"}}:9103{{"{{"}}end{{"}}"}}
      labels:
        role: master

coreos:

 units:

  - name: "etcd2.service"
    command: "start"

  - name: "fleet.service"
    command: "start"

  - name: "flanneld.service"
    command: "start"
    drop-ins:
     - name: 50-network-config.conf
       content: |
        [Service]
        ExecStartPre=/usr/bin/etcdctl set /coreos.com/network/config '{ "Network": "{{.FlannelNetwork}}","SubnetLen":{{.FlannelSubnetLen}} ,"SubnetMin": "{{.FlannelSubnetMin}}","SubnetMax": "{{.FlannelSubnetMax}}","Backend": {"Type": "{{.FlannelBackend}}"} }'

  - name: "format-ephemeral.service"
    command: "start"
    content: |
     [Unit]
     Description=Formats the ephemeral drive
     After=dev-xvdb.device
     Requires=dev-xvdb.device

     [Service]
     Type=oneshot
     RemainAfterExit=yes
     ExecStart=/usr/sbin/wipefs -f /dev/xvdb
     ExecStart=/usr/sbin/mkfs.ext4 -F /dev/xvdb

  - name: "var-lib-docker.mount"
    command: "start"
    content: |
     [Unit]
     Description=Mount ephemeral to /var/lib/docker
     After=format-ephemeral.service
     Requires=format-ephemeral.service

     [Mount]
     What=/dev/xvdb
     Where=/var/lib/docker
     Type=ext4

  - name: "docker.service"
    drop-ins:
     - name: "10-wait-docker.conf"
       content: |
        [Unit]
        After=var-lib-docker.mount
        Requires=var-lib-docker.mount

     - name: "20-docker-opts.conf"
       content: |
        [Service]
        Environment='DOCKER_OPTS=--registry-mirror=http://external-registry-sys.marathon:5000'

  - name: "zookeeper.service"
    command: "start"
    content: |
     [Unit]
     Description=Zookeeper
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull h0tbird/zookeeper:v3.4.8-2
     ExecStart=/usr/bin/sh -c 'docker run \
       --net host \
       --name %p \
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
       h0tbird/zookeeper:v3.4.8-2'
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "mesos-master.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos Master
     After=docker.service zookeeper.service
     Requires=docker.service zookeeper.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull mesosphere/mesos-master:0.28.1
     ExecStartPre=/usr/bin/echo ruok | ncat $(hostname -i) 2181 | grep -q imok
     ExecStart=/usr/bin/sh -c "docker run \
       --privileged \
       --name %p \
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
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "mesos-dns.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos DNS
     After=docker.service zookeeper.service mesos-master.service
     Requires=docker.service zookeeper.service mesos-master.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull h0tbird/mesos-dns:v0.5.2-1
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
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
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "marathon.service"
    command: "start"
    content: |
     [Unit]
     Description=Marathon
     After=docker.service zookeeper.service mesos-master.service
     Requires=docker.service zookeeper.service mesos-master.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull mesosphere/marathon:v1.1.1
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
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
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "confd.service"
    command: "start"
    content: |
     [Unit]
     Description=Lightweight configuration management tool
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull katosys/confd:v0.11.0-2
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       --volume /etc:/etc:rw \
       katosys/confd:v0.11.0-2 \
       -node 127.0.0.1:2379 \
       -watch"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "rexray.service"
    command: "start"
    content: |
     [Unit]
     Description=REX-Ray volume plugin
     Before=docker.service

     [Service]
     EnvironmentFile=/etc/rexray/rexray.env
     ExecStartPre=-/bin/bash -c '\
       REXRAY_URL=https://dl.bintray.com/emccode/rexray/stable/0.3.3/rexray-Linux-x86_64-0.3.3.tar.gz; \
       [ -f /opt/bin/rexray ] || { curl -sL $${REXRAY_URL} | tar -xz -C /opt/bin; }; \
       [ -x /opt/bin/rexray ] || { chmod +x /opt/bin/rexray; }'
     ExecStart=/opt/bin/rexray start -f
     ExecReload=/bin/kill -HUP $MAINPID
     KillMode=process

     [Install]
     WantedBy=docker.service

  - name: "prometheus.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus service
     After=docker.service rexray.service confd.service
     Requires=docker.service rexray.service
     Wants=confd.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull prom/prometheus:0.20.0
     ExecStartPre=-/usr/bin/docker volume create --name ${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID} -d rexray
     ExecStart=/usr/bin/sh -c "docker run \
       --net host \
       --name %p \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --volume /etc/prometheus:/etc/prometheus:ro \
       --volume ${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID}:/prometheus:rw \
       prom/prometheus:0.20.0 \
       -config.file=/etc/prometheus/prometheus.yml \
       -storage.local.path=/prometheus \
       -web.console.libraries=/etc/prometheus/console_libraries \
       -web.console.templates=/etc/prometheus/consoles \
       -web.listen-address=:9191"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "cadvisor.service"
    command: "start"
    content: |
     [Unit]
     Description=cAdvisor service
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull google/cadvisor:v0.23.2
     ExecStart=/usr/bin/sh -c "docker run \
       --net host \
       --name %p \
       --volume /:/rootfs:ro \
       --volume /var/run:/var/run:rw \
       --volume /sys:/sys:ro \
       --volume /var/lib/docker/:/var/lib/docker:ro \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       google/cadvisor:v0.23.2 \
       --listen_ip $(hostname -i) \
       --logtostderr \
       --port=4194"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "update-ca-certificates.service"
    drop-ins:
     - name: 50-rehash-certs.conf
       content: |
        [Unit]
        ConditionPathIsSymbolicLink=

        [Service]
        ExecStart=
        ExecStart=/usr/sbin/update-ca-certificates

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

  - name: "mesos-master-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus mesos master exporter
     After=docker.service mesos-master.service
     Requires=docker.service mesos-master.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull katosys/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       katosys/exporters:v0.1.0-1 mesos_exporter \
       -master http://$(hostname):5050 \
       -addr :9104"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "node-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus node exporter
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull katosys/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       katosys/exporters:v0.1.0-1 node_exporter \
       -web.listen-address :9101"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

  - name: "zookeeper-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus zookeeper exporter
     After=docker.service zookeeper.service
     Requires=docker.service zookeeper.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull katosys/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       katosys/exporters:v0.1.0-1 zookeeper_exporter \
       -web.listen-address :9103 \
       $(echo ${KATO_ZK} | tr , ' ')"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target

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
