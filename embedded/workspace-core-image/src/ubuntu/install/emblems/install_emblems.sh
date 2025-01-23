#!/bin/bash
set -e

# Icon list to ingest
IFS=$'\n'
icons='https://upload.wikimedia.org/wikipedia/commons/b/bc/Amazon-S3-Logo.svg|s3
https://upload.wikimedia.org/wikipedia/commons/6/60/Nextcloud_Logo.svg|nextcloud
https://upload.wikimedia.org/wikipedia/commons/3/3c/Microsoft_Office_OneDrive_%282019%E2%80%93present%29.svg|onedrive
https://upload.wikimedia.org/wikipedia/commons/1/12/Google_Drive_icon_%282020%29.svg|gdrive
https://upload.wikimedia.org/wikipedia/commons/7/78/Dropbox_Icon.svg|dropbox
https://kasm-ci.s3.amazonaws.com/kasm.svg|kasm'

# Create the emblems directory
EMBLEM_DIR="/usr/share/icons/hicolor/scalable/emblems"
mkdir -p "$EMBLEM_DIR"

# Download icons and create corresponding .icon files
for icon in $icons; do
  URL=$(echo "${icon}" | awk -F'|' '{print $1}')
  NAME=$(echo "${icon}" | awk -F'|' '{print $2}')

  echo "Downloading icon: $NAME from $URL"
  curl -o "${EMBLEM_DIR}/${NAME}-emblem.svg" -L "${URL}"

  echo "Creating .icon file for $NAME"
  cat >"${EMBLEM_DIR}/${NAME}-emblem.icon" <<EOL
[Icon Data]
DisplayName=${NAME}-emblem
EOL
done

# Update the icon cache
gtk-update-icon-cache -f /usr/share/icons/hicolor

# Add dynamic icons on startup
cat >>/etc/xdg/autostart/emblems.desktop <<EOL
[Desktop Entry]
Type=Application
Name=Folder Emblems
Exec=/dockerstartup/emblems.sh
EOL

chmod +x /etc/xdg/autostart/emblems.desktop

echo "Emblems installed successfully."
