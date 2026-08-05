#!/usr/bin/env bash
set -euo pipefail

app_id="roof-inspection-desktop-app"
app_name="Spartan Roof Inspection"
project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
data_home="${XDG_DATA_HOME:-$HOME/.local/share}"
install_dir="$data_home/$app_id"
desktop_dir="$data_home/applications"
icon_dir="$data_home/icons/hicolor/1024x1024/apps"

if ! command -v wails >/dev/null 2>&1; then
  printf 'Wails CLI is required. Install it with:\n'
  printf '  go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0\n'
  exit 1
fi

cd "$project_root"
wails build

mkdir -p "$install_dir/data" "$install_dir/models" "$install_dir/runtime" "$install_dir/sql" "$desktop_dir" "$icon_dir"
install -m 755 "build/bin/$app_id" "$install_dir/$app_id"
cp -a "models/." "$install_dir/models/"
cp -a "runtime/." "$install_dir/runtime/"
cp -a "sql/schema" "$install_dir/sql/"
install -m 644 "build/appicon.png" "$icon_dir/$app_id.png"

cat > "$desktop_dir/$app_id.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=$app_name
Comment=AI-assisted roof inspection and claim-support reporting
Exec=$install_dir/$app_id
Path=$install_dir
Icon=$app_id
Terminal=false
Categories=Office;
StartupNotify=true
EOF

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "$desktop_dir"
fi
if command -v gtk-update-icon-cache >/dev/null 2>&1 && [[ -f "$data_home/icons/hicolor/index.theme" ]]; then
  gtk-update-icon-cache -f "$data_home/icons/hicolor" || true
fi

printf '%s installed. Open it from the Applications menu.\n' "$app_name"
