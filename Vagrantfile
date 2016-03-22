require 'fileutils'

Vagrant.require_version ">= 1.6.0"

#------------------------------------------------------------------------------
# Variables:
#------------------------------------------------------------------------------

$vm_cpus = 2
$vm_memory = 1024
$update_channel = "alpha"
$image_version = "current"
$box_url = "https://storage.googleapis.com/%s.release.core-os.net/amd64-usr/%s/coreos_production_vagrant.json"
$katoctl = "katoctl udata -k %s -d %s -i %s -r %s -c %s"
$discovery_url = "https://discovery.etcd.io/new?size=3"
$ns1_api_key = ENV['KATO_NS1_API_KEY']
$domain = ENV['KATO_NS1_DOMAIN']
$ca_cert = "~/certificates/certs/server-crt.pem"

#------------------------------------------------------------------------------
# Generate a new etcd discovery token:
#------------------------------------------------------------------------------

if $discovery_url && ARGV[0].eql?('up')
  require 'open-uri'
  token = open($discovery_url).read.split("/")[-1]
end

#------------------------------------------------------------------------------
# Configure:
#------------------------------------------------------------------------------

Vagrant.configure("2") do |config|

  config.ssh.insert_key = false
  config.vm.box = "coreos-%s" % $update_channel
  config.vm.box_url = $box_url % [$update_channel, $image_version]

  if $image_version != "current"
    config.vm.box_version = $image_version
  end

  config.vm.provider :virtualbox do |v|
    v.check_guest_additions = false
    v.functional_vboxsf     = false
  end

  if Vagrant.has_plugin?("vagrant-vbguest") then
    config.vbguest.auto_update = false
  end

  #-----------------
  # Start 3 masters
  #-----------------

  (1..3).each do |i|

    config.vm.define vm_name = "master-%d" % i do |conf|

      conf.vm.hostname = vm_name

      conf.vm.provider :virtualbox do |vb|
        vb.gui = false
        vb.memory = $vm_memory
        vb.cpus = $vm_cpus
      end

      ip = "172.17.8.#{i+100}"
      conf.vm.network :private_network, ip: ip

      if ARGV[0].eql?('up')

        if $discovery_url
          cmd = $katoctl + " -e %s | gzip --best > user_mdata_%s"
          system cmd % [$ns1_api_key, $domain, i, 'master', $ca_cert, token, i ]
        else
          cmd = $katoctl + " | gzip --best > user_mdata_%s"
          system cmd % [$ns1_api_key, $domain, i, 'master', $ca_cert, i ]
        end

        if File.exist?("user_mdata_%s" % i)
          conf.vm.provision :file, :source => "user_mdata_%s" % i, :destination => "/tmp/vagrantfile-user-data"
          conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
        end

      end
    end
  end

  #---------------
  # Start 3 nodes
  #---------------

  (1..3).each do |i|

    config.vm.define vm_name = "node-%d" % i do |conf|

      conf.vm.hostname = vm_name

      conf.vm.provider :virtualbox do |vb|
        vb.gui = false
        vb.memory = $vm_memory
        vb.cpus = $vm_cpus
      end

      ip = "172.17.8.#{i+110}"
      conf.vm.network :private_network, ip: ip

      if ARGV[0].eql?('up')

        cmd = $katoctl + " | gzip --best > user_ndata_%s"
        system cmd % [$ns1_api_key, $domain, i, 'node', $ca_cert, i ]

        if File.exist?("user_ndata_%s" % i)
          conf.vm.provision :file, :source => "user_ndata_%s" % i, :destination => "/tmp/vagrantfile-user-data"
          conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
        end

      end
    end
  end
end

# vi: set ft=ruby :
