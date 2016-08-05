package udata

//---------------------------------------------------------------------------
// CoreOS edge user data:
//---------------------------------------------------------------------------

const templEdge = `#cloud-config

hostname: "{{.HostName}}-{{.HostID}}.{{.Domain}}"

write_files:

 - path: "/etc/hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 {{.HostName}}-{{.HostID}}.{{.Domain}} {{range .Aliases}}{{.}}-{{$.HostID}} {{end}}marathon-lb
    $private_ipv4 {{.HostName}}-{{.HostID}}.int.{{.Domain}}{{range .Aliases}} {{.}}-{{$.HostID}}.int{{end}}

 - path: "/etc/.hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 {{.HostName}}-{{.HostID}}.{{.Domain}} {{range .Aliases}}{{.}}-{{$.HostID}} {{end}}marathon-lb
    $private_ipv4 {{.HostName}}-{{.HostID}}.int.{{.Domain}}{{range .Aliases}} {{.}}-{{$.HostID}}.int{{end}}

 - path: "/etc/resolv.conf"
   content: |
    search {{.Domain}}
    nameserver 8.8.8.8

 - path: "/etc/kato.env"
   content: |
    KATO_CLUSTER_ID={{.ClusterID}}
    KATO_QUORUM_COUNT={{.QuorumCount}}
    KATO_ROLES='{{range .Roles}}{{.}} {{end}}'
    KATO_HOST_NAME={{.HostName}}
    KATO_HOST_ID={{.HostID}}
    KATO_ZK={{.ZkServers}}
    KATO_SYSTEMD_UNITS='{{range .SystemdUnits}}{{.}} {{end}}'

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
    export PATH='/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin'

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
    PUSH=$(sed 's/ marathon-lb//' /etc/.hosts | grep -v localhost)
    for i in $(echo ${KATO_ROLES}); do
    etcdctl set /hosts/${i}/$(hostname -f) "${PUSH}"; done
    KEYS=$(etcdctl ls --recursive /hosts | grep $(hostname -d) | \
    grep -v $(hostname -f) | rev | sort | rev | uniq -s 14 | sort)
    for i in $KEYS; do PULL+=$(etcdctl get ${i})$'\n'; done
    cat /etc/.hosts > /etc/hosts
    echo "${PULL}" >> /etc/hosts

 - path: "/opt/bin/loopssh"
   permissions: "0755"
   content: |
    #!/bin/bash
    G=$(tput setaf 2); N=$(tput sgr0)
    A=$(grep $1 /etc/hosts | awk '{print $2}' | sort -u | grep -v int)
    for i in $A; do echo "${G}--[ $i ]--${N}"; ssh -o UserKnownHostsFile=/dev/null \
    -o StrictHostKeyChecking=no $i -C "${@:2}" 2> /dev/null; done

 - path: "/opt/bin/awscli"
   permissions: "0755"
   content: |
    #!/bin/bash
    docker run -i --rm \
    --volume /home/core/.aws:/root/.aws:ro \
    --volume ${PWD}:/aws \
    katosys/awscli:v1.10.47-1 "${@}"

 - path: "/opt/bin/katostat"
   permissions: "0755"
   content: |
    #!/bin/bash
    source /etc/kato.env
    systemctl -p Id,LoadState,ActiveState,SubState show ${KATO_SYSTEMD_UNITS} | \
    awk 'BEGIN {RS="\n\n"; FS="\n";} {print $2"\t"$3"\t"$4"\t"$1}'

 {{if .CaCert}}- path: "/opt/bin/getcerts"
   permissions: "0755"
   content: |
    #!/bin/bash
    [ -d /etc/certs ] || mkdir /etc/certs && cd /etc/certs
    [ -f certs.tar.bz2 ] || /opt/bin/awscli s3 cp s3://{{.Domain}}/certs.tar.bz2 .
 {{- end}}

 {{if .CaCert}}- path: "/opt/bin/custom-ca"
   permissions: "0755"
   content: |
    #!/bin/bash
    source /etc/kato.env
    [ -f /etc/ssl/certs/${KATO_CLUSTER_ID}.pem ] && {
      ID=$(sed -n 2p /etc/ssl/certs/${KATO_CLUSTER_ID}.pem)
      NU=$(grep -lir $ID /etc/ssl/certs/* | wc -l)
      [ "$NU" -lt "2" ] && update-ca-certificates &> /dev/null
    }
 {{- end}}

coreos:

 units:

  - name: "etcd2.service"
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
     Requires=format-ephemeral.service
     After=format-ephemeral.service

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
     WantedBy=docker.service

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
     WantedBy=multi-user.target

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
     WantedBy=multi-user.target

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

 flannel:
  interface: $private_ipv4

 etcd2:
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} name: "edge-{{.HostID}}"
  initial-cluster: "master-1=http://master-1:2380,master-2=http://master-2:2380,master-3=http://master-3:2380"{{end}}
  advertise-client-urls: "http://$private_ipv4:2379"
  listen-client-urls: "http://127.0.0.1:2379,http://$private_ipv4:2379"
  proxy: on
`
