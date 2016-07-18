package udata

//---------------------------------------------------------------------------
// CoreOS worker user data:
//---------------------------------------------------------------------------

const templWorker = `#cloud-config

hostname: "worker-{{.HostID}}.{{.Domain}}"

write_files:

 - path: "/etc/hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 worker-{{.HostID}}.{{.Domain}} worker-{{.HostID}} marathon-lb
    $private_ipv4 worker-{{.HostID}}.int.{{.Domain}} worker-{{.HostID}}.int

 - path: "/etc/.hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 worker-{{.HostID}}.{{.Domain}} worker-{{.HostID}} marathon-lb
    $private_ipv4 worker-{{.HostID}}.int.{{.Domain}} worker-{{.HostID}}.int

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

 - path: "/opt/bin/getcerts"
   permissions: "0755"
   content: |
    #!/bin/bash

    [ -d /etc/certs ] || mkdir /etc/certs && cd /etc/certs
    /opt/bin/awscli s3 cp s3://{{.Domain}}/certs.tar.bz2 .

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

  - name: "go-dnsmasq.service"
    command: "start"
    content: |
     [Unit]
     Description=Lightweight caching DNS proxy
     After=docker.service
     Requires=docker.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0
     ExecStartPre=-/usr/bin/docker kill %p
     ExecStartPre=-/usr/bin/docker rm -f %p
     ExecStartPre=-/usr/bin/docker pull janeczku/go-dnsmasq:release-1.0.6
     ExecStartPre=/usr/bin/sh -c " \
       etcdctl member list 2>1 | awk -F [/:] '{print $9}' | tr '\n' ',' > /tmp/ns && \
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
     WantedBy=multi-user.target

  - name: "mesos-agent.service"
    command: "start"
    content: |
     [Unit]
     Description=Mesos agent
     After=docker.service go-dnsmasq.service
     Wants=go-dnsmasq.service
     Requires=docker.service

     [Service]
     Restart=on-failure
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
       --ip=$(hostname -i) \
       --containerizers=docker \
       --executor_registration_timeout=2mins \
       --master=zk://${KATO_ZK}/mesos \
       --work_dir=/var/lib/mesos/node \
       --log_dir=/var/log/mesos/node"
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

  - name: "getcerts.service"
    command: "start"
    content: |
     [Unit]
     Description=Get certificates from private S3 bucket
     Requires=docker.service
     After=docker.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/getcerts

  - name: "etchost.service"
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
       for i in $$(comm -23 images.local images.running); do docker rmi $$i; done; true'

  - name: "docker-gc.timer"
    command: start
    content: |
     [Unit]
     Description=Run docker-gc.service every 30 minutes

     [Timer]
     OnBootSec=1min
     OnUnitActiveSec=30min

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

 flannel:
  interface: $private_ipv4

 fleet:
  public-ip: "$private_ipv4"
  metadata: "role=worker,id={{.HostID}}"

 etcd2:
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} name: "worker-{{.HostID}}"
  initial-cluster: "master-1=http://master-1:2380,master-2=http://master-2:2380,master-3=http://master-3:2380"{{end}}
  advertise-client-urls: "http://$private_ipv4:2379"
  listen-client-urls: "http://127.0.0.1:2379,http://$private_ipv4:2379"
  proxy: on
`
