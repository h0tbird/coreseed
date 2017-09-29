Vagrant.require_version ">= 1.6.0"

#------------------------------------------------------------------------------
# Variables:
#------------------------------------------------------------------------------

$cluster_id     = ENV['KATO_CLUSTER_ID'] || 'vagrant-kato'
$node_cpus      = ENV['KATO_NODE_CPUS'] || 2
$node_memory    = ENV['KATO_NODE_MEMORY'] || 4096
$kato_version   = ENV['KATO_VERSION'] || 'v0.1.1'
$coreos_channel = ENV['KATO_COREOS_CHANNEL'] || 'alpha'
$coreos_version = ENV['KATO_COREOS_VERSION'] || 'current'
$monitoring     = ENV['KATO_MONITORING'] || false
$domain         = ENV['KATO_DOMAIN'] || 'cell-1.dc-1.kato'
$ip_address     = ENV['KATO_IP_ADDRESS'] || '172.17.8.11'
$tmp_path       = ENV['KATO_TMP_PATH'] || "/tmp/kato"
$code_path      = ENV['KATO_CODE_PATH'] || "~/git/"
$ca_cert_path   = ENV['KATO_CA_CERT_PATH']
$certs_path     = ENV['KATO_CERTS_PATH']

#------------------------------------------------------------------------------
# URLs:
#------------------------------------------------------------------------------

$katoctl_url = "https://github.com/katosys/kato/releases/download/%s/katoctl-%s-%s"
$box_url = "https://storage.googleapis.com/%s.release.core-os.net/amd64-usr/%s/coreos_production_vagrant.json"

#------------------------------------------------------------------------------
# Install plugins:
#------------------------------------------------------------------------------

required_plugins = %w(vagrant-ignition)
plugins_to_install = required_plugins.select { |plugin| not Vagrant.has_plugin? plugin }

if not plugins_to_install.empty?
  puts "Installing plugins: #{plugins_to_install.join(' ')}"
  if system "vagrant plugin install #{plugins_to_install.join(' ')}"
    exec "vagrant #{ARGV.join(' ')}"
  else
    abort "Installation of one or more plugins has failed. Aborting."
  end
end

#------------------------------------------------------------------------------
# Forge the katoctl command:
#------------------------------------------------------------------------------

if ARGV[0].eql?('up')

  Dir.mkdir $tmp_path unless File.exists?($tmp_path)

  if !File.file?($tmp_path + "/katoctl")
    print "Downloading katoctl...\n"
    cmd = "wget -q " + $katoctl_url + " -O " + $tmp_path + "/katoctl"
    system cmd % [ $kato_version, `uname`.tr("\n",'').downcase, `uname -m`.tr("\n",'') ]
    if !system "chmod +x " + $tmp_path + "/katoctl"
      print "Ops! Where is katoctl?\n"
      exit
    end
  end

  $katoctl = $tmp_path + "/katoctl udata " +
    "--roles quorum,master,worker " +
    "--rexray-storage-driver virtualbox " +
    "--rexray-endpoint-ip 172.17.8.1 " +
    "--iaas-provider vagrant-virtualbox " +
    "--cluster-state new " +
    "--quorum-count 1 " +
    "--master-count 1 " +
    "--host-name %s " +
    "--cluster-id %s " +
    "--domain %s " +
    "--host-id %s "

  if $monitoring
    $katoctl = $katoctl + "--prometheus "
  end

end

#------------------------------------------------------------------------------
# Configure:
#------------------------------------------------------------------------------

Vagrant.configure("2") do |config|

  config.ssh.forward_agent = true
  config.ssh.insert_key = false
  config.vm.box = "coreos-%s" % $coreos_channel
  config.vm.box_url = $box_url % [$coreos_channel, $coreos_version]
  config.ignition.enabled = true

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

#------------------------------------------------------------------------------
# Start the all-in-one instance:
#------------------------------------------------------------------------------

  config.vm.define "kato-1" do |conf|

    conf.vm.hostname = "kato-1.%s" % $domain
    config.ignition.hostname = "kato-1.%s" % $domain
    config.ignition.drive_root = $tmp_path
    config.ignition.drive_name = "ignition"

    conf.vm.provider :virtualbox do |vb|
      vb.gui = false
      vb.name = "kato-1.%s" % $domain
      vb.memory = $node_memory
      vb.cpus = $node_cpus
      vb.customize ["modifyvm", :id, "--macaddress1", "auto" ]
      config.ignition.config_obj = vb
      if `VBoxManage showvminfo #{vb.name} 2>/dev/null | grep SATA` == ''
        vb.customize ["storagectl", :id, "--name", "SATA", "--add", "sata"]
      end
    end

    conf.vm.network :private_network, ip: $ip_address
    config.ignition.ip = $ip_address

    # Purge /etc/hosts records:
    if ARGV[0].eql?('destroy') || ARGV[0].eql?('halt') || ARGV[0].eql?('up')
      system "sudo sed -i.bak '/%s quorum-1.%s/d' /etc/hosts" % [ $ip_address, $domain ]
      system "sudo sed -i.bak '/%s master-1.%s/d' /etc/hosts" % [ $ip_address, $domain ]
      system "sudo sed -i.bak '/%s worker-1.%s/d' /etc/hosts" % [ $ip_address, $domain ]
    end

    if ARGV[0].eql?('up')

      # Add /etc/hosts records:
      system 'sudo bash -c "echo %s quorum-1.%s quorum-1 >> /etc/hosts"' % [ $ip_address, $domain ]
      system 'sudo bash -c "echo %s master-1.%s master-1 >> /etc/hosts"' % [ $ip_address, $domain ]
      system 'sudo bash -c "echo %s worker-1.%s worker-1 >> /etc/hosts"' % [ $ip_address, $domain ]

      conf.vm.synced_folder $code_path, "/code", id: "core", :nfs => true, :mount_options => ['nolock,vers=3,udp']

      if $ca_cert_path
        cmd = $katoctl + " --ca-cert-path %s > " + $tmp_path + "/user_data_kato-1"
        system cmd % [ 'kato', $cluster_id, $domain, 1, $ca_cert_path ]
      else
        cmd = $katoctl + " > " + $tmp_path + "/user_data_kato-1"
        system cmd % [ 'kato', $cluster_id, $domain, 1 ]
      end

      if $certs_path
        conf.vm.provision :file, :source => $certs_path, :destination => "/tmp/certs.tar.bz2"
        conf.vm.provision :shell, :inline => "mkdir /etc/certs; mv /tmp/certs.tar.bz2 /etc/certs", :privileged => true
      end

      if File.exist?($tmp_path + "/user_data_kato-1")
        config.ignition.path = $tmp_path + "/user_data_kato-1"
      end

    end
  end
end

# vi: set ft=ruby :
