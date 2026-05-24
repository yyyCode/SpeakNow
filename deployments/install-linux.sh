#!/usr/bin/env bash
# Install SpeakNow as a systemd service on Linux.
# Usage (from repo root):
#   sudo bash deployments/install-linux.sh
#   sudo bash deployments/install-linux.sh --no-build   # use existing bin/speaknow
set -euo pipefail

INSTALL_PREFIX="${INSTALL_PREFIX:-/opt/speaknow}"
CONFIG_DIR="${CONFIG_DIR:-/etc/speaknow}"
SERVICE_USER="${SERVICE_USER:-speaknow}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DO_BUILD=1

usage() {
  cat <<EOF
Usage: sudo $0 [options]

Options:
  --no-build          Skip "go build"; install bin/speaknow from repo root
  --prefix PATH       Install root (default: /opt/speaknow)
  --config-dir PATH   Config directory (default: /etc/speaknow)
  -h, --help          Show this help

Environment:
  INSTALL_PREFIX, CONFIG_DIR, SERVICE_USER  Same as flags above
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-build) DO_BUILD=0; shift ;;
    --prefix) INSTALL_PREFIX="$2"; shift 2 ;;
    --config-dir) CONFIG_DIR="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Please run as root (e.g. sudo $0)" >&2
  exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
  echo "systemd (systemctl) is required." >&2
  exit 1
fi

if ! id "$SERVICE_USER" &>/dev/null; then
  useradd --system --home-dir "$INSTALL_PREFIX" --shell /usr/sbin/nologin "$SERVICE_USER"
  echo "Created system user: $SERVICE_USER"
fi

mkdir -p "$INSTALL_PREFIX/bin" "$INSTALL_PREFIX/web" "$CONFIG_DIR"

BINARY_SRC="$REPO_ROOT/bin/speaknow"
if [[ "$DO_BUILD" -eq 1 ]]; then
  if ! command -v go >/dev/null 2>&1; then
    echo "Go is not installed. Use --no-build after building locally, or install Go." >&2
    exit 1
  fi
  echo "Building speaknow..."
  (cd "$REPO_ROOT" && CGO_ENABLED=0 go build -o "$BINARY_SRC" ./cmd/server)
else
  if [[ ! -f "$BINARY_SRC" ]]; then
    echo "Missing $BINARY_SRC — run 'go build -o bin/speaknow ./cmd/server' first." >&2
    exit 1
  fi
fi

install -m 0755 "$BINARY_SRC" "$INSTALL_PREFIX/bin/speaknow"
cp -a "$REPO_ROOT/web/." "$INSTALL_PREFIX/web/"

if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
  if [[ -f "$REPO_ROOT/configs/config.yaml" ]]; then
    install -m 0640 "$REPO_ROOT/configs/config.yaml" "$CONFIG_DIR/config.yaml"
  else
    install -m 0640 "$REPO_ROOT/configs/config.example.yaml" "$CONFIG_DIR/config.yaml"
    echo "Installed template config — edit $CONFIG_DIR/config.yaml before starting."
  fi
else
  echo "Keeping existing $CONFIG_DIR/config.yaml"
fi

chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_PREFIX"
chown root:"$SERVICE_USER" "$CONFIG_DIR"
chmod 0750 "$CONFIG_DIR"
chmod 0640 "$CONFIG_DIR/config.yaml"

UNIT_DST="/etc/systemd/system/speaknow.service"
sed \
  -e "s|/opt/speaknow|$INSTALL_PREFIX|g" \
  -e "s|/etc/speaknow|$CONFIG_DIR|g" \
  -e "s|User=speaknow|User=$SERVICE_USER|g" \
  -e "s|Group=speaknow|Group=$SERVICE_USER|g" \
  "$REPO_ROOT/deployments/speaknow.service" > "$UNIT_DST"

systemctl daemon-reload
systemctl enable speaknow.service

echo ""
echo "Install complete."
echo "  Binary:  $INSTALL_PREFIX/bin/speaknow"
echo "  Web:     $INSTALL_PREFIX/web/"
echo "  Config:  $CONFIG_DIR/config.yaml"
echo ""
echo "Next:"
echo "  1. Edit $CONFIG_DIR/config.yaml (ASR keys, Redis, etc.)"
echo "  2. sudo systemctl start speaknow"
echo "  3. sudo systemctl status speaknow"
echo "  4. journalctl -u speaknow -f"
