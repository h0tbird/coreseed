require 'fileutils'

Vagrant.require_version ">= 1.6.0"

#------------------------------------------------------------------------------
# Variables:
#------------------------------------------------------------------------------

$num_instances = 3
$instance_name_prefix = "core"
$update_channel = "alpha"
$image_version = "current"
$box_url = "https://storage.googleapis.com/%s.release.core-os.net/amd64-usr/%s/coreos_production_vagrant.json"
$coreseed = "coreseed data -k %s -d %s -h core-%s -r %s -c %s"
$discovery_url = "https://discovery.etcd.io/new?size=#{$num_instances}"
$vm_gui = false
$vm_memory = 1024
$vm_cpus = 2
$ns1_api_key = "aabbccddeeaabbccddee"
$domain = "cell-1.dc-1.demo.lan"
$role = "master"
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

  (1..$num_instances).each do |i|

    config.vm.define vm_name = "%s-%d" % [$instance_name_prefix, i] do |conf|
    conf.vm.hostname = vm_name

    conf.vm.provider :virtualbox do |vb|
      vb.gui = $vm_gui
      vb.memory = $vm_memory
      vb.cpus = $vm_cpus
    end

    ip = "172.17.8.#{i+100}"
    conf.vm.network :private_network, ip: ip

    if ARGV[0].eql?('up')

      if $discovery_url
        cmd = $coreseed + " -e %s > user_data_%s"
        system cmd % [$ns1_api_key, $domain, i, $role, $ca_cert, token, i ]
      else
        cmd = $coreseed + " > user_data_%s"
        system cmd % [$ns1_api_key, $domain, i, $role, $ca_cert, i ]
      end

      if File.exist?("user_data_%s" % i)
        conf.vm.provision :file, :source => "user_data_%s" % i, :destination => "/tmp/vagrantfile-user-data"
        conf.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true
      end
    end

    end
  end
end

# vi: set ft=ruby :
