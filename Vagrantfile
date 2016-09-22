Vagrant.require_version ">= 1.6.0"

#------------------------------------------------------------------------------
# Variables:
#------------------------------------------------------------------------------

$cluster_id     = ENV['KATO_CLUSTER_ID'] || 'vagrant-kato'
$node_cpus      = ENV['KATO_NODE_CPUS'] || 2
$node_memory    = ENV['KATO_NODE_MEMORY'] || 4096
$kato_version   = ENV['KATO_VERSION'] || 'v0.1.0-beta'
$coreos_channel = ENV['KATO_COREOS_CHANNEL'] || 'stable'
$coreos_version = ENV['KATO_COREOS_VERSION'] || 'current'
$domain         = ENV['KATO_DOMAIN'] || 'cell-1.dc-1.kato'
$code_path      = ENV['KATO_CODE_PATH'] || "~/git/"
$ip_address     = ENV['KATO_IP_ADDRESS'] || '172.17.8.11'
$ca_cert        = ENV['KATO_CA_CERT']
$box_url        = "https://storage.googleapis.com/%s.release.core-os.net/amd64-usr/%s/coreos_production_vagrant.json"

#------------------------------------------------------------------------------
# Forge the katoctl command:
#------------------------------------------------------------------------------

if ARGV[0].eql?('up')

  if !File.file?("katoctl")
    print "Downloading katoctl...\n"
    cmd = "wget -q https://github.com/katosys/kato/releases/download/%s/katoctl-%s-%s -O katoctl"
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
    "--quorum-count 1 " +
    "--master-count 1 " +
    "--host-name %s " +
    "--cluster-id %s " +
    "--domain %s " +
    "--host-id %s " +
    "--roles %s "
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

    conf.vm.network :private_network, ip: $ip_address

    if ARGV[0].eql?('destroy') || ARGV[0].eql?('halt') || ARGV[0].eql?('up')
      system "sudo sed -i.bak '/%s quorum-1.%s/d' /etc/hosts" % [ $ip_address, $domain ]
      system "sudo sed -i.bak '/%s master-1.%s/d' /etc/hosts" % [ $ip_address, $domain ]
      system "sudo sed -i.bak '/%s worker-1.%s/d' /etc/hosts" % [ $ip_address, $domain ]
    end

    if ARGV[0].eql?('up')

      system 'sudo bash -c "echo %s quorum-1.%s >> /etc/hosts"' % [ $ip_address, $domain ]
      system 'sudo bash -c "echo %s master-1.%s >> /etc/hosts"' % [ $ip_address, $domain ]
      system 'sudo bash -c "echo %s worker-1.%s >> /etc/hosts"' % [ $ip_address, $domain ]

      conf.vm.synced_folder $code_path, "/code", id: "core", :nfs => true, :mount_options => ['nolock,vers=3,udp']

      if $ca_cert
        cmd = $katoctl + " --ca-cert %s > user_data_kato-1"
        system cmd % [ 'kato', $cluser_id, $domain, 1, 'quorum,master,worker', $ca_cert ]
      else
        cmd = $katoctl + " > user_data_kato-1"
        system cmd % [ 'kato', $cluster_id, $domain, 1, 'quorum,master,worker' ]
      end

      if File.exist?("user_data_kato-1")
        conf.vm.provision :file, :source => "user_data_kato-1", :destination => "/tmp/vagrantfile-user-data"
        conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
      end

    end
  end
end

# vi: set ft=ruby :
