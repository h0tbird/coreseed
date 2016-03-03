//----------------------------------------------------------------------------
// Package membership:
//----------------------------------------------------------------------------

package main

//---------------------------------------------------------------------------
// CoreOS master user data:
//---------------------------------------------------------------------------

const templ_master = `# cloud-config

hostname: "{{.Hostname}}.{{.Domain}}"

write_files:

 - path: "/etc/hosts"
   content: |
    $private_ipv4 {{.Hostname}}.{{.Domain}} {{.Hostname}}
    $private_ipv4 {{.Hostname}}.int.{{.Domain}} {{.Hostname}}.int
    $public_ipv4 {{.Hostname}}.ext.{{.Domain}} {{.Hostname}}.ext

 - path: "/etc/resolv.conf"
   content: |
    search {{.Domain}}
    nameserver 8.8.8.8

 - path: "/etc/docker/certs.d/internal-registry-sys.marathon:5000/ca.crt"
   content: |
    -----BEGIN CERTIFICATE-----
    MIIFyDCCA7CgAwIBAgIJAKT0QG8NnVNcMA0GCSqGSIb3DQEBCwUAMEExCzAJBgNV
    BAYTAkVTMQ8wDQYDVQQHDAZNYWRyaWQxDjAMBgNVBAoMBU1lc29zMREwDwYDVQQD
    DAhTb2Z0b25pYzAeFw0xNjAyMDMxNDE2MTJaFw0xNzAyMDIxNDE2MTJaMEExCzAJ
    BgNVBAYTAkVTMQ8wDQYDVQQHDAZNYWRyaWQxDjAMBgNVBAoMBU1lc29zMREwDwYD
    VQQDDAhTb2Z0b25pYzCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBALlX
    Cweq0Rsc8CHqTEpMVZKnntSS4NGCepjOakI45Tp0u6zWCjoett8AFZiXM98AUz9p
    e/msWV7GZ7alpHpYnkdciU4NhIcG0YWhS43GhVW9qvyL7dQSKaeRxxJNj5i5CbS8
    cppEZZjutimuJdmeiS6k8RDa/23b4Xv3JPmiU5jZiJdkoLcAYXv+RiSGtYhX1/79
    ckM9zEHgbYrXF1f9v1YamsN4Ct2WgUq1XBLyT9EnLzvnG7NhNvgXvfLmu3Pl5Raq
    d91JnR9MDUUM+mUo8eB/NgYW1+hYV5aLkMFzztjjLu0uUb+YyD+0BeGgwCr1oWkX
    vyVzgL7PvyYM+NJhkPJee/fTFowflzsAOgfGC/20IFs0V42o16l/Y1g6wUR7Mowo
    GlgmHMMV1tAar/GQNUYYyJ2wUyydwSkMEhuVtLWbRvgqFgQnao+G91czsnJ7hdbf
    InkjwhHSl7bmUeaUU0x7CLiB/ZB4m/8I/66tEw+jq10+Fj7oirkEN2K6febnEPK2
    d3qwCqVqrmH14d9HyIf1q6/y4eYn5vOvkgZs21pKE3BgMlSqiuSO9WSzJcPABlRg
    8WRjSoUl1KjVza5/xWBycxNHMDPxI9LcaV59myazcoB1EG/sL7PtkSQUXm9L6ovn
    8RdtQkqCdcwU8NVVLUegRGPgFsV54e9AroA9Of73AgMBAAGjgcIwgb8wHQYDVR0O
    BBYEFP23BevDaBDhmwCd7eyhP0X5Qv/4MB8GA1UdIwQYMBaAFP23BevDaBDhmwCd
    7eyhP0X5Qv/4MAwGA1UdEwQFMAMBAf8wbwYDVR0RBGgwZoIKKi5tYXJhdGhvboIb
    Ki5tYXJhdGhvbi5jZWxsLTEubWFkLm1lc29zghIqLmNlbGwtMS5tYWQubWVzb3OC
    FiouY2VsbC0xLm1hZC54bm9vZC5jb22CCWxvY2FsaG9zdIcEfwAAATANBgkqhkiG
    9w0BAQsFAAOCAgEApZFs7krtZCT45b6PeA0iL+CxVfJ8Su53GrafQ/WV37WzcSnF
    mBVYrlo8bKiZ+evOajVkaKvu6dB660d5x7Ki0DsRyL8EK8iyWdj0hAI5RnjhTgxB
    2a6LfnF9O+lgGbsGQ4ViQTfMbHBB/1F+d8psswgt+rT3PehQhk1uzud6Msk68JQl
    Mus3aCt+CcyApmyj7yiHEniFK+kJTAeJuzwAVU/zNHltgFj7bvcFEpb3cM7aBFxi
    Ha/1lB10A2Bdj/7Pg8ip/ipwJkWEk9YWyZRXvms1E50HDm/CbtwpiWRrZAhQdLm1
    b1Xd2MAzviX0mW8pOkEuEQ8GzSMp4VJKUUnvWmudGuMblCA3bCChO6lcF7BC7+xt
    WWQZQGozsHUIgN5SS9rtwjzOnf+SbWLXR611qZGkVUGDfUWfGj+32Z3C8qNIbWkV
    iJfH01cj1Td9CFiE9ngTHG8IeXqgnpgbN8jBusS722qm0cJjNaMXTtEfzILHhKpn
    rzUQOHR7fLtPuRw2n2eaqvtncSTwi4x/Qdtw6aerKuqpkuFVlGTTyJTPiMQOCg00
    aL2BVPinbeDfLO+gVlqlAl88q8xbF8Jj1gnBRnx8AVIBOB1ZQK4rhYoeE5ufqUq0
    T2pxO7jB3jvpHYXm0qwUiKvgglHVmD5sbO7eyGZa+9eKvvNb/FkgM4mbeSo=
    -----END CERTIFICATE-----

 - path: "/etc/systemd/system/docker.service.d/50-docker-opts.conf"
   content: |
    [Service]
    Environment='DOCKER_OPTS=--registry-mirror=http://external-registry-sys.marathon:5000'

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
    PermitUserEnvironment yes

 - path: "/opt/bin/ns1dns"
   permissions: "0755"
   content: |
    #!/bin/bash

    readonly HOST="$(hostname -s)"
    readonly DOMAIN="$(hostname -d)"
    readonly APIURL='https://api.nsone.net/v1'
    readonly APIKEY='{{.Ns1apikey}}'
    declare -A IP=(['ext']='$public_ipv4' ['int']='$private_ipv4')

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

    # Push config:
    ROLE=$(fleetctl list-machines | grep $(hostname -i) | egrep -o 'slave|master' | uniq)

    [ "${ROLE}" ] && {
      PUSH=$(cat /etc/hosts | grep $(hostname -s)) \
      && etcdctl set /hosts/core/${ROLE}/$(hostname) "${PUSH}"
    }

    # Pull config:
    for i in $(etcdctl ls /hosts/core/master 2>/dev/null | sort) \
    $(etcdctl ls /hosts/core/slave 2>/dev/null | sort); do
      PULL+=$(etcdctl get ${i})$'\n'
    done

    [ "${PULL}" ] && echo "${PULL}" | grep -q $(hostname -s) && echo "${PULL}" > /etc/hosts

 - path: "/opt/bin/ceph"
   permissions: "0755"
   content: |
    #!/bin/bash

    readonly CEPH_DOCKER_IMAGE=h0tbird/ceph
    readonly CEPH_DOCKER_TAG=v9.2.0-2
    readonly CEPH_USER=root

    machinename=$(echo "${CEPH_DOCKER_IMAGE}-${CEPH_DOCKER_TAG}" | sed -r 's/[^a-zA-Z0-9_.-]/_/g')
    machinepath="/var/lib/toolbox/${machinename}"
    osrelease="${machinepath}/etc/os-release"

    [ -f ${osrelease} ] || {
      sudo mkdir -p "${machinepath}"
      sudo chown ${USER}: "${machinepath}"
      docker pull "${CEPH_DOCKER_IMAGE}:${CEPH_DOCKER_TAG}"
      docker run --name=${machinename} "${CEPH_DOCKER_IMAGE}:${CEPH_DOCKER_TAG}" /bin/true
      docker export ${machinename} | sudo tar -x -C "${machinepath}" -f -
      docker rm ${machinename}
      sudo touch ${osrelease}
    }

    [ "$1" == 'dryrun' ] || {
      sudo systemd-nspawn \
      --quiet \
      --directory="${machinepath}" \
      --capability=all \
      --share-system \
      --bind=/dev:/dev \
      --bind=/etc/ceph:/etc/ceph \
      --bind=/var/lib/ceph:/var/lib/ceph \
      --user="${CEPH_USER}" \
      --setenv=CMD="$(basename $0)" \
      --setenv=ARG="$*" \
      /bin/bash -c '\
      mount -o remount,rw -t sysfs sysfs /sys; \
      $CMD $ARG'
    }

 - path: "/opt/bin/loopssh"
   permissions: "0755"
   content: |
    #!/bin/bash
    A=$(fleetctl list-machines -fields=ip -no-legend)
    for i in $A; do ssh -o UserKnownHostsFile=/dev/null \
    -o StrictHostKeyChecking=no $i -C "$*"; done

coreos:

 units:

  - name: "etcd2.service"
    command: "start"

  - name: "fleet.service"
    command: "start"

  - name: "flanneld.service"
    command: "start"
    drop-ins:
     - name: "50-network-config.conf"
       content: |
        [Service]
        ExecStartPre=/usr/bin/etcdctl set /coreos.com/network/config '{ "Network": "10.128.0.0/21","SubnetLen": 27,"SubnetMin": "10.128.0.192","SubnetMax": "10.128.7.224","Backend": {"Type": "host-gw"} }'

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

  - name: "ceph-tools.service"
    command: "start"
    content: |
     [Unit]
     Description=Ceph tools
     Requires=docker.service
     After=docker.service

     [Service]
     Type=oneshot
     RemainAfterExit=yes
     ExecStart=/bin/bash -c '\
       [ -h /opt/bin/rbd ] || { ln -fs ceph /opt/bin/rbd; }; \
       [ -h /opt/bin/rados ] || { ln -fs ceph /opt/bin/rados; }; \
       /opt/bin/ceph dryrun'

  - name: "docker-volume-rbd.service"
    command: "start"
    content: |
     [Unit]
     Description=Docker RBD volume plugin
     Requires=docker.service
     After=docker.service

     [Service]
     Restart=on-failure
     RestartSec=10
     TimeoutStartSec=0

     Environment="PATH=/sbin:/bin:/usr/sbin:/usr/bin:/opt/bin"
     ExecStartPre=-/usr/bin/wget https://github.com/h0tbird/docker-volume-rbd/releases/download/v0.1.2/docker-volume-rbd -O /opt/bin/docker-volume-rbd
     ExecStartPre=-/usr/bin/chmod 755 /opt/bin/docker-volume-rbd
     ExecStart=/opt/bin/docker-volume-rbd

 fleet:
  public-ip: "$private_ipv4"
  metadata: "role=master,masterid={{.Hostid}}"

 etcd2:
  name: "{{.Hostname}}"
  initial-cluster: "core-1=http://core-1:2380,core-2=http://core-2:2380,core-3=http://core-3:2380"
  listen-peer-urls: "http://{{.Hostname}}:2380"
  listen-client-urls: "http://127.0.0.1:2379,http://{{.Hostname}}:2379"
  initial-advertise-peer-urls: "http://{{.Hostname}}:2380"
  advertise-client-urls: "http://{{.Hostname}}:2379"
  initial-cluster-state: "new"
`
