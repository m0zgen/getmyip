#!/bin/bash
# Cretaed by Yevgeniy Gonvharov, https://lab.sys-adm.in

# Envs
# ---------------------------------------------------\
PATH=$PATH:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin
SCRIPT_PATH=$(cd `dirname "${BASH_SOURCE[0]}"` && pwd); cd $SCRIPT_PATH

# Vars
# ---------------------------------------------------\
_APP_NAME="getmyip"
_APP_USER_NAME="getmyipusr"
_DESTINATION=/opt/${_APP_NAME}

SERVER_IP=$(hostname -I | cut -d' ' -f1)
SERVER_NAME=$(hostname)
SYSTEMD_UNIT=/etc/systemd/system/$_APP_NAME.service

# Functions
# ---------------------------------------------------\

# Check is root function
isRoot() {
    if [[ $EUID -ne 0 ]]; then
        echo "You must be a root user" 2>&1
        exit 1
    fi
}

# Check _DESTINATION exists
isDestinationExists() {
    if [[ ! -d $_DESTINATION ]]; then
        mkdir -p $_DESTINATION
    else
        echo "Destination $_DESTINATION exists. Exit. Bye!"
        exit 1
    fi
}

# Create _APP_USER_NAME user
createUser() {
    useradd -r -s /bin/false $_APP_USER_NAME
}

# Copy binary from build to _DESTINATION
copyBinary() {
    cp -f $SCRIPT_PATH/builds/$_APP_NAME $_DESTINATION/
}

# Set pcap capabilities o _DESTINATION and binary
setPcap() {
    setcap cap_net_bind_service=ep $_DESTINATION/$_APP_NAME
#    setcap cap_net_raw,cap_net_admin=eip $_DESTINATION/$_APP_NAME
}

# Allow restart systemd service with _APP_USER_NAME user with sudo
allowRestart() {
    echo "$_APP_USER_NAME ALL=NOPASSWD: /bin/systemctl restart $_APP_NAME.service,/bin/systemctl stop $_APP_NAME.service,/bin/systemctl start $_APP_NAME.service" >> /etc/sudoers.d/$_APP_USER_NAME
}

# Create systemd unit file
createSystemdUnit() {
    cat <<EOF > $SYSTEMD_UNIT
[Unit]
Description=GetMyIP service
After=network.target

[Service]
Type=simple
User=$_APP_USER_NAME
Group=$_APP_USER_NAME
WorkingDirectory=$_DESTINATION
ExecStart=$_DESTINATION/$_APP_NAME
ExecStop=/usr/bin/kill -s TERM ${MAINPID}
Restart=on-failure
RestartSec=5s

StandardOutput=journal
StandardError=journal+console
SyslogIdentifier=${_APP_NAME}

CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE
LimitNOFILE=32768
OOMScoreAdjust=-100
Nice=-1

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd daemon
systemctl daemon-reload

# Enable and start service
systemctl enable --now $_APP_NAME.service

# Check status
systemctl status $_APP_NAME.service
}

# If systemd unit not running
isRunning() {
    if [[ $(systemctl is-active $_APP_NAME.service) != "active" ]]; then
        echo "Service $_APP_NAME is not running. Exit. Bye!"
        exit 1
    else
        echo "Service $_APP_NAME is running"
        echo "Check it: curl -s http://$SERVER_IP:8080"
        echo  "Service installed to $_DESTINATION. Enjoy!. Bye!"
        exot 0
    fi
}

# Check is _APP_USER_NAME user exists
isUserExists() {
    if id "$_APP_USER_NAME" >/dev/null 2>&1; then
        echo "User $_APP_USER_NAME exists. Exit. Bye!"
        exit 1
    else
        echo "User $_APP_USER_NAME does not exist. Create $_APP_USER_NAME user"
        createUser
        isDestinationExists
        copyBinary
        setPcap
        allowRestart
        createSystemdUnit
        isRunning
    fi
}

