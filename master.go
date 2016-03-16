//----------------------------------------------------------------------------
// Package membership:
//----------------------------------------------------------------------------

package main

//---------------------------------------------------------------------------
// CoreOS master user data:
//---------------------------------------------------------------------------

const templ_master = `#cloud-config

hostname: "{{.Hostname}}.{{.Domain}}"

write_files:

 - path: "/etc/hosts"
   content: |
    127.0.0.1 localhost
    $private_ipv4 {{.Hostname}}.{{.Domain}} {{.Hostname}}
    $private_ipv4 {{.Hostname}}.int.{{.Domain}} {{.Hostname}}.int
    $public_ipv4 {{.Hostname}}.ext.{{.Domain}} {{.Hostname}}.ext

 - path: "/etc/resolv.conf"
   content: |
    search {{.Domain}}
    nameserver 8.8.8.8

 {{if .CAcert }}- path: "/etc/docker/certs.d/internal-registry-sys.marathon:5000/ca.crt"
   content: |
    {{.CAcert}}{{end}}

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

    PUSH=$(cat /etc/hosts | grep $(hostname -s)) \
    && etcdctl set /hosts/$(hostname) "${PUSH}"

    PULL='127.0.0.1 localhost'$'\n'
    for i in $(etcdctl ls /hosts 2>/dev/null | sort); do
      PULL+=$(etcdctl get ${i})$'\n'
    done

    echo "${PULL}" | grep -q $(hostname -s) && echo "${PULL}" > /etc/hosts

 - path: "/opt/bin/ceph"
   permissions: "0755"
   content: |
    #!/bin/bash
    sudo rkt run \
    --interactive \
    --net=host \
    --insecure-options=all \
    --stage1-name=coreos.com/rkt/stage1-fly \
    --volume volume-etc-ceph,kind=host,source=/etc/ceph \
    --volume volume-var-lib-ceph,kind=host,source=/var/lib/ceph \
    docker://h0tbird/ceph:v9.2.0-2 \
    --exec /usr/bin/$(basename $0) -- "$@" 2>/dev/null

 - path: "/opt/bin/loopssh"
   permissions: "0755"
   content: |
    #!/bin/bash
    A=$(fleetctl list-machines -fields=ip -no-legend)
    for i in $A; do ssh -o UserKnownHostsFile=/dev/null \
    -o StrictHostKeyChecking=no $i -C "$*"; done

 - path: "/etc/fleet/zookeeper@.service"
   content: |
    [Unit]
    Description=Zookeeper
    After=docker.service
    Before=mesos-master.service marathon.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill zookeeper-%i
    ExecStartPre=-/usr/bin/docker rm zookeeper-%i
    ExecStartPre=-/usr/bin/docker pull h0tbird/zookeeper:v3.4.8-1
    ExecStart=/usr/bin/sh -c "docker run \
      --net host \
      --name zookeeper-%i \
      --env ZK_SERVER_ID=%i \
      --env ZK_TICK_TIME=2000 \
      --env ZK_INIT_LIMIT=5 \
      --env ZK_SYNC_LIMIT=2 \
      --env ZK_SERVERS=core-1,core-2,core-3 \
      --env ZK_DATA_DIR=/var/lib/zookeeper \
      --env ZK_CLIENT_PORT=2181 \
      --env ZK_CLIENT_PORT_ADDRESS=$(hostname -i) \
      --env JMXDISABLE=true \
      h0tbird/zookeeper:v3.4.8-1"
    ExecStop=/usr/bin/docker stop -t 5 zookeeper-%i

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    MachineMetadata="role=master" "masterid=%i"
    X-Conflicts=zookeeper@*.service

 - path: "/etc/fleet/mesos-master.service"
   content: |
    [Unit]
    Description=Mesos Master
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill mesos-master
    ExecStartPre=-/usr/bin/docker rm mesos-master
    ExecStartPre=-/usr/bin/docker pull mesosphere/mesos-master:0.27.2-2.0.15.ubuntu1404
    ExecStart=/usr/bin/sh -c "docker run \
      --privileged \
      --name mesos-master \
      --net host \
      --volume /var/lib/mesos:/var/lib/mesos \
      --volume /etc/resolv.conf:/etc/resolv.conf \
      mesosphere/mesos-master:0.27.2-2.0.15.ubuntu1404 \
      --ip=$(hostname -i) \
      --zk=zk://core-1:2181,core-2:2181,core-3:2181/mesos \
      --work_dir=/var/lib/mesos/master \
      --log_dir=/var/log/mesos \
      --quorum=2"
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
    ExecStartPre=-/usr/bin/docker kill mesos-node
    ExecStartPre=-/usr/bin/docker rm mesos-node
    ExecStartPre=-/usr/bin/docker pull mesosphere/mesos-slave:0.27.2-2.0.15.ubuntu1404
    ExecStart=/usr/bin/sh -c "docker run \
      --privileged \
      --name mesos-node \
      --net host \
      --pid host \
      --volume /sys:/sys \
      --volume /etc/resolv.conf:/etc/resolv.conf \
      --volume /usr/bin/docker:/usr/bin/docker:ro \
      --volume /var/run/docker.sock:/var/run/docker.sock \
      --volume /lib64/libdevmapper.so.1.02:/lib/libdevmapper.so.1.02:ro \
      --volume /lib64/libsystemd.so.0:/lib/libsystemd.so.0:ro \
      --volume /lib64/libgcrypt.so.20:/lib/libgcrypt.so.20:ro \
      --volume /var/lib/mesos:/var/lib/mesos \
      mesosphere/mesos-slave:0.27.2-2.0.15.ubuntu1404 \
      --ip=$(hostname -i) \
      --containerizers=docker \
      --executor_registration_timeout=2mins \
      --master=zk://core-1:2181,core-2:2181,core-3:2181/mesos \
      --work_dir=/var/lib/mesos/node \
      --log_dir=/var/log/mesos/node"
    ExecStop=/usr/bin/docker stop -t 5 mesos-node

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=node

 - path: "/etc/fleet/mesos-dns.service"
   content: |
    [Unit]
    Description=Mesos DNS
    After=docker.service mesos-master.service
    Requires=docker.service mesos-master.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill mesos-dns
    ExecStartPre=-/usr/bin/docker rm mesos-dns
    ExecStartPre=-/usr/bin/docker pull h0tbird/mesos-dns:v0.5.2-1
    ExecStart=/usr/bin/sh -c "docker run \
      --name mesos-dns \
      --net host \
      --env MDNS_ZK=zk://core-1:2181,core-2:2181,core-3:2181/mesos \
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
    After=docker.service mesos-master.service
    Requires=docker.service mesos-master.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0
    ExecStartPre=-/usr/bin/docker kill marathon
    ExecStartPre=-/usr/bin/docker rm marathon
    ExecStartPre=-/usr/bin/docker pull mesosphere/marathon:v0.15.3
    ExecStart=/usr/bin/sh -c "docker run \
      --name marathon \
      --net host \
      --env LIBPROCESS_IP=$(hostname -i) \
      --env LIBPROCESS_PORT=9090 \
      --volume /etc/resolv.conf:/etc/resolv.conf \
      mesosphere/marathon:v0.15.3 \
      --http_address $(hostname -i) \
      --master zk://core-1:2181,core-2:2181,core-3:2181/mesos \
      --zk zk://core-1:2181,core-2:2181,core-3:2181/marathon \
      --task_launch_timeout 240000 \
      --checkpoint"
    ExecStop=/usr/bin/docker stop -t 5 marathon

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=master

 - path: "/etc/fleet/ceph-mon.service"
   content: |
    [Unit]
    Description=Ceph monitor
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0

    ExecStartPre=-/usr/bin/docker kill ceph-mon
    ExecStartPre=-/usr/bin/docker rm ceph-mon
    ExecStartPre=-/usr/bin/docker pull h0tbird/ceph:v9.2.0-2
    ExecStartPre=/opt/bin/ceph2etcd

    ExecStart=/usr/bin/sh -c "docker run \
      --net host \
      --name ceph-mon \
      --volume /var/lib/ceph:/var/lib/ceph \
      --volume /etc/ceph:/etc/ceph \
      --env CLUSTER='ceph' \
      --env MON_IP=$(hostname -i) \
      --env MON_NAME=$(hostname -s) \
      --env CEPH_PUBLIC_NETWORK=10.0.0.0/8 \
      --env KV_TYPE=etcd \
      --env KV_IP=127.0.0.1 \
      --env KV_PORT=2379 \
      h0tbird/ceph:v9.2.0-2 mon"

    ExecStartPost=/usr/bin/sleep 30
    ExecStartPost=/usr/bin/sh -c "docker exec \
      ceph-mon ceph mon getmap -o /etc/ceph/monmap"
    ExecStop=/usr/bin/docker stop -t 5 ceph-mon

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=master

 - path: "/etc/fleet/ceph-osd.service"
   content: |
    [Unit]
    Description=Ceph OSD
    After=docker.service
    Requires=docker.service

    [Service]
    Restart=on-failure
    RestartSec=20
    TimeoutStartSec=0

    ExecStartPre=-/usr/bin/docker kill ceph-osd
    ExecStartPre=-/usr/bin/docker rm ceph-osd
    ExecStartPre=-/usr/bin/docker pull h0tbird/ceph:v9.2.0-2

    ExecStart=/usr/bin/sh -c "docker run \
      --privileged=true \
      --net host \
      --name ceph-osd \
      --volume /var/lib/ceph:/var/lib/ceph \
      --volume /etc/ceph:/etc/ceph \
      --volume /dev:/dev \
      --env CLUSTER='ceph' \
      --env CEPH_GET_ADMIN_KEY=1 \
      --env OSD_DEVICE=/dev/sdb \
      --env OSD_FORCE_ZAP=1 \
      --env KV_TYPE=etcd \
      --env KV_IP=127.0.0.1 \
      --env KV_PORT=2379 \
      h0tbird/ceph:v9.2.0-2 osd"

    ExecStop=/usr/bin/docker stop -t 5 ceph-osd

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=node

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
    ExecStartPre=-/usr/bin/docker pull janeczku/go-dnsmasq:release-1.0.0
    ExecStartPre=/usr/bin/sh -c " \
      dig core-1 core-2 core-3 +short | tr '\n' ',' > /tmp/ns && \
      awk '/^nameserver/ {print $2; exit}' /run/systemd/resolve/resolv.conf >> /tmp/ns"
    ExecStart=/usr/bin/sh -c "docker run \
      --name dnsmasq \
      --net host \
      --volume /etc/resolv.conf:/etc/resolv.conf \
      janeczku/go-dnsmasq:release-1.0.0 \
      --listen $(hostname -i) \
      --nameservers $(cat /tmp/ns) \
      --default-resolver \
      --search-domains $(hostname -d | cut -d. -f-2).mesos,$(hostname -d) \
      --append-search-domains"
    ExecStop=/usr/bin/docker stop -t 5 dnsmasq

    [Install]
    WantedBy=multi-user.target

    [X-Fleet]
    Global=true
    MachineMetadata=role=node

 - path: "/opt/bin/ceph2etcd"
   permissions: "0755"
   content: |
    #!/bin/bash

    readonly CLUSTER_NAME='ceph'
    readonly ETCD_PATH="/ceph-config/${CLUSTER_NAME}"
    readonly DESIRED_HASH=$(cat $0 | md5sum | awk '{ print $1 }')
    readonly E_BAD_CMD=10
    readonly E_BAD_LOCK=11
    readonly E_BAD_SET=12

    function mutex_lock() {

      while true; do

        LOCK=$(etcdctl get ${ETCD_PATH}/block 2> /dev/null)

        [ "$?" -eq 4 ] && {
          etcdctl mk ${ETCD_PATH}/block '0' &>/dev/null || { sleep 1; continue; }
        } || { [ "${LOCK}" != 0 ] && etcdctl watch ${ETCD_PATH}/block &>/dev/null; }

        etcdctl set --swap-with-value '0' ${ETCD_PATH}/block '1' &>/dev/null && return 0
        sleep 2

      done
    }

    function mutex_unlock() {
      etcdctl set ${ETCD_PATH}/block '0'
    }

    function push_config_to_etcd() {

      # auth:
      etcdctl set ${ETCD_PATH}/auth/cephx 'true' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/auth/cephx_require_signatures 'false' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/auth/cephx_cluster_require_signatures 'true' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/auth/cephx_service_require_signatures 'false' &> /dev/null || return -1

      # global:
      etcdctl set ${ETCD_PATH}/global/max_open_files '131072' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/global/osd_pool_default_pg_num '128' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/global/osd_pool_default_pgp_num '128' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/global/osd_pool_default_size '3' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/global/osd_pool_default_min_size '1' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/global/mon_osd_full_ratio '0.95' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/global/mon_osd_nearfull_ratio '0.85' &> /dev/null || return -1

      # mon:
      etcdctl set ${ETCD_PATH}/mon/mon_osd_down_out_interval '600' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/mon/mon_osd_min_down_reporters '4' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/mon/mon_clock_drift_allowed '0.15' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/mon/mon_clock_drift_warn_backoff '30' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/mon/mon_osd_report_timeout '300' &> /dev/null || return -1

      # osd:
      etcdctl set ${ETCD_PATH}/osd/journal_size '100' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/cluster_network '10.128.0.0/25' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/public_network '10.128.0.0/25' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_mkfs_type 'xfs' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_mkfs_options_xfs ' -f -i size=2048' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_mon_heartbeat_interval '30' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/pool_default_crush_rule '0' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_crush_update_on_start 'true' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_objectstore 'filestore' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/filestore_merge_threshold '40' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/filestore_split_multiple '8' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_op_threads '8' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/filestore_op_threads '8' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/filestore_max_sync_interval '5' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_max_scrubs '1' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_recovery_max_active '5' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_max_backfills '2' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_recovery_op_priority '2' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_client_op_priority '63' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_recovery_max_chunk '1048576' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/osd_recovery_threads '1' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/ms_bind_port_min '6800' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/osd/ms_bind_port_max '7100' &> /dev/null || return -1

      # client:
      etcdctl set ${ETCD_PATH}/client/rbd_cache_enabled 'true' &> /dev/null || return -1
      etcdctl set ${ETCD_PATH}/client/rbd_cache_writethrough_until_flush 'false' &> /dev/null || return -1

      # mds:
      etcdctl set ${ETCD_PATH}/mds/mds_cache_size '100000' &> /dev/null || return -1
    }

    function main() {

      # Get the bootstrap lock:
      local MSG1="[Populate etcd] Attempting to acquire the bootstrap lock..."
      local MSG2="[Populate etcd] Ok! Lock acquired"
      local MSG3="[Populate etcd] Ops! Cannot acquire the lock"
      echo ${MSG1}; mutex_lock && echo ${MSG2} || { echo ${MSG3}; exit ${E_BAD_LOCK}; }

      # Get the current hash from etcd:
      local MSG1="[Populate etcd] Getting the current config hash from etcd..."
      local MSG2="[Populate etcd] OK! The hash has been retrieved"
      local MSG3="[Populate etcd] Ops! Cannot retrieve the config hash from etcd"
      local MSG4="[Populate etcd] Ok! config hash initialized"
      local MSG5="[Populate etcd] Ops! Cannot initialize the config hash"
      echo ${MSG1}; CURRENT_HASH=$(etcdctl get ${ETCD_PATH}/config_hash 2> /dev/null) && \
      echo ${MSG2} || { echo ${MSG3}; etcdctl set ${ETCD_PATH}/config_hash 0 &> /dev/null && \
      echo ${MSG4} || { echo ${MSG5}; exit ${E_BAD_SET}; }; }

      # Check whether changes are needed:
      local MSG1="[Populate etcd] Comparing '${DESIRED_HASH}' with '${CURRENT_HASH}'"
      local MSG2="[Populate etcd] Ok! There is no need to push config to etcd"
      local MSG3="[Populate etcd] Ops! There are pending configuration changes"
      echo ${MSG1}; [ "${DESIRED_HASH}" == "${CURRENT_HASH}" ] && mutex_unlock && echo ${MSG2} && exit 0
      echo ${MSG3}

      # Push the config to Etcd:
      local MSG1="[Populate etcd] Pushing configuration to etcd..."
      local MSG2="[Populate etcd] Ok! The config has been updated"
      local MSG3="[Populate etcd] Ops! Something whent wrong while pushing to etcd"
      echo ${MSG1}; push_config_to_etcd && echo ${MSG2} || { echo ${MSG3}; exit ${E_BAD_SET}; }

      # Set the new configuration hash:
      local MSG1="[Populate etcd] Updating the configuration hash in etcd..."
      local MSG2="[Populate etcd] Ok! The new hash is ${DESIRED_HASH}"
      local MSG3="[Populate etcd] Ops! Cannot set the new hash value"
      echo ${MSG1}; etcdctl set ${ETCD_PATH}/config_hash ${DESIRED_HASH} &> /dev/null && \
      echo ${MSG2} || { echo ${MSG3}; exit ${E_BAD_SET}; }

      # Release the bootstrap lock:
      local MSG1="[Populate etcd] Releasing the bootstrap lock..."
      local MSG2="[Populate etcd] Ok! Bootstrap lock released"
      local MSG3="[Populate etcd] Ops! Cannot release the lock"
      echo ${MSG1}; mutex_unlock && echo ${MSG2} || { echo ${MSG3}; exit ${E_BAD_LOCK}; }
    }

    main "@"

coreos:

 units:

  - name: "etcd2.service"
    command: "start"

  - name: "fleet.service"
    command: "start"

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
       rkt --insecure-options=image fetch /usr/share/rkt/stage1-fly.aci; \
       rkt --insecure-options=image fetch docker://h0tbird/ceph:v9.2.0-2'

 fleet:
  public-ip: "$private_ipv4"
  metadata: "role=master,masterid={{.Hostid}}"

 etcd2:
 {{if .EtcdTkn }} discovery: https://discovery.etcd.io/{{.EtcdTkn}}{{else}} name: "{{.Hostname}}"
  initial-cluster: "core-1=http://core-1:2380,core-2=http://core-2:2380,core-3=http://core-3:2380"
  initial-cluster-state: "new"{{end}}
  advertise-client-urls: "http://$private_ipv4:2379"
  initial-advertise-peer-urls: "http://$private_ipv4:2380"
  listen-client-urls: "http://127.0.0.1:2379,http://$private_ipv4:2379"
  listen-peer-urls: "http://$private_ipv4:2380"
`
