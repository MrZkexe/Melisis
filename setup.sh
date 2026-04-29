#!/bin/bash

FGGREEN='\033[0;32m'
FGYELLOW='\033[0;33m'
FGRED='\033[0;31m'
FGBLUE='\033[0;34m'
RESET='\033[0m'
BOLD='\033[1m'

log_info() {
  printf "${FGBLUE}${BOLD}[INFO]${RESET} %s\n" "$1"
}

log_success() {
  printf "${FGGREEN}${BOLD}[SUCCESS]${RESET} %s\n" "$1"
}

log_warn() {
  printf "${FGYELLOW}${BOLD}[WARNING]${RESET} %s\n" "$1"
}

log_error() {
  printf "${FGRED}${BOLD}[ERROR]${RESET} %s\n" "$1"
}

if [[ "$(id -u)" -ne 0 ]]; then
  log_error "You must run this script as root."
  exit 1
fi

FILE="./melisis"
SAVEFILE="/usr/bin/melisis"
DOWNLOADFILE="https://github.com/MrZkexe/Melisis/releases/download/binario/melisis"

log_info "Starting Melisis installation process..."

log_info "Checking if binary exists locally..."
if [ ! -f "$FILE" ]; then
    log_warn "Binary not found locally."

    if [ -z "$DOWNLOADFILE" ]; then
        log_error "DOWNLOADFILE is empty. Cannot download binary."
        exit 1
    fi

    log_info "Downloading binary from: $DOWNLOADFILE"
    wget "$DOWNLOADFILE" -O "$FILE"

    if [ $? -ne 0 ]; then
        log_error "Download failed."
        exit 1
    fi

    log_success "Download completed."
else
    log_success "Binary found."
fi

log_info "Installing binary to $SAVEFILE..."
mv "$FILE" "$SAVEFILE"
chmod +x "$SAVEFILE"
log_success "Binary installed."

log_info "Running binary once to generate default files..."
"$SAVEFILE" &
PID=$!

sleep 1
log_info "Waiting 5 seconds..."
sleep 5

log_info "Stopping temporary execution..."
kill $PID
log_success "Initial setup completed."

USUARIO="melisis"
log_info "Creating system user: $USUARIO"

if id "$USUARIO" &>/dev/null; then
    log_warn "User already exists."
else
    useradd --system --no-create-home --shell /bin/false $USUARIO
    log_success "User created."
fi

log_info "Setting permissions..."
chown $USUARIO:$USUARIO $SAVEFILE
chmod 550 $SAVEFILE
log_success "Permissions configured."

log_info "Enabling and starting Fail2Ban..."
systemctl enable fail2ban
systemctl start fail2ban
log_success "Fail2Ban is active."

SERVICE_FILE="/etc/systemd/system/melisis.service"

log_info "Creating systemd service file..."

cat <<EOF > $SERVICE_FILE
[Unit]
Description=Melisis SSH honeypot that emulates a server to capture unauthorized access attempts.
After=network.target

[Service]
ExecStart=${SAVEFILE}
User=${USUARIO}
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
Group=${USUARIO}
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

log_success "Service file created."

log_info "Reloading systemd daemon..."
systemctl daemon-reexec
systemctl daemon-reload

# log_info "Enabling Melisis service..."
# systemctl enable melisis

# log_info "Starting Melisis service..."
# systemctl start melisis

log_success "Melisis installation completed successfully!"
