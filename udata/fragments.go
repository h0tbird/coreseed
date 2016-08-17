package udata

import "os"

//---------------------------------------------------------------------------
// func: loadFragments
//---------------------------------------------------------------------------

func (d *Data) loadFragments() {

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `#cloud-config
hostname: "{{.HostName}}-{{.HostID}}.{{.Domain}}"
write_files:`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/etc/hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 {{.HostName}}-{{.HostID}}.{{.Domain}} {{range .Aliases}}{{.}}-{{$.HostID}} {{end}}marathon-lb`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/etc/.hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 {{.HostName}}-{{.HostID}}.{{.Domain}} {{range .Aliases}}{{.}}-{{$.HostID}} {{end}}marathon-lb`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/etc/resolv.conf"
   content: |
    search {{.Domain}}
    nameserver 8.8.8.8`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/etc/kato.env"
   content: |
    KATO_CLUSTER_ID={{.ClusterID}}
    KATO_QUORUM_COUNT={{.QuorumCount}}
    KATO_MASTER_COUNT={{.MasterCount}}
    KATO_ROLES='{{range .Roles}}{{.}} {{end}}'
    KATO_HOST_NAME={{.HostName}}
    KATO_HOST_ID={{.HostID}}
    KATO_ZK={{.ZkServers}}
    KATO_SYSTEMD_UNITS='{{range .SystemdUnits}}{{.}} {{end}}'`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
{{if .CaCert}} - path: "/etc/ssl/certs/{{.ClusterID}}.pem"
   content: |
    {{.CaCert}}
{{end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"vbox"},
		},
		data: `
 - path: "/etc/rexray/rexray.env"
 - path: "/etc/rexray/config.yml"
{{- if .RexrayStorageDriver }}
   content: |
    rexray:
     storageDrivers:
     - {{.RexrayStorageDriver}}
    virtualbox:
     endpoint: http://` + d.RexrayEndpointIP + `:18083
     volumePath: ` + os.Getenv("HOME") + `/VirtualBox Volumes
     controllerName: SATA
{{- end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"ec2"},
		},
		data: `
 - path: "/etc/rexray/rexray.env"
 - path: "/etc/rexray/config.yml"
{{- if .RexrayStorageDriver }}
   content: |
    rexray:
     storageDrivers:
     - {{.RexrayStorageDriver}}
    aws:
     rexrayTag: kato
{{- end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/home/core/.bashrc"
   owner: "core:core"
   content: |
    [[ $- = *i* ]] && {
      alias ls='ls -hF --color=auto --group-directories-first'
      alias grep='grep --color=auto'
    } || shopt -s expand_aliases
    alias l='ls -l'
    alias ll='ls -la'
    alias dim='docker images'
    alias dps='docker ps'
    alias drm='docker rm -v $(docker ps -qaf status=exited)'
    alias drmi='docker rmi $(docker images -qf dangling=true)'
    alias drmv='docker volume rm $(docker volume ls -qf dangling=true)'
    export PATH='/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin'`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/home/core/.aws/config"
   owner: "core:core"
   permissions: "0644"
   content: |
    [default]
    region = {{.Ec2Region}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
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
    ChallengeResponseAuthentication no`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/opt/bin/ns1dns"
   permissions: "0755"
   content: |
    #!/bin/bash

    source /etc/kato.env
    readonly DOMAIN="$(hostname -d)"
    readonly APIURL='https://api.nsone.net/v1'
    readonly APIKEY='{{.Ns1ApiKey}}'
    readonly IP_PUB="$(dig +short myip.opendns.com @resolver1.opendns.com)"
    readonly IP_PRI="$(hostname -i)"
    declare -A IP=(['ext']="${IP_PUB}" ['int']="${IP_PRI}")

    for ROLE in ${KATO_ROLES}; do
      for i in ext int; do
        curl -sX GET -H "X-NSONE-Key: ${APIKEY}" \
        ${APIURL}/zones/${i}.${DOMAIN}/${ROLE}-${KATO_HOST_ID}.${i}.${DOMAIN}/A | \
        grep -q 'record not found' && METHOD='PUT' || METHOD='POST'

        curl -sX ${METHOD} -H "X-NSONE-Key: ${APIKEY}" \
        ${APIURL}/zones/${i}.${DOMAIN}/${ROLE}-${KATO_HOST_ID}.${i}.${DOMAIN}/A -d "{
          \"zone\":\"${i}.${DOMAIN}\",
          \"domain\":\"${ROLE}-${KATO_HOST_ID}.${i}.${DOMAIN}\",
          \"type\":\"A\",
          \"answers\":[{\"answer\":[\"${IP[${i}]}\"]}]}"
      done
    done`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/opt/bin/etchost"
   permissions: "0755"
   content: |
    #!/bin/bash
    source /etc/kato.env
    PUSH=$(sed 's/ marathon-lb//' /etc/.hosts | grep -v localhost)
    for i in $(echo ${KATO_ROLES}); do
    etcdctl set /hosts/${i}/$(hostname -f) "${PUSH}"; done
    KEYS=$(etcdctl ls --recursive /hosts | grep $(hostname -d) | \
    grep -v $(hostname -f) | rev | sort | rev | uniq -s 14 | sort)
    for i in $KEYS; do PULL+=$(etcdctl get ${i})$'\n'; done
    cat /etc/.hosts > /etc/hosts
    echo "${PULL}" >> /etc/hosts`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/opt/bin/loopssh"
   permissions: "0755"
   content: |
    #!/bin/bash
    G=$(tput setaf 2); N=$(tput sgr0)
    A=$(grep $1 /etc/hosts | awk '{print $2}' | sort -u | grep -v int)
    for i in $A; do echo "${G}--[ $i ]--${N}"; ssh -o UserKnownHostsFile=/dev/null \
    -o StrictHostKeyChecking=no $i -C "${@:2}" 2> /dev/null; done`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/opt/bin/awscli"
   permissions: "0755"
   content: |
    #!/bin/bash
    docker run -i --rm \
    --volume /home/core/.aws:/root/.aws:ro \
    --volume ${PWD}:/aws \
    katosys/awscli:v1.10.47-1 "${@}"`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/opt/bin/katostat"
   permissions: "0755"
   content: |
    #!/bin/bash
    source /etc/kato.env
    systemctl -p Id,LoadState,ActiveState,SubState show ${KATO_SYSTEMD_UNITS} | \
    awk 'BEGIN {RS="\n\n"; FS="\n";} {print $2"\t"$3"\t"$4"\t"$1}'`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
{{if .CaCert}} - path: "/opt/bin/getcerts"
   permissions: "0755"
   content: |
    #!/bin/bash
    [ -d /etc/certs ] || mkdir /etc/certs && cd /etc/certs
    [ -f certs.tar.bz2 ] || /opt/bin/awscli s3 cp s3://{{.Domain}}/certs.tar.bz2 .
{{end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
{{if .CaCert}} - path: "/opt/bin/custom-ca"
   permissions: "0755"
   content: |
    #!/bin/bash
    source /etc/kato.env
    [ -f /etc/ssl/certs/${KATO_CLUSTER_ID}.pem ] && {
      ID=$(sed -n 2p /etc/ssl/certs/${KATO_CLUSTER_ID}.pem)
      NU=$(grep -lir $ID /etc/ssl/certs/* | wc -l)
      [ "$NU" -lt "2" ] && update-ca-certificates &> /dev/null
    }
{{end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
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
          - /etc/prometheus/targets/zookeeper.yml`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
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
      - {{"{{"}}base .Key{{"}}"}}:9105{{"{{"}}end{{"}}"}}
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
        role: master`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
coreos:
 units:
  - name: "etcd2.service"
    command: "start"`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master", "worker", "border"},
		},
		data: `
  - name: "flanneld.service"
    command: "start"
    drop-ins:
     - name: 50-network-config.conf
       content: |
        [Service]
        ExecStartPre=/usr/bin/etcdctl set /coreos.com/network/config '{ "Network": "{{.FlannelNetwork}}","SubnetLen":{{.FlannelSubnetLen}} ,"SubnetMin": "{{.FlannelSubnetMin}}","SubnetMax": "{{.FlannelSubnetMax}}","Backend": {"Type": "{{.FlannelBackend}}"} }'`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
{{if .CaCert}}  - name: "custom-ca.service"
    command: "start"
    content: |
     [Unit]
     Description=Re-hash SSL certificates
     Before=docker.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/custom-ca
{{end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"ec2"},
		},
		data: `
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
     ExecStart=/usr/sbin/mkfs.ext4 -F /dev/xvdb`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"ec2"},
		},
		data: `
  - name: "var-lib-docker.mount"
    command: "start"
    content: |
     [Unit]
     Description=Mount ephemeral to /var/lib/docker
     Requires=format-ephemeral.service
     After=format-ephemeral.service

     [Mount]
     What=/dev/xvdb
     Where=/var/lib/docker
     Type=ext4`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"ec2"},
		},
		data: `
  - name: "docker.service"
    drop-ins:
     - name: "10-var-lib-docker.conf"
       content: |
        [Unit]
        After=var-lib-docker.mount
        Requires=var-lib-docker.mount
     - name: "20-docker-opts.conf"
       content: |
        [Unit]
        After=flanneld.service
        [Service]
        Environment='DOCKER_OPTS=--registry-mirror=http://external-registry-sys.marathon:5000'`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf:  []string{"quorum", "master", "worker", "border"},
			noneOf: []string{"ec2"},
		},
		data: `
  - name: "docker.service"
    drop-ins:
     - name: "20-docker-opts.conf"
       content: |
        [Service]
        [Unit]
        After=flanneld.service
        Environment='DOCKER_OPTS=--registry-mirror=http://external-registry-sys.marathon:5000'`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
		},
		data: `
  - name: "zookeeper.service"
    command: "start"
    content: |
     [Unit]
     Description=Zookeeper
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"quorum", "master"},
		},
		data: `
  - name: "mesos-master.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos Master
     After=docker.service zookeeper.service
     Requires=docker.service zookeeper.service

     [Service]
     Restart=always
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
       --quorum=$(($KATO_QUORUM_COUNT/2 + 1))"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf:  []string{"master"},
			noneOf: []string{"quorum"},
		},
		data: `
  - name: "mesos-master.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos Master
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull mesosphere/mesos-master:0.28.1
     ExecStartPre=/usr/bin/echo ruok | ncat quorum-1 2181 | grep -q imok
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
       --quorum=$(($KATO_QUORUM_COUNT/2 + 1))"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf:  []string{"master"},
			noneOf: []string{"worker"},
		},
		data: `
  - name: "mesos-dns.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos DNS
     After=docker.service mesos-master.service
     Requires=docker.service mesos-master.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"worker"},
		},
		data: `
  - name: "mesos-dns.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos DNS
     After=docker.service mesos-master.service go-dnsmasq.service
     Requires=docker.service mesos-master.service go-dnsmasq.service

     [Service]
     Restart=always
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
       --env MDNS_PORT=54 \
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "marathon.service"
    command: "start"
    content: |
     [Unit]
     Description=Marathon
     After=docker.service mesos-master.service
     Requires=docker.service mesos-master.service

     [Service]
     Restart=always
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
       --hostname master-${KATO_HOST_ID}.$(hostname -d) \
       --checkpoint"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "confd.service"
    command: "start"
    content: |
     [Unit]
     Description=Lightweight configuration management tool
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "rexray.service"
    command: "start"
    content: |
     [Unit]
     Description=REX-Ray volume plugin
     Before=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/rexray/rexray.env
     ExecStartPre=-/bin/bash -c '\
       REXRAY_URL=https://dl.bintray.com/emccode/rexray/stable/0.3.3/rexray-Linux-x86_64-0.3.3.tar.gz; \
       [ -f /opt/bin/rexray ] || { curl -sL $${REXRAY_URL} | tar -xz -C /opt/bin; }; \
       [ -x /opt/bin/rexray ] || { chmod +x /opt/bin/rexray; }'
     ExecStart=/opt/bin/rexray start -f
     ExecReload=/bin/kill -HUP $MAINPID
     KillMode=process

     [Install]
     WantedBy=docker.service`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "prometheus.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus service
     After=docker.service rexray.service confd.service
     Requires=docker.service rexray.service
     Wants=confd.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "cadvisor.service"
    command: "start"
    content: |
     [Unit]
     Description=cAdvisor service
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "ns1dns.service"
    command: "start"
    content: |
     [Unit]
     Description=Publish DNS records to nsone
     Before=etcd2.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/ns1dns`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "etchost.service"
    command: "start"
    content: |
     [Unit]
     Description=Stores IP and hostname in etcd
     Requires=etcd2.service
     After=etcd2.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/etchost`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "etchost.timer"
    command: "start"
    content: |
     [Unit]
     Description=Run etchost.service every 5 minutes

     [Timer]
     OnBootSec=2min
     OnUnitActiveSec=5min`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "mesos-master-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus mesos master exporter
     After=docker.service mesos-master.service
     Requires=docker.service mesos-master.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "node-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus node exporter
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
		},
		data: `
  - name: "zookeeper-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus zookeeper exporter
     After=docker.service zookeeper.service
     Requires=docker.service zookeeper.service

     [Service]
     Restart=always
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
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"border"},
		},
		data: `
  - name: "mongodb.service"
    command: "start"
    content: |
     [Unit]
     Description=MongoDB
     After=docker.service rexray.service
     Requires=docker.service rexray.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull mongo:3.2
     ExecStartPre=-/usr/bin/docker volume create --name ${KATO_CLUSTER_ID}-pritunl-mongo -d rexray
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --volume ${KATO_CLUSTER_ID}-pritunl-mongo:/data/db:rw \
       mongo:3.2 \
       --bind_ip 127.0.0.1"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"border"},
		},
		data: `
  - name: "pritunl.service"
    command: "start"
    content: |
     [Unit]
     Description=Pritunl
     After=docker.service mongodb.service
     Requires=docker.service mongodb.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull h0tbird/pritunl:v1.21.954.48-3
     ExecStart=/usr/bin/sh -c "docker run \
       --privileged \
       --name %p \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --env MONGODB_URI=mongodb://127.0.0.1:27017/pritunl \
       h0tbird/pritunl:v1.21.954.48-3"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf:  []string{"worker"},
			noneOf: []string{"master"},
		},
		data: `
  - name: "go-dnsmasq.service"
    command: "start"
    content: |
     [Unit]
     Description=Lightweight caching DNS proxy
     After=docker.service ns1dns.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull janeczku/go-dnsmasq:release-1.0.6
     ExecStartPre=/usr/bin/sh -c " \
       for i in $(seq ${KATO_MASTER_COUNT}); do \
       dig @dns1.p01.nsone.net +short master-${i}.$(hostname -d); done \
       | tr '\n' ',' > /tmp/ns && echo 8.8.8.8 >> /tmp/ns"
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:rw \
       --volume /etc/hosts:/etc/hosts:ro \
       janeczku/go-dnsmasq:release-1.0.6 \
       --listen $(hostname -i) \
       --nameservers $(cat /tmp/ns) \
       --hostsfile /etc/hosts \
       --hostsfile-poll 60 \
       --default-resolver \
       --search-domains $(hostname -d | cut -d. -f-2).mesos,$(hostname -d) \
       --append-search-domains"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
			allOf: []string{"master"},
		},
		data: `
  - name: "go-dnsmasq.service"
    command: "start"
    content: |
     [Unit]
     Description=Lightweight caching DNS proxy
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull janeczku/go-dnsmasq:release-1.0.6
     ExecStartPre=/usr/bin/sh -c " \
       etcdctl member list 2>1 | awk -F [/:] '{print $9\":54\"}' | tr '\n' ',' > /tmp/ns && \
       awk '/^nameserver/ {print $2; exit}' /run/systemd/resolve/resolv.conf >> /tmp/ns"
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:rw \
       --volume /etc/hosts:/etc/hosts:ro \
       janeczku/go-dnsmasq:release-1.0.6 \
       --listen $(hostname -i) \
       --nameservers $(cat /tmp/ns) \
       --hostsfile /etc/hosts \
       --hostsfile-poll 60 \
       --default-resolver \
       --search-domains $(hostname -d | cut -d. -f-2).mesos,$(hostname -d) \
       --append-search-domains"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "mesos-agent.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos agent
     After=docker.service go-dnsmasq.service
     Wants=go-dnsmasq.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull mesosphere/mesos-slave:0.28.1
     ExecStart=/usr/bin/sh -c "docker run \
       --privileged \
       --net host \
       --pid host \
       --name %p \
       --volume /sys:/sys \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --volume /usr/bin/docker:/usr/bin/docker:ro \
       --volume /var/run/docker.sock:/var/run/docker.sock:rw \
       --volume /lib64/libdevmapper.so.1.02:/lib/libdevmapper.so.1.02:ro \
       --volume /lib64/libsystemd.so.0:/lib/libsystemd.so.0:ro \
       --volume /lib64/libgcrypt.so.20:/lib/libgcrypt.so.20:ro \
       --volume /lib64/libgpg-error.so.0:/lib/x86_64-linux-gnu/libgpg-error.so.0:ro \
       --volume /var/lib/mesos:/var/lib/mesos:rw \
       --volume /etc/certs:/etc/certs:ro \
       mesosphere/mesos-slave:0.28.1 \
       --hostname=worker-${KATO_HOST_ID}.$(hostname -d) \
       --ip=$(hostname -i) \
       --containerizers=docker \
       --executor_registration_timeout=2mins \
       --master=zk://${KATO_ZK}/mesos \
       --work_dir=/var/lib/mesos/node \
       --log_dir=/var/log/mesos/node"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "marathon-lb.service"
    command: "start"
    content: |
     [Unit]
     Description=Marathon load balancer
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=-/usr/bin/docker pull mesosphere/marathon-lb:v1.3.0
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --privileged \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --env PORTS=9090,9091 \
       mesosphere/marathon-lb:v1.3.0 sse \
       --marathon http://marathon:8080 \
       --health-check \
       --group external \
       --group internal"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
{{if .SysdigAccessKey}}  - name: "sysdig-agent.service"
    command: "start"
    content: |
     [Unit]
     Description=Sysdig Cloud Agent
     Requires=docker.service
     After=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=/usr/bin/docker pull sysdig/agent
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --privileged \
       --net host \
       --pid host \
       --env ACCESS_KEY={{.SysdigAccessKey}} \
       --env TAGS=name:${KATO_HOST_NAME} \
       --volume /var/run/docker.sock:/host/var/run/docker.sock \
       --volume /dev:/host/dev \
       --volume /proc:/host/proc:ro \
       --volume /boot:/host/boot:ro \
       sysdig/agent"
     ExecStop=/usr/bin/docker stop -t 5 %p
{{end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
{{if .CaCert}}  - name: "getcerts.service"
    command: "start"
    content: |
     [Unit]
     Description=Get certificates from private S3 bucket
     Requires=docker.service
     After=docker.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/getcerts
{{end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "docker-gc.service"
    command: start
    content: |
     [Unit]
     Description=Docker garbage collector
     Requires=etcd2.service docker.service
     After=etcd2.service docker.service

     [Service]
     Type=oneshot
     WorkingDirectory=/tmp
     ExecStart=/bin/bash -c '\
       docker ps -aq --no-trunc | sort -u > containers.all; \
       docker ps -q --no-trunc | sort -u > containers.running; \
       docker rm $$(comm -23 containers.all containers.running) 2>/dev/null; \
       docker rmi $$(docker images -qf dangling=true) 2>/dev/null; \
       docker volume rm $(docker volume ls -f dangling=true | awk "/^local/ {print $2}") 2>/dev/null; \
       etcdctl set /docker/images/$$(hostname) "$$(docker ps --format "{{"{{"}}.Image{{"}}"}}" | sort -u)"; \
       for i in $$(etcdctl ls /docker/images); do etcdctl get $$i; done | sort -u > images.running; \
       docker images | awk "{print \$$1\\":\\"\$$2}" | sed 1d | sort -u > images.local; \
       for i in $$(comm -23 images.local images.running | grep -v katosys); do docker rmi $$i; done; true'`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "docker-gc.timer"
    command: start
    content: |
     [Unit]
     Description=Run docker-gc.service every 60 minutes

     [Timer]
     OnBootSec=1min
     OnUnitActiveSec=30min`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "haproxy-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus haproxy exporter
     After=docker.service marathon-lb.service
     Requires=docker.service marathon-lb.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull katosys/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       katosys/exporters:v0.1.0-1 haproxy_exporter \
       -haproxy.scrape-uri 'http://localhost:9090/haproxy?stats;csv' \
       -web.listen-address :9102"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "mesos-agent-exporter.service"
    command: "start"
    content: |
     [Unit]
     Description=Prometheus mesos agent exporter
     After=docker.service mesos-agent.service
     Requires=docker.service mesos-agent.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull katosys/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       katosys/exporters:v0.1.0-1 mesos_exporter \
       -slave http://$(hostname):5051 \
       -addr :9105"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master", "worker", "border"},
		},
		data: `
 flannel:
  interface: $private_ipv4`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
		},
		data: `
 etcd2:
  name: "{{.HostName}}-{{.HostID}}"
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} initial-cluster: "quorum-1=http://quorum-1:2380,quorum-2=http://quorum-2:2380,quorum-3=http://quorum-3:2380"
  initial-cluster-state: "new"{{end}}
  advertise-client-urls: "http://$private_ipv4:2379"
  initial-advertise-peer-urls: "http://$private_ipv4:2380"
  listen-client-urls: "http://127.0.0.1:2379,http://$private_ipv4:2379"
  listen-peer-urls: "http://$private_ipv4:2380"`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf:  []string{"master", "worker", "border"},
			noneOf: []string{"quorum"},
		},
		data: `
 etcd2:
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} name: "{{.HostName}}-{{.HostID}}"
  initial-cluster: "quorum-1=http://quorum-1:2380,quorum-2=http://quorum-2:2380,quorum-3=http://quorum-3:2380"{{end}}
  advertise-client-urls: "http://$private_ipv4:2379"
  listen-client-urls: "http://127.0.0.1:2379,http://$private_ipv4:2379"
  proxy: on`,
	})
}
