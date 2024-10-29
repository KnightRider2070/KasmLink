#!/usr/bin/env bash
set -e


echo "Install some common tools for further installation"
  sed -i 's/download.opensuse.org/mirrorcache-us.opensuse.org/g' /etc/zypp/repos.d/*.repo
  zypper install -yn wget net-tools bzip2 tar vim gzip iputils bc