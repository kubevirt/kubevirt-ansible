# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = '2'

node_count = ENV['NODE_COUNT'] || 2

NETWORK_BASE = '192.168.50'
INTEGRATION_START_SEGMENT = 20

BOX_IMAGE = "centos/7"

# ###########################
# generate the new secure key
# ###########################
# this new key is used for all VMs instead of randomly generated
# vagrant default this easier the ansible use
if not File.exist?("common/files/keys/.vagrant_access")
  system("ssh-keygen -t rsa -b 4096 -C vagrant@ansible.test -f common/files/keys/.vagrant_access -N \"\"")
end

# ###########################
# Start of the vagrant config
# ###########################
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  if Vagrant.has_plugin?("vagrant-cachier")
    config.cache.scope = :machine
  end

  config.hostmanager.enabled = true
  config.hostmanager.manage_host = true
  config.hostmanager.manage_guest = true
  config.hostmanager.ignore_private_ip = false

  # setup the alternative ssh key
  # new one first, insecure as backup
  # also this one is used for accessing from ansible
  config.ssh.insert_key = false
  config.ssh.forward_agent = true
  config.ssh.private_key_path = ["common/files/keys/.vagrant_access", "~/.vagrant.d/insecure_private_key"]

  config.vm.box = BOX_IMAGE

  config.vm.provider "libvirt" do |libvirt|
    libvirt.cpus = 1
    libvirt.cpu_mode = "host-model"
    libvirt.memory = 1024
    libvirt.driver = 'kvm'
    libvirt.random :model => "random"
  end

  # Suppress the default sync in both CentOS base and CentOS Atomic Host
  config.vm.synced_folder '.', '/vagrant', disabled: true
  config.vm.synced_folder '.', '/home/vagrant/sync', disabled: true


  # ############
  # provisioning
  # ############

  # Configure eth0 via script, will disable NetworkManager and enable legacy network daemon:
  # config.vm.provision "shell", path: "provision/setup.sh", args: [NETWORK_BASE]

  # copy generated key to all machines, the same key will easier the dev setup
  config.vm.provision "file", source: "common/files/keys/.vagrant_access.pub", destination: "~/.ssh/authorized_keys"
  config.vm.provision "file", source: "common/files/keys/.vagrant_access", destination: "~/.ssh/id_rsa"


  # #############
  # setup the VMs
  # #############

  config.vm.define "master" do |master|
    master.vm.network :private_network, ip: "#{NETWORK_BASE}.#{INTEGRATION_START_SEGMENT}"
    master.vm.hostname = "master.openshift.test"
    master.hostmanager.aliases = %w(master)
  end

  (1..node_count).each do |i|
    config.vm.define "node-#{i}" do |node|
      node.vm.network :private_network, ip: "#{NETWORK_BASE}.#{INTEGRATION_START_SEGMENT + i}"
      node.vm.hostname = "node-#{i}.openshift.test"
      node.hostmanager.aliases = %w(node-#{i})
    end
  end

  config.vm.define "admin" do |admin|
    admin.vm.network :private_network, ip: "#{NETWORK_BASE}.#{INTEGRATION_START_SEGMENT + node_count + 1}"
    admin.vm.hostname = "admin.openshift.test"
    admin.hostmanager.aliases = %w(admin)  
  end

  config.vm.provision "shell", inline: <<-SHELL
    #!/bin/bash
    set -xe
    sed -i -e "s/PasswordAuthentication no/PasswordAuthentication yes/" /etc/ssh/sshd_config
    sed -i -e "s/# Hosts */Hosts/" /etc/ssh/ssh_config
    sed -i -e "s/#   StrictHostKeyChecking ask/   StrictHostKeyChecking no/" /etc/ssh/ssh_config
    systemctl restart sshd
    # FIXME, sometimes eth1 does not come up on Vagrant on latest fc26
    sudo ifup eth1
  SHELL

end
