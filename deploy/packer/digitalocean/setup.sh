#!/bin/bash -e

# Disable interactive apt prompts
export DEBIAN_FRONTEND=noninteractive

# https://stackoverflow.com/questions/54327058/aws-ami-need-to-explicitly-remove-apt-locks-when-provisioning-ami-from-bionic
while [ ! -f /var/lib/cloud/instance/boot-finished ]; do
  echo "Waiting for cloud init ..."
  sleep 5
done

while fuser /var/lib/apt/lists/lock >/dev/null 2>&1 ; do
  echo "Waiting for cloud init ..."
  sleep 5
done

# Add the HashiCorp GPG key.
curl --fail --silent --show-error --location https://apt.releases.hashicorp.com/gpg | \
      gpg --dearmor | \
      sudo dd of=/usr/share/keyrings/hashicorp-archive-keyring.gpg

# Add the official HashiCorp Linux repository.
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
 sudo tee -a /etc/apt/sources.list.d/hashicorp.list

# Add Docker’s official GPG key:
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Add the official Docker Linux repository.
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y \
  consul=1.15.3-* \
  nomad=1.5.6-* \
  docker-ce=5:24.0.0-*
sudo apt-get autoremove
