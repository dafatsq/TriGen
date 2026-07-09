# TriGen

TriGen is a desktop GUI for creating, editing, validating, and exporting NVIDIA Triton Inference Server `config.pbtxt` files and model repository folders.

## Pre-built Releases

For a quick setup without installing Go, C compilers, or graphics libraries, you can run the pre-built files in the `releases/` directory directly:

* **Windows**: Run `releases/TriGen.exe` directly (or unzip `releases/TriGen.exe.zip` first if you prefer).
* **Ubuntu / Linux**: Run the pre-compiled binary `releases/TriGen` directly:
  ```sh
  chmod +x releases/TriGen
  ./releases/TriGen
  ```
  *(Or extract and run the compressed `releases/TriGen.tar.xz` if you prefer).*


## Fresh Clone

```sh
git clone <repo-url> trigen
cd trigen
go mod download
```

Use a Go version compatible with `go.mod`. This app uses Fyne, so development builds need CGO, a C compiler, and native graphics headers/toolchains for the current OS.

## macOS

Install Go and Xcode Command Line Tools:

```sh
xcode-select --install
```

Run from source:

```sh
go run ./cmd/app
```

Build local binary:

```sh
go build -o bin/trigen ./cmd/app
./bin/trigen
```

Package `.app` bundle:

```sh
go install fyne.io/tools/cmd/fyne@latest
(cd cmd/app && fyne package -os darwin -name TriGen)
```

## Windows

Use the MSYS2 MinGW 64-bit shell so CGO can find GCC.

Install dependencies:

```sh
pacman -Syu
pacman -S git mingw-w64-x86_64-toolchain mingw-w64-x86_64-go
echo "export PATH=$PATH:~/Go/bin" >> ~/.bashrc
```

Clone and run:

```sh
git clone <repo-url> trigen
cd trigen
go mod download
go run ./cmd/app
```

Build local `.exe`:

```sh
go build -o bin/trigen.exe ./cmd/app
./bin/trigen.exe
```

Package Windows app:

```sh
go install fyne.io/tools/cmd/fyne@latest
(cd cmd/app && fyne package -os windows -name TriGen)
```

## Linux

Debian/Ubuntu dependencies:

```sh
sudo apt-get update
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

Fedora dependencies:

```sh
sudo dnf install golang golang-misc gcc libXcursor-devel libXrandr-devel mesa-libGL-devel libXi-devel libXinerama-devel libXxf86vm-devel libxkbcommon-devel wayland-devel
```

Run from source:

```sh
go run ./cmd/app
```

Build local binary:

```sh
go build -o bin/trigen ./cmd/app
./bin/trigen
```

Package Linux archive:

```sh
go install fyne.io/tools/cmd/fyne@latest
(cd cmd/app && fyne package -os linux -name TriGen)
```

## Tests

Run before pushing changes:

```sh
go test ./...
go vet ./...
```

`*_test.go` files are test-only. They are not included in `go build` output, but they should stay in Git because they protect parser, validation, export, and UI regression fixes.

## GitHub Build Artifacts

The repo includes `.github/workflows/build.yml`. It builds downloadable artifacts without committing generated binaries:

- `trigen-linux-amd64.tar.gz` for Ubuntu/Linux
- `TriGen-windows-amd64.zip` containing `TriGen-windows-amd64.exe`
- `trigen-macos-arm64.tar.gz` for Apple Silicon macOS

Run it from GitHub:

1. Open the repository on GitHub.
2. Go to `Actions`.
3. Select `Build desktop artifacts`.
4. Click `Run workflow`.
5. Download artifacts from the finished workflow run.

For a release, push a version tag:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Tag builds also create a GitHub Release with the same files attached.

## Release Notes

- Build/package on each target OS or matching CI runner. Plain `GOOS`/`GOARCH` cross-compilation is not reliable for this app because Fyne uses CGO and native graphics toolchains.
- Build outputs belong in `bin/`, `dist/`, or platform package files. Do not commit generated binaries.
- Fyne apps built for end users do not require users to install Go or Fyne development dependencies.

## Model Repository Export

- Exported ZIP archives use Triton-compatible `/` archive paths on all OSes.
- Save/export operations avoid fixed temporary filenames so multiple app instances do not collide on `config.pbtxt.tmp`.
- Export refuses to write a ZIP inside its own source repository folder.
