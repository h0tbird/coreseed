Vagrant.require_version ">= 1.6.0"

#------------------------------------------------------------------------------
# Variables:
#------------------------------------------------------------------------------

$cluster_id     = ENV['KATO_CLUSTER_ID'] || 'cell-1-demo'
$quorum_count   = ENV['KATO_QUORUM_COUNT'] || 1
$node_cpus      = ENV['KATO_NODE_CPUS'] || 2
$node_memory    = ENV['KATO_NODE_MEMORY'] || 4096
$kato_version   = ENV['KATO_VERSION'] || 'v0.1.0-alpha'
$coreos_channel = ENV['KATO_COREOS_CHANNEL'] || 'stable'
$coreos_version = ENV['KATO_COREOS_VERSION'] || 'current'
$ns1_api_key    = ENV['KATO_NS1_API_KEY'] || 'x'
$domain         = ENV['KATO_DOMAIN'] || 'cell-1.dc-1.demo.lan'
$ca_cert        = ENV['KATO_CA_CERT']
$code_path      = ENV['KATO_CODE_PATH'] || "~/git/"
$box_url        = "https://storage.googleapis.com/%s.release.core-os.net/amd64-usr/%s/coreos_production_vagrant.json"
$discovery_url  = "https://discovery.etcd.io/new?size=%s"

#------------------------------------------------------------------------------
# Forge the katoctl command:
#------------------------------------------------------------------------------

if ARGV[0].eql?('up')

  if !File.file?("katoctl")
    print "Downloading katoctl...\n"
    cmd = "wget -q https://github.com/h0tbird/kato/releases/download/%s/katoctl-%s-%s -O katoctl"
    system cmd % [ $kato_version, `uname`.tr("\n",'').downcase, `uname -m`.tr("\n",'') ]
    if !system "chmod +x katoctl"
      print "Ops! Where is katoctl?\n"
      exit
    end
  end

  $katoctl = "./katoctl udata " +
    "--rexray-storage-driver virtualbox " +
    "--rexray-endpoint-ip 172.17.8.1 " +
    "--flannel-backend host-gw " +
    "--iaas-provider vbox " +
    "--quorum-count %s " +
    "--host-name %s " +
    "--cluster-id %s " +
    "--ns1-api-key %s " +
    "--domain %s " +
    "--host-id %s " +
    "--roles %s " +
    "--etcd-token %s"
end

#------------------------------------------------------------------------------
# Generate a new etcd discovery token:
#------------------------------------------------------------------------------

if $discovery_url && ARGV[0].eql?('up')
  require 'open-uri'
  token = open($discovery_url % $quorum_count).read.split("/")[-1]
end

#------------------------------------------------------------------------------
# Configure:
#------------------------------------------------------------------------------

Vagrant.configure("2") do |config|

  config.ssh.forward_agent = true
  config.ssh.insert_key = false
  config.vm.box = "coreos-%s" % $coreos_channel
  config.vm.box_url = $box_url % [$coreos_channel, $coreos_version]

  if ARGV[0].eql?('up')
    config.vm.provider :virtualbox do |vb|
      vb.customize ["setproperty", "websrvauthlibrary", "null"]
      vb.check_guest_additions = false
      vb.functional_vboxsf = false
      if `pgrep vboxwebsrv` == ''
        `vboxwebsrv -H 0.0.0.0 -b`
      end
    end
  end

  #-----------------------
  # Start the all-in-one:
  #-----------------------

  config.vm.define "kato-1" do |conf|

    conf.vm.hostname = "kato-1.%s" % $domain

    conf.vm.provider :virtualbox do |vb|
      vb.gui = false
      vb.name = "kato-1.%s" % $domain
      vb.memory = $node_memory
      vb.cpus = $node_cpus
      vb.customize ["modifyvm", :id, "--macaddress1", "auto" ]
      if `VBoxManage showvminfo #{vb.name} 2>/dev/null | grep SATA` == ''
        vb.customize ["storagectl", :id, "--name", "SATA", "--add", "sata"]
      end
    end

    ip_pri = "172.17.8.11"
    conf.vm.network :private_network, ip: ip_pri

    if ARGV[0].eql?('up')

      conf.vm.synced_folder $code_path, "/code", id: "core", :nfs => true, :mount_options => ['nolock,vers=3,udp']

      if $ca_cert
        cmd = $katoctl + " -c %s > user_data_kato-1"
        system cmd % [ $quorum_count, 'kato', $cluser_id, $ns1_api_key, $domain, 1, 'quorum,master,worker', token, $ca_cert ]
      else
        cmd = $katoctl + " > user_data_kato-1"
        system cmd % [ $quorum_count, 'kato', $cluster_id, $ns1_api_key, $domain, 1, 'quorum,master,worker', token ]
      end

      if File.exist?("user_data_kato-1")
        conf.vm.provision :file, :source => "user_data_kato-1", :destination => "/tmp/vagrantfile-user-data"
        conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
      end

    end
  end
end

# vi: set ft=ruby :
