#!/bin/bash

USERNAME=$1
PASSWORD=$2

adduser --gecos "" $USERNAME || exit
echo $USERNAME":"$PASSWORD | chpasswd