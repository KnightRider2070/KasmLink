#!/bin/bash
set -ex

ARCH=$(arch | sed 's/aarch64/arm64/g' | sed 's/x86_64/amd64/g')
if [[ "${ARCH}" == "arm64" ]]; then
  LIBSSLDEB=$(curl -sL http://ports.ubuntu.com/pool/main/o/openssl/ | awk -F'(href="|">)' '/libssl1.1.*ubuntu2.[0-9][0-9]_arm64.deb/ {print $4}')
  LIBSSLRPM=$(curl -sL https://ap.edge.kernel.org/fedora/releases/39/Everything/aarch64/os/Packages/o/ | awk -F'(href="|">)' '/openssl1.1-1/ {print $2}')
  LIBSSLURL="http://ports.ubuntu.com/pool/main/o/openssl/${LIBSSLDEB}"
  RPMLIBSSL="https://ap.edge.kernel.org/fedora/releases/39/Everything/aarch64/os/Packages/o/${LIBSSLRPM}"
else
  LIBSSLDEB=$(curl -sL http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/ | awk -F'(href="|">)' '/libssl1.1.*ubuntu2.[0-9][0-9]_amd64.deb/ {print $4}')
  LIBSSLRPM=$(curl -sL https://ap.edge.kernel.org/fedora/releases/39/Everything/x86_64/os/Packages/o/ | awk -F'(href="|">)' '/openssl1.1-1.*x86_64/ {print $2}')
  LIBSSLURL="http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/${LIBSSLDEB}"
  RPMLIBSSL="https://ap.edge.kernel.org/fedora/releases/39/Everything/x86_64/os/Packages/o/${LIBSSLRPM}"
fi

# intall squid
SQUID_COMMIT='1149fc830c7edcb383eec390cce2beba16befde5'
wget -qO- https://kasmweb-build-artifacts.s3.amazonaws.com/kasm-squid-builder/${SQUID_COMMIT}/output/kasm-squid-builder_${ARCH}.tar.gz | tar -xzf - -C /

# update squid conf with userService info
  useradd --system --shell /usr/sbin/nologin --home-dir /bin proxy
  groupadd -g 65511 proxy
  usermod -a -G proxy proxy

mkdir /usr/local/squid/etc/ssl_cert -p
chown proxy:proxy /usr/local/squid/etc/ssl_cert -R
chmod 700 /usr/local/squid/etc/ssl_cert -R
cd /usr/local/squid/etc/ssl_cert


/usr/local/squid/libexec/security_file_certgen -c -s /usr/local/squid/var/logs/ssl_db -M 4MB
chown proxy:proxy /usr/local/squid/var/logs/ssl_db -R

chown -R proxy:proxy /usr/local/squid -R

mkdir -p /etc/squid/

# Trick so we can auto re-direct blocked urls to a special page
cat >>/etc/squid/blocked.acl <<EOL
.access_denied
EOL
chown -R proxy:proxy /etc/squid/blocked.acl


  zypper install -yn memcached cyrus-sasl iproute2 libatomic1


# Enable SASL in the memchache config
echo "-S" >> /etc/memcached.conf

mkdir -p /etc/sasl2
cat >>/etc/sasl2/memcached.conf <<EOL
mech_list: plain
log_level: 5
sasldb_path: /etc/sasl2/memcached-sasldb2
EOL


COMMIT_ID="f8a1049969e7bde2fa0814eb3e5e09f4359efca1"
BRANCH="develop"
COMMIT_ID_SHORT=$(echo "${COMMIT_ID}" | cut -c1-6)


wget -qO- https://kasmweb-build-artifacts.s3.amazonaws.com/kasm_squid_adapter/${COMMIT_ID}/kasm_squid_adapter_glibc_${ARCH}_${BRANCH}.${COMMIT_ID_SHORT}.tar.gz | tar xz -C /etc/squid/

echo "${BRANCH}:${COMMIT_ID}" > /etc/squid/kasm_squid_adapter.version
ls -la /etc/squid
chmod +x /etc/squid/kasm_squid_adapter

# FIXME - This likely should be moved somewhere else to be more explicit
# Install Cert utilities

  zypper install -yn mozilla-nss-tools


# Create an empty cert9.db. This will be used by applications like Chrome
mkdir -p $HOME/.pki/nssdb/
certutil -N -d sql:$HOME/.pki/nssdb/ --empty-password
chown 1000:1000 $HOME/.pki/nssdb/


cat >/usr/bin/filter_ready <<EOL
#!/usr/bin/env bash
if [ "\${http_proxy}" == "http://127.0.0.1:3128" ] ;
then
    while netstat -lnt | awk '\$4 ~ /:3128/ {exit 1}'; do sleep 1; done
    echo 'filter is ready'
else
    echo 'filter is not configured'
fi

EOL
chmod +x /usr/bin/filter_ready
