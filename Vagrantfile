require 'fileutils'

Vagrant.require_version ">= 1.6.0"

#------------------------------------------------------------------------------
# Variables:
#------------------------------------------------------------------------------

$master_count   = ENV['KATO_MASTER_COUNT'] || 3
$node_count     = ENV['KATO_NODE_COUNT'] || 2
$edge_count     = ENV['KATO_EDGE_COUNT'] || 1
$master_cpus    = ENV['KATO_MASTER_CPUS'] || 2
$master_memory  = ENV['KATO_MASTER_MEMORY'] || 1024
$node_cpus      = ENV['KATO_NODE_CPUS'] || 2
$node_memory    = ENV['KATO_NODE_MEMORY'] || 1024
$edge_cpus      = ENV['KATO_EDGE_CPUS'] || 2
$edge_memory    = ENV['KATO_EDGE_MEMORY'] || 1024
$coreos_channel = ENV['KATO_COREOS_CHANNEL'] || 'alpha'
$coreos_version = ENV['KATO_COREOS_VERSION'] || 'current'
$ns1_api_key    = ENV['KATO_NS1_API_KEY'] || 'aabbccddeeaabbccddee'
$domain         = ENV['KATO_DOMAIN'] || 'cell-1.dc-1.demo.lan'
$ca_cert        = ENV['KATO_CA_CERT'] || ''
$box_url        = "https://storage.googleapis.com/%s.release.core-os.net/amd64-usr/%s/coreos_production_vagrant.json"
$katoctl        = "katoctl udata --master-count %s -k %s -d %s -i %s -r %s -c %s -g"
$discovery_url  = "https://discovery.etcd.io/new?size=%s"

#------------------------------------------------------------------------------
# Generate a new etcd discovery token:
#------------------------------------------------------------------------------

if $discovery_url && ARGV[0].eql?('up')
  require 'open-uri'
  token = open($discovery_url % $master_count).read.split("/")[-1]
end

#------------------------------------------------------------------------------
# Configure:
#------------------------------------------------------------------------------

Vagrant.configure("2") do |config|

  config.ssh.forward_agent = true
  config.ssh.insert_key = false
  config.vm.box = "coreos-%s" % $coreos_channel
  config.vm.box_url = $box_url % [$coreos_channel, $coreos_version]

  if $coreos_version != "current"
    config.vm.box_version = $coreos_version
  end

  config.vm.provider :virtualbox do |v|
    v.check_guest_additions = false
    v.functional_vboxsf     = false
  end

  if Vagrant.has_plugin?("vagrant-vbguest") then
    config.vbguest.auto_update = false
  end

  #----------------------------
  # Start master_count masters
  #----------------------------

  (1..$master_count.to_i).each do |i|

    config.vm.define vm_name = "master-%d" % i do |conf|

      conf.vm.hostname = vm_name

      conf.vm.provider :virtualbox do |vb|
        vb.gui = false
        vb.memory = $master_memory
        vb.cpus = $master_cpus
      end

      ip = "172.17.8.#{i+100}"
      conf.vm.network :private_network, ip: ip

      if ARGV[0].eql?('up')

        if $discovery_url
          cmd = $katoctl + " -e %s > user_mdata_%s"
          system cmd % [$master_count, $ns1_api_key, $domain, i, 'master', $ca_cert, token, i ]
        else
          cmd = $katoctl + " > user_mdata_%s"
          system cmd % [$master_count, $ns1_api_key, $domain, i, 'master', $ca_cert, i ]
        end

        if File.exist?("user_mdata_%s" % i)
          conf.vm.provision :file, :source => "user_mdata_%s" % i, :destination => "/tmp/vagrantfile-user-data"
          conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
        end

      end
    end
  end

  #------------------------
  # Start node_count nodes
  #------------------------

  (1..$node_count.to_i).each do |i|

    config.vm.define vm_name = "node-%d" % i do |conf|

      conf.vm.hostname = vm_name

      conf.vm.provider :virtualbox do |vb|
        vb.gui = false
        vb.memory = $node_memory
        vb.cpus = $node_cpus
      end

      ip = "172.17.8.#{i+110}"
      conf.vm.network :private_network, ip: ip

      if ARGV[0].eql?('up')

        cmd = $katoctl + " > user_ndata_%s"
        system cmd % [$master_count, $ns1_api_key, $domain, i, 'node', $ca_cert, i ]

        if File.exist?("user_ndata_%s" % i)
          conf.vm.provision :file, :source => "user_ndata_%s" % i, :destination => "/tmp/vagrantfile-user-data"
          conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
        end

      end
    end
  end

  #------------------------
  # Start edge_count edges
  #------------------------


  (1..$edge_count.to_i).each do |i|

    config.vm.define vm_name = "edge-%d" % i do |conf|

      conf.vm.hostname = vm_name

      conf.vm.provider :virtualbox do |vb|
        vb.gui = false
        vb.memory = $edge_memory
        vb.cpus = $edge_cpus
      end

      ip = "172.17.8.#{i+120}"
      conf.vm.network :private_network, ip: ip

      if ARGV[0].eql?('up')

        cmd = $katoctl + " > user_edata_%s"
        system cmd % [$master_count, $ns1_api_key, $domain, i, 'edge', $ca_cert, i ]

        if File.exist?("user_edata_%s" % i)
          conf.vm.provision :file, :source => "user_edata_%s" % i, :destination => "/tmp/vagrantfile-user-data"
          conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
        end

      end
    end
  end
end

# vi: set ft=ruby :
