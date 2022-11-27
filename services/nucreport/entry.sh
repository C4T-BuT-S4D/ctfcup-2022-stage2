#!/bin/bash

sudo chown -R root:users /users
sudo chmod 0755 /users

service ssh start
sleep 10


sudo -u nuclear -E /usr/local/bin/app
