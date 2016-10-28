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
    $private_ipv4 {{.HostName}}-{{.HostID}}.{{.Domain}} {{.HostName}}-{{.HostID}} marathon-lb
{{range .Aliases}}    $private_ipv4 {{.}}-{{$.HostID}}.{{$.Domain}} {{.}}-{{$.HostID}}
{{end}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 - path: "/etc/.hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 {{.HostName}}-{{.HostID}}.{{.Domain}} {{.HostName}}-{{.HostID}} marathon-lb
{{range .Aliases}}    $private_ipv4 {{.}}-{{$.HostID}}.{{$.Domain}} {{.}}-{{$.HostID}}
{{end}}`,
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
    KATO_ROLES='{{range .Roles}}{{.}} {{end}}'
    KATO_HOST_NAME={{.HostName}}
    KATO_HOST_ID={{.HostID}}
    KATO_ZK={{.ZkServers}}
    KATO_SYSTEMD_UNITS='{{range .SystemdUnits}}{{.}} {{end}}'
    KATO_ALERT_MANAGERS={{.AlertManagers}}
    DOCKER_VERSION="$(docker -v | awk -F'[ ,]' '{print $3}')"`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"cacert"},
		},
		data: `
 - path: "/etc/ssl/certs/{{.ClusterID}}.pem"
   content: |
    {{.CaCert}}`,
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
			allOf: []string{"ns1"},
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
    -o StrictHostKeyChecking=no -o ConnectTimeout=3 $i -C "${@:2}" 2> /dev/null; done`,
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
    --net host \
    --volume /home/core/.aws:/root/.aws:ro \
    --volume ${PWD}:/aws \
    quay.io/kato/awscli:v1.10.47-1 "${@}"`,
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
			allOf: []string{"cacert"},
		},
		data: `
 - path: "/opt/bin/getcerts"
   permissions: "0755"
   content: |
    #!/bin/bash
    [ -d /etc/certs ] || mkdir /etc/certs && cd /etc/certs
    [ -f certs.tar.bz2 ] || /opt/bin/awscli s3 cp s3://{{.Domain}}/certs.tar.bz2 .`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"cacert"},
		},
		data: `
 - path: "/opt/bin/custom-ca"
   permissions: "0755"
   content: |
    #!/bin/bash
    source /etc/kato.env
    [ -f /etc/ssl/certs/${KATO_CLUSTER_ID}.pem ] && {
      ID=$(sed -n 2p /etc/ssl/certs/${KATO_CLUSTER_ID}.pem)
      NU=$(grep -lir $ID /etc/ssl/certs/* | wc -l)
      [ "$NU" -lt "2" ] && update-ca-certificates &> /dev/null
    }; exit 0`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
 - path: "/etc/alertmanager/config.yml"
   permissions: "0600"
   content: |
    global:
{{- if .SMTPURL}}
      smtp_smarthost: {{.SMTPHost}}:{{.SMTPPort}}
      smtp_from: alertmanager@{{.Domain}}
      smtp_auth_username: {{.SMTPUser}}
      smtp_auth_password: {{.SMTPPass}}{{end}}
{{- if .SlackWebhook}}
      slack_api_url: {{.SlackWebhook}}{{end}}

    templates:
    - '/etc/alertmanager/template/*.tmpl'

    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 3h
      receiver: operators

    receivers:
    - name: 'operators'
{{- if .SMTPURL}}
      email_configs:
{{- if .AdminEmail}}
      - to: '{{.AdminEmail}}'{{end}}{{end}}
{{- if .SlackWebhook}}
      slack_configs:
      - send_resolved: true
        channel: kato{{end}}
`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
 - path: "/etc/prometheus/targets/prometheus.yml"
 - path: "/etc/prometheus/prometheus.yml"
   permissions: "0600"
   content: |
    global:
     external_labels:
      master: {{.HostID}}
     scrape_interval: 15s
     scrape_timeout: 10s
     evaluation_interval: 10s

    rule_files:
     - /etc/prometheus/recording.rules
     - /etc/prometheus/alerting.rules

    alerting:
     alert_relabel_configs:
     - source_labels: [master]
       action: replace
       replacement: 'all'
       target_label: master

    scrape_configs:

     - job_name: 'prometheus'
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/prometheus.yml

     - job_name: 'cadvisor'
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/cadvisor.yml

     - job_name: 'etcd'
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/etcd.yml

     - job_name: 'node'
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/node.yml

     - job_name: 'mesos'
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/mesos.yml

     - job_name: 'haproxy'
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/haproxy.yml

     - job_name: 'zookeeper'
       file_sd_configs:
        - files:
          - /etc/prometheus/targets/zookeeper.yml`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
 - path: "/etc/prometheus/alerting.rules"
   permissions: "0600"
   content: |
    ALERT ScrapeDown
      IF up == 0
      FOR 5m
      LABELS { severity = "page" }
      ANNOTATIONS {
        summary = "Scrape instance {{"{{"}} $labels.instance {{"}}"}} down",
        description = "Job {{"{{"}} $labels.job {{"}}"}} has been down for more than 5 minutes.",
      }`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
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
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:9191{{"{{"}}end{{"}}"}}
      labels:
        role: master
        shard: {{.HostID}}

 - path: "/etc/confd/conf.d/prom-cadvisor.toml"
   content: |
    [template]
    src = "prom-cadvisor.tmpl"
    dest = "/etc/prometheus/targets/cadvisor.yml"
    keys = [
      "/hosts/quorum",
      "/hosts/master",
      "/hosts/worker",
    ]

 - path: "/etc/confd/templates/prom-cadvisor.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:4194{{"{{"}}end{{"}}"}}
      labels:
        role: quorum
        shard: {{.HostID}}
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:4194{{"{{"}}end{{"}}"}}
      labels:
        role: master
        shard: {{.HostID}}
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:4194{{"{{"}}end{{"}}"}}
      labels:
        role: worker
        shard: {{.HostID}}

 - path: "/etc/confd/conf.d/prom-etcd.toml"
   content: |
    [template]
    src = "prom-etcd.tmpl"
    dest = "/etc/prometheus/targets/etcd.yml"
    keys = [
      "/hosts/quorum",
      "/hosts/master",
      "/hosts/worker",
    ]

 - path: "/etc/confd/templates/prom-etcd.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:2379{{"{{"}}end{{"}}"}}
      labels:
        role: quorum
        shard: {{.HostID}}
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:2379{{"{{"}}end{{"}}"}}
      labels:
        role: master
        shard: {{.HostID}}
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:2379{{"{{"}}end{{"}}"}}
      labels:
        role: worker
        shard: {{.HostID}}

 - path: "/etc/confd/conf.d/prom-node.toml"
   content: |
    [template]
    src = "prom-node.tmpl"
    dest = "/etc/prometheus/targets/node.yml"
    keys = [
      "/hosts/quorum",
      "/hosts/master",
      "/hosts/worker",
    ]

 - path: "/etc/confd/templates/prom-node.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:9101{{"{{"}}end{{"}}"}}
      labels:
        role: quorum
        shard: {{.HostID}}
    - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:9101{{"{{"}}end{{"}}"}}
      labels:
        role: master
        shard: {{.HostID}}
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:9101{{"{{"}}end{{"}}"}}
      labels:
        role: worker
        shard: {{.HostID}}

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
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:9104{{"{{"}}end{{"}}"}}
      labels:
        role: master
        shard: {{.HostID}}
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:9105{{"{{"}}end{{"}}"}}
      labels:
        role: worker
        shard: {{.HostID}}

 - path: "/etc/confd/conf.d/prom-haproxy.toml"
   content: |
    [template]
    src = "prom-haproxy.tmpl"
    dest = "/etc/prometheus/targets/haproxy.yml"
    keys = [ "/hosts/worker" ]

 - path: "/etc/confd/templates/prom-haproxy.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:9102{{"{{"}}end{{"}}"}}
      labels:
        role: worker
        shard: {{.HostID}}

 - path: "/etc/confd/conf.d/prom-zookeeper.toml"
   content: |
    [template]
    src = "prom-zookeeper.tmpl"
    dest = "/etc/prometheus/targets/zookeeper.yml"
    keys = [ "/hosts/quorum" ]

 - path: "/etc/confd/templates/prom-zookeeper.tmpl"
   content: |
    - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
      {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:9103{{"{{"}}end{{"}}"}}
      labels:
        role: quorum
        shard: {{.HostID}}`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
coreos:
 units:
  - name: "etcd2.service"
{{- if eq .ClusterState "existing" }}
    command: "stop"
    enable: false
{{- else}}
    command: "start"
{{- end}}`,
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
			allOf: []string{"cacert"},
		},
		data: `
  - name: "custom-ca.service"
    command: "start"
    content: |
     [Unit]
     Description=Re-hash SSL certificates
     Before=docker.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/custom-ca`,
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
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "kato.target"
    command: "start"
    enable: true
    content: |
     [Unit]
     Description=The Káto System

     [Install]
     WantedBy=multi-user.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
		},
		data: `
  - name: "zookeeper.service"
{{- if eq .ClusterState "existing" }}
    command: "stop"
    enable: false
{{- else}}
    enable: true
{{- end}}
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
     ExecStartPre=/usr/bin/docker pull quay.io/kato/zookeeper:v3.4.8-4
     ExecStart=/usr/bin/sh -c 'docker run \
       --net host \
       --name %p \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --volume /var/lib/zookeeper:/var/lib/zookeeper:rw \
       --env ZK_SERVER_ID=${KATO_HOST_ID} \
       --env ZK_TICK_TIME=2000 \
       --env ZK_INIT_LIMIT=5 \
       --env ZK_SYNC_LIMIT=2 \
       --env ZK_SERVERS=$${KATO_ZK//:2181/} \
       --env ZK_DATA_DIR=/var/lib/zookeeper \
       --env ZK_CLIENT_PORT=2181 \
       --env ZK_CLIENT_PORT_ADDRESS=$(hostname -i) \
       --env JMXDISABLE=false \
       quay.io/kato/zookeeper:v3.4.8-4'
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "mesos-master.service"
    enable: true
    content: |
     [Unit]
     Description=Mesos Master
     After=docker.service zookeeper.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill m3s0s-master
     ExecStartPre=-/usr/bin/docker rm m3s0s-master
     ExecStartPre=/usr/bin/sh -c "docker pull quay.io/kato/mesos:v1.0.1-${DOCKER_VERSION}-2"
     ExecStartPre=/usr/bin/echo ruok | ncat quorum-1 2181 | grep -q imok
     ExecStart=/usr/bin/sh -c "docker run \
       --privileged \
       --name m3s0s-master \
       --net host \
       --volume /var/lib/mesos:/var/lib/mesos:rw \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       quay.io/kato/mesos:v1.0.1-${DOCKER_VERSION}-2 mesos-master \
       --hostname=master-${KATO_HOST_ID}.$(hostname -d) \
       --cluster={{.ClusterID}} \
       --ip=$(hostname -i) \
       --zk=zk://${KATO_ZK}/mesos \
       --work_dir=/var/lib/mesos/master \
       --log_dir=/var/log/mesos \
       --quorum=$(($KATO_QUORUM_COUNT/2 + 1))"
     ExecStop=/usr/bin/docker stop -t 5 m3s0s-master

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "mesos-dns.service"
    enable: true
    content: |
     [Unit]
     Description=Mesos DNS
     After=docker.service mesos-master.service go-dnsmasq.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill m3s0s-dns
     ExecStartPre=-/usr/bin/docker rm m3s0s-dns
     ExecStartPre=/usr/bin/docker pull quay.io/kato/mesos-dns:v0.6.0-2
     ExecStart=/usr/bin/sh -c "docker run \
       --name m3s0s-dns \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --env MDNS_ZK=zk://${KATO_ZK}/mesos \
       --env MDNS_REFRESHSECONDS=45 \
       --env MDNS_LISTENER=$(hostname -i) \
       --env MDNS_PORT={{.MesosDNSPort}} \
       --env MDNS_HTTPON=false \
       --env MDNS_TTL=45 \
       --env MDNS_RESOLVERS=8.8.8.8 \
       --env MDNS_DOMAIN=$(hostname -d | cut -d. -f-2).mesos \
       --env MDNS_IPSOURCE=netinfo \
       quay.io/kato/mesos-dns:v0.6.0-2"
     ExecStartPost=/usr/bin/sh -c ' \
       echo search $(hostname -d | cut -d. -f-2).mesos $(hostname -d) > /etc/resolv.conf && \
       echo "nameserver $(hostname -i)" >> /etc/resolv.conf'
     ExecStop=/usr/bin/sh -c ' \
       echo search $(hostname -d) > /etc/resolv.conf && \
       echo "nameserver 8.8.8.8" >> /etc/resolv.conf'
     ExecStop=/usr/bin/docker stop -t 5 m3s0s-dns

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "marathon.service"
    enable: true
    content: |
     [Unit]
     Description=Marathon
     After=docker.service mesos-master.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=/usr/bin/docker pull quay.io/kato/marathon:v1.3.5-1
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --env LIBPROCESS_IP=$(hostname -i) \
       --env LIBPROCESS_PORT=9292 \
       quay.io/kato/marathon:v1.3.5-1 \
       --http_address $(hostname -i) \
       --master zk://${KATO_ZK}/mesos \
       --zk zk://${KATO_ZK}/marathon \
       --task_launch_timeout 240000 \
       --hostname master-${KATO_HOST_ID}.$(hostname -d) \
       --checkpoint"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "confd.service"
    enable: true
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
     ExecStartPre=/usr/bin/docker pull quay.io/kato/confd:v0.11.0-2
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       --volume /etc:/etc:rw \
       quay.io/kato/confd:v0.11.0-2 \
       -node 127.0.0.1:2379 \
       -watch"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "rexray.service"
    enable: true
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
       REXRAY_URL=https://emccode.bintray.com/rexray/stable/0.3.3/rexray-Linux-i386-0.3.3.tar.gz; \
       [ -f /opt/bin/rexray ] || { curl -sL $${REXRAY_URL} | tar -xz -C /opt/bin; }; \
       [ -x /opt/bin/rexray ] || { chmod +x /opt/bin/rexray; }'
     ExecStart=/opt/bin/rexray start -f
     ExecReload=/bin/kill -HUP $MAINPID
     KillMode=process

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "alertmanager.service"
    enable: true
    content: |
     [Unit]
     Description=Alertmanager service
     After=docker.service
     Before=prometheus.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=/usr/bin/docker pull prom/alertmanager:v0.4.2
     ExecStart=/usr/bin/sh -c "docker run \
       --net host \
       --name %p \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --volume /etc/alertmanager:/etc/alertmanager:ro \
       --volume /var/lib/alertmanager:/var/lib/alertmanager:rw \
       prom/alertmanager:v0.4.2 \
       -log.level=info \
       -web.listen-address=$(hostname -i | awk '{print $1}'):9093 \
       -web.external-url=http://master-${KATO_HOST_ID}.$(hostname -d):9093 \
       -config.file=/etc/alertmanager/config.yml \
       -storage.path=/var/lib/alertmanager"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "prometheus.service"
    enable: true
    content: |
     [Unit]
     Description=Prometheus service
     After=docker.service rexray.service confd.service
     Requires=docker.service rexray.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=/usr/bin/docker pull prom/prometheus:v1.2.1
     ExecStartPre=-/usr/bin/docker volume create --name ${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID} -d rexray
     ExecStart=/usr/bin/sh -c "docker run \
       --net host \
       --name %p \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --volume /etc/prometheus:/etc/prometheus:ro \
       --volume ${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID}:/prometheus:rw \
       prom/prometheus:v1.2.1 \
       -config.file=/etc/prometheus/prometheus.yml \
       -storage.local.path=/prometheus \
       -alertmanager.url ${KATO_ALERT_MANAGERS} \
       -web.external-url=http://master-${KATO_HOST_ID}.$(hostname -d):9191 \
       -web.console.libraries=/etc/prometheus/console_libraries \
       -web.console.templates=/etc/prometheus/consoles \
       -web.listen-address=$(hostname -i | awk '{print $1}'):9191"
     ExecStop=/usr/bin/docker stop -t 60 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "cadvisor.service"
    enable: true
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
     ExecStartPre=/usr/bin/docker pull google/cadvisor:v0.24.1
     ExecStart=/usr/bin/sh -c "docker run \
       --net host \
       --name %p \
       --volume /:/rootfs:ro \
       --volume /var/run:/var/run:rw \
       --volume /sys:/sys:ro \
       --volume /var/lib/docker/:/var/lib/docker:ro \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       google/cadvisor:v0.24.1 \
       --listen_ip $(hostname -i) \
       --logtostderr \
       --port=4194"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"ns1"},
		},
		data: `
  - name: "ns1dns.service"
    enable: true
    content: |
     [Unit]
     Description=Publish DNS records to nsone
     Before=etcd2.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/ns1dns

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "etchost.service"
    enable: true
    content: |
     [Unit]
     Description=Stores IP and hostname in etcd
     Requires=etcd2.service
     After=etcd2.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/etchost

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "etchost.timer"
    enable: true
    content: |
     [Unit]
     Description=Run etchost.service every 5 minutes
     Requires=etcd2.service
     After=etcd2.service

     [Timer]
     OnBootSec=1min
     OnUnitActiveSec=5min

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "mesos-master-exporter.service"
    enable: true
    content: |
     [Unit]
     Description=Prometheus mesos master exporter
     After=docker.service mesos-master.service
     Requires=docker.service mesos-master.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill m3s0s-master-exporter
     ExecStartPre=-/usr/bin/docker rm -f m3s0s-master-exporter
     ExecStartPre=/usr/bin/docker pull quay.io/kato/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name m3s0s-master-exporter \
       quay.io/kato/exporters:v0.1.0-1 mesos_exporter \
       -master http://$(hostname):5050 \
       -addr :9104"
     ExecStop=/usr/bin/docker stop -t 5 m3s0s-master-exporter

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "node-exporter.service"
    enable: true
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
     ExecStartPre=/usr/bin/docker pull quay.io/kato/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       quay.io/kato/exporters:v0.1.0-1 node_exporter \
       -web.listen-address :9101"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "zookeeper-exporter.service"
{{- if eq .ClusterState "existing" }}
    command: "stop"
    enable: false
{{- else}}
    enable: true
{{- end}}
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
     ExecStartPre=/usr/bin/docker pull quay.io/kato/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       quay.io/kato/exporters:v0.1.0-1 zookeeper_exporter \
       -web.listen-address :9103 \
       $(echo ${KATO_ZK} | tr , ' ')"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"border"},
		},
		data: `
  - name: "mongodb.service"
    enable: true
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
     ExecStartPre=/usr/bin/docker pull mongo:3.2
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
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"border"},
		},
		data: `
  - name: "pritunl.service"
    enable: true
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
     ExecStartPre=/usr/bin/docker pull quay.io/kato/pritunl:v1.25.1093.62-1
     ExecStart=/usr/bin/sh -c "docker run \
       --privileged \
       --name %p \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --env MONGODB_URI=mongodb://127.0.0.1:27017/pritunl \
       quay.io/kato/pritunl:v1.25.1093.62-1"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "go-dnsmasq.service"
    enable: true
    content: |
     [Unit]
     Description=Lightweight caching DNS proxy
     After=docker.service etchost.timer
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=/usr/bin/docker pull quay.io/kato/go-dnsmasq:v1.0.7-1
     ExecStartPre=/usr/bin/etcdctl ls /hosts/master
     ExecStartPre=/usr/bin/sh -c " \
       { for i in $(etcdctl ls /hosts/master); do \
       etcdctl get $${i} | awk '/master/ {print $1\":{{.MesosDNSPort}}\"}'; done \
       | tr '\n' ','; echo 8.8.8.8; } > /tmp/ns"
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --volume /etc/resolv.conf:/etc/resolv.conf:rw \
       --volume /etc/hosts:/etc/hosts:ro \
       quay.io/kato/go-dnsmasq:v1.0.7-1 \
       --listen $(hostname -i) \
       --nameservers $(cat /tmp/ns) \
       --hostsfile /etc/hosts \
       --hostsfile-poll 60 \
       --default-resolver \
       {{range .StubZones}}--stubzones {{.}} \
       {{end -}}
       --search-domains $(hostname -d | cut -d. -f-2).mesos,$(hostname -d) \
       --append-search-domains"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "mesos-agent.service"
    enable: true
    content: |
     [Unit]
     Description=Mesos agent
     After=docker.service go-dnsmasq.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     EnvironmentFile=/etc/kato.env
     ExecStartPre=-/usr/bin/docker kill m3s0s-agent
     ExecStartPre=-/usr/bin/docker rm m3s0s-agent
     ExecStartPre=/usr/bin/sh -c "docker pull quay.io/kato/mesos:v1.0.1-${DOCKER_VERSION}-2"
     ExecStart=/usr/bin/sh -c "docker run \
       --privileged \
       --net host \
       --pid host \
       --name m3s0s-agent \
       --volume /sys:/sys \
       --volume /etc/cni:/etc/cni:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --volume /etc/certs:/etc/certs:ro \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /var/lib/mesos:/var/lib/mesos:rw \
       --volume /var/run/docker.sock:/var/run/docker.sock:rw \
       quay.io/kato/mesos:v1.0.1-${DOCKER_VERSION}-2 mesos-agent \
       --docker_mesos_image=quay.io/kato/mesos:v1.0.1-${DOCKER_VERSION}-1 \
       --hostname=worker-${KATO_HOST_ID}.$(hostname -d) \
       --ip=$(hostname -i) \
       --containerizers=docker \
       --executor_registration_timeout=2mins \
       --master=zk://${KATO_ZK}/mesos \
       --work_dir=/var/lib/mesos/node \
       --log_dir=/var/log/mesos/node \
       --network_cni_config_dir=/etc/cni \
       --network_cni_plugins_dir=/var/lib/mesos/cni-plugins"
     ExecStop=/usr/bin/docker stop -t 5 m3s0s-agent

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "marathon-lb.service"
    enable: true
    content: |
     [Unit]
     Description=Marathon load balancer
     After=docker.service marathon.service
     Requires=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=/usr/bin/docker pull mesosphere/marathon-lb:v1.4.1
     ExecStartPre=/usr/bin/sh -c "until host marathon; do sleep 3; done"
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --privileged \
       --volume /etc/resolv.conf:/etc/resolv.conf:ro \
       --volume /etc/hosts:/etc/hosts:ro \
       --env PORTS=9090,9091 \
       --env HAPROXY_RELOAD_SIGTERM_DELAY=5 \
       mesosphere/marathon-lb:v1.4.1 sse \
       --marathon http://marathon:8080 \
       --health-check \
       --group external \
       --group internal \
       --haproxy-map"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"sysdig"},
		},
		data: `
  - name: "sysdig-agent.service"
    enable: true
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

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"datadog"},
		},
		data: `
  - name: "datadog-agent.service"
    enable: true
    content: |
     [Unit]
     Description=Datadog Agent
     Requires=docker.service
     After=docker.service

     [Service]
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm %p
     ExecStartPre=/usr/bin/docker pull datadog/docker-dd-agent
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --hostname $(hostname) \
       --volume /var/run/docker.sock:/var/run/docker.sock:ro \
       --volume /proc/:/host/proc/:ro \
       --volume /sys/fs/cgroup/:/host/sys/fs/cgroup:ro \
       --env API_KEY={{.DatadogAPIKey}} \
       datadog/docker-dd-agent"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "cni-plugins.service"
    enable: true
    content: |
     [Unit]
     Description=Get the CNI plugins
     Requires=docker.service
     After=docker.service
     Before=mesos-agent.service

     [Service]
     Type=oneshot
     ExecStart=/usr/bin/sh -c "docker run \
       --name %p \
       --net host \
       --volume /var/lib/mesos/cni-plugins:/tmp \
       quay.io/kato/cni-plugins:v0.3.0-1"

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
			allOf: []string{"cacert"},
		},
		data: `
  - name: "getcerts.service"
    enable: true
    content: |
     [Unit]
     Description=Get certificates from private S3 bucket
     Requires=docker.service
     Before=go-dnsmasq.service
     After=docker.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/getcerts

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "docker-gc.service"
    enable: true
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
       for i in $$(comm -23 images.local images.running | grep -v kato | grep -v mesosphere); \
       do docker rmi $$i; done; true'

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "docker-gc.timer"
    enable: true
    content: |
     [Unit]
     Description=Run docker-gc.service every 12 hours

     [Timer]
     OnBootSec=0s
     OnUnitActiveSec=12h

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "haproxy-exporter.service"
    enable: true
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
     ExecStartPre=/usr/bin/docker pull quay.io/kato/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       quay.io/kato/exporters:v0.1.0-1 haproxy_exporter \
       -haproxy.scrape-uri 'http://localhost:9090/haproxy?stats;csv' \
       -web.listen-address :9102"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
	})

	d.frags = append(d.frags, fragment{
		filter: filter{
			anyOf: []string{"worker"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "mesos-agent-exporter.service"
    enable: true
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
     ExecStartPre=/usr/bin/docker pull quay.io/kato/exporters:v0.1.0-1
     ExecStart=/usr/bin/sh -c "docker run --rm \
       --net host \
       --name %p \
       quay.io/kato/exporters:v0.1.0-1 mesos_exporter \
       -slave http://$(hostname):5051 \
       -addr :9105"
     ExecStop=/usr/bin/docker stop -t 5 %p

     [Install]
     WantedBy=kato.target`,
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
  name: "quorum-{{.HostID}}"
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} initial-cluster: "{{.EtcdServers}}"
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
  initial-cluster: "{{.EtcdServers}}"{{end}}
  advertise-client-urls: "http://$private_ipv4:2379"
  listen-client-urls: "http://127.0.0.1:2379,http://$private_ipv4:2379"
  proxy: on`,
	})
}
