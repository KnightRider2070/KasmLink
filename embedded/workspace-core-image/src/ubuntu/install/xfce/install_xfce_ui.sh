#!/usr/bin/env bash
### every exit != 0 fails the script
set -e

echo "Install Xfce4 UI components"

  zypper install -yn -t pattern xfce
  zypper install -yn \
    gvfs \
    xclip \
    xfce4-terminal \
    xfce4-notifyd \
	  kdialog \
    xset
  # Pidof is no longer shipped in OpenSuse
  ln -s /usr/bin/pgrep /usr/bin/pidof

# Override default login script so users cant log themselves out of the desktop dession
cat >/usr/bin/xfce4-session-logout <<EOL
#!/usr/bin/env bash
notify-send "Logout" "Please logout or destroy this desktop using the Kasm Control Panel" -i /usr/share/icons/ubuntu-mono-dark/actions/22/system-shutdown-panel-restart.svg
EOL

# Add a script for launching Thunar with libnss wrapper.
# This is called by ~.config/xfce4/xfconf/xfce-perchannel-xml/xfce4-session.xml
cat >/usr/bin/execThunar.sh <<EOL
#!/bin/sh
. $STARTUPDIR/generate_container_user
/usr/bin/Thunar --daemon
EOL
chmod +x /usr/bin/execThunar.sh

cat >/usr/bin/desktop_ready <<EOL
#!/usr/bin/env bash
if [ -z \${START_DE+x} ]; then \
  START_DE="xfce4-session"
fi
until pids=\$(pidof \${START_DE}); do sleep .5; done
EOL
chmod +x /usr/bin/desktop_ready

# Change the default behavior of the delete key which is to move to trash. This will now prompt the user to permanently
# delete the file instead of moving it to trash
mkdir -p /etc/xdg/Thunar/
cat >>/etc/xdg/Thunar/accels.scm<<EOL
(gtk_accel_path "<Actions>/ThunarStandardView/delete" "Delete")
(gtk_accel_path "<Actions>/ThunarLauncher/delete" "Delete")
(gtk_accel_path "<Actions>/ThunarLauncher/trash-delete-2" "")
(gtk_accel_path "<Actions>/ThunarLauncher/trash-delete" "")
EOL

# Support desktop icon trust
cat >>/etc/xdg/autostart/desktop-icons.desktop<<EOL
[Desktop Entry]
Type=Application
Name=Desktop Icon Trust
Exec=/dockerstartup/trustdesktop.sh
EOL
chmod +x /etc/xdg/autostart/desktop-icons.desktop
