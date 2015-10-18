# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.provider "virtualbox" do |vb|
    vb.gui = true
    vb.customize ["modifyvm", :id, "--usb", "on", "--usbehci", "on"]
  end

  config.vm.provision "shell", inline: <<-SHELL
curl https://packagecloud.io/gpg.key 2>/dev/null | sudo apt-key add -
echo "deb https://packagecloud.io/lstoll/packages/ubuntu/ trusty main" > /etc/apt/sources.list.d/packagecloud_io_lstoll.list
sudo apt-get update
sudo apt-get install -y rtlamr git
cd /tmp
curl -O https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz
mkdir -p /usr/local
cd /usr/local
tar -zxvf /tmp/go1.5.1.linux-amd64.tar.gz
echo 'PATH=/usr/local/go/bin:$PATH' > /etc/profile.d/99go_path.sh
echo 'export GOPATH=$HOME/gocode' >> ~vagrant/.profile
echo 'export PATH=$GOPATH/bin:$PATH' >> ~vagrant/.profile
mkdir -p ~vagrant/gocode/src/github.com/lstoll
ln -s /vagrant ~vagrant/gocode/src/github.com/lstoll/sensorhub
chown vagrant ~vagrant/gocode
chown vagrant ~vagrant/.profile
SHELL
end
