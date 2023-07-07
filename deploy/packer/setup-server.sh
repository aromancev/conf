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

sudo apt-get update
sudo apt-get install -y \
  consul=1.15.3-* \
  nomad=1.5.6-* \
  vault=1.14.0-*
sudo apt-get autoremove
