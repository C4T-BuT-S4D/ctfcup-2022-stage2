#!/bin/bash

USERNAME=$1
PASSWORD=$2

adduser --gecos "" $USERNAME || exit 1
echo $USERNAME":"$PASSWORD | chpasswd