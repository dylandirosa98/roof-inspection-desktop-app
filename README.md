# Spartan Roof Inspection

Inspector-facing desktop software for reviewing roof photos, validating local hail-damage detections, and generating claim-support PDF reports.

## Features

- Local ONNX hail-damage inference with 640x640 letterbox preprocessing and non-maximum suppression.
- Editable image annotations with explicit reviewer approval.
- Approved damage photos only in the report; approved empty annotations document no damage and are excluded.
- SQLite-backed projects, annotations, report metadata, and last generated report path.
- Letter-size PDF reports with the Spartan Exteriors logo, photo notes, and adaptive pagination.

## Install On Linux

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

The production Linux executable is written to `build/bin/roof-inspection-desktop-app`.

## Workflow

1. Create a project from a folder of roof photos.
2. Run local analysis.
3. Edit boxes and add optional notes.
4. Approve each reviewed image. Approved images with damage boxes are report candidates; approved images with no boxes remain excluded.
5. Enter report details and generate the PDF.

The application stores its SQLite database at `data/app.db`. Do not commit customer images, reports, or databases. Obtain permission and remove customer information before sharing a demo.
