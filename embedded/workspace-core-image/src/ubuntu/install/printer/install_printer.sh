#!/usr/bin/env bash
### every exit != 0 fails the script
set -e

STARTUPDIR="/dockerstartup"
DISTRO=${DISTRO:-"default_distro_value"}

echo "DISTRO is set to: $DISTRO"

# Refresh zypper repositories
echo "Refreshing zypper repositories..."
zypper ref || {
echo "Error: Failed to refresh zypper repositories."
exit 1
}

# Install required printer packages
echo "Installing printer packages..."
zypper install -y cups cups-client cups-pdf || {
echo "Error: Failed to install cups packages."
exit 1
}

# Configure cups-pdf output directory
if [[ -f /etc/cups/cups-pdf.conf ]]; then
sed -i -r -e "s:^(Out\s).*:\1/home/kasm-user/PDF:" /etc/cups/cups-pdf.conf
else
echo "Error: /etc/cups/cups-pdf.conf not found"
exit 1
fi

# Download and extract printer service tarball
COMMIT_ID="30ca302fa364051fd4c68982da7c5474a7bda6b8"
BRANCH="develop"
COMMIT_ID_SHORT=$(echo "${COMMIT_ID}" | cut -c1-6)
ARCH=$(arch | sed 's/aarch64/arm64/g' | sed 's/x86_64/amd64/g')

mkdir -p $STARTUPDIR/printer

echo "Downloading and extracting printer service tarball..."
wget -qO- https://kasmweb-build-artifacts.s3.amazonaws.com/kasm_printer_service/${COMMIT_ID}/kasm_printer_service_${ARCH}_${BRANCH}.${COMMIT_ID_SHORT}.tar.gz \
| tar -xvz -C $STARTUPDIR/printer/ || {
echo "Error: Failed to download or extract kasm_printer_service tarball."
exit 1
}

# Save printer version
echo "${BRANCH}:${COMMIT_ID}" > $STARTUPDIR/printer/kasm_printer.version

# Create the printer_ready script
echo "Creating printer_ready script..."
cat >/usr/bin/printer_ready <<EOL
#!/usr/bin/env bash
set -x
if [[ \${KASM_SVC_PRINTER:-1} == 1 ]]; then
PRINTER_NAME=\${KASM_PRINTER_NAME:-Kasm-Printer}
until [[ "\$(lpstat -r)" == "scheduler is running" ]]; do sleep 1; done
echo "Scheduler is running"

until lpstat -p "\$PRINTER_NAME" | grep -q "is idle"; do
    sleep 1
done
echo "Printer \$PRINTER_NAME is idle."
else
echo "Printing service is not enabled"
fi
EOL

chmod +x /usr/bin/printer_ready

echo "Printer setup complete."