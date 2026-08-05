# Spartan Roof Inspection

Inspector-facing desktop software for reviewing roof photos, validating local hail-damage detections, and generating claim-support PDF reports.

## Download For Windows

1. Open the [latest release](https://github.com/dylandirosa98/roof-inspection-desktop-app/releases/latest).
2. Download `Spartan-Roof-Inspection-Setup.exe`.
3. Double-click the downloaded file and complete the installer.
4. Open **Spartan Roof Inspection** from the Start menu or desktop shortcut.

The installer includes the application, local ONNX hail-damage model, ONNX Runtime, and WebView2. No repository clone, terminal, Go, Node.js, Wails, or internet connection after installation is required.

Windows may show a SmartScreen warning until the installer is code-signed. Choose **More info** then **Run anyway** only when the installer was downloaded from this repository's GitHub Release.

On first launch, the app starts with an empty local database. Customer images and reports remain on that computer.

## Features

- Local ONNX hail-damage inference with 640x640 letterbox preprocessing and non-maximum suppression.
- Editable image annotations with explicit reviewer approval.
- Approved damage photos only in the report; approved empty annotations document no damage and are excluded.
- SQLite-backed projects, annotations, report metadata, and last generated report path.
- Letter-size PDF reports with the Spartan Exteriors logo, photo notes, and adaptive pagination.

## Install From Source On Linux

From a clone of this repository, install the Wails CLI once:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
```

Then install the application with one command:

```bash
bash scripts/install-linux.sh
```

The installer builds the app, installs it under `~/.local/share/roof-inspection-desktop-app`, copies the ONNX model and runtime, and creates the **Spartan Roof Inspection** desktop launcher with the Spartan icon. Open it from the Applications menu. To add it to the GNOME dock, right-click the running app and select **Pin to Dash**.

## Development

```bash
wails dev
```

## Build And Test

```bash
npm --prefix frontend run build
go test ./...
go vet ./...
wails build
```

The production Linux executable is written to `build/bin/spartan-roof-inspection`.

## Workflow

1. Create a project from a folder of roof photos.
2. Run local analysis.
3. Edit boxes and add optional notes.
4. Approve each reviewed image. Approved images with damage boxes are report candidates; approved images with no boxes remain excluded.
5. Enter report details and generate the PDF.

The application stores its SQLite database at `data/app.db`. Do not commit customer images, reports, or databases. Obtain permission and remove customer information before sharing a demo.
