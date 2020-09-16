#!/bin/sh
cp template/networkd/as{{.Bob.ASNumberSuffix}}_{{.Bob.Name}}.net* /run/systemd/network
sudo systemctl restart systemd-networkd
