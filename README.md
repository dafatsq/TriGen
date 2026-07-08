# Triton Config Studio

Triton Config Studio is a desktop GUI for creating, editing, validating, and exporting NVIDIA Triton Inference Server `config.pbtxt` files and model repository folders.

## Prerequisites

This project is a Go/Fyne desktop app. Fyne uses CGO and native graphics libraries, so each operating system needs a working native toolchain.

- Go compatible with the version in `go.mod`
- Fyne desktop build dependencies for your OS

### macOS

Install Go and Xcode Command Line Tools:

```sh
xcode-select --install
```

### Windows

Install Go and an MSYS2/MinGW-w64 toolchain so CGO can compile Fyne's native dependencies. Build from an environment where the MinGW compiler is on `PATH`.

### Linux Debian/Ubuntu

```sh
sudo apt-get update
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

Other Linux distributions need equivalent Go, GCC, OpenGL, X11, and keyboard-common development packages.

## Run from source

```sh
go run ./cmd/app
```

## Test and vet

```sh
go test ./...
go vet ./...
```

## Build for the current platform

```sh
go build -o bin/triton-config-studio ./cmd/app
```

On Windows, use an `.exe` output name:

```sh
go build -o bin/triton-config-studio.exe ./cmd/app
```

## Package desktop artifacts

Install the Fyne packaging CLI:

```sh
go install fyne.io/tools/cmd/fyne@latest
```

Package on each target OS:

```sh
fyne package -os darwin
fyne package -os windows
fyne package -os linux
```

For reliable release artifacts, build/package on native macOS, Windows, and Linux machines or CI runners. Plain `GOOS`/`GOARCH` cross-compilation is not enough for this app because Fyne requires CGO and target-platform native graphics toolchains.

## Notes for model repository export

- Exported ZIP archives use Triton-compatible `/` archive paths on all OSes.
- Save/export operations avoid fixed temporary filenames so multiple app instances do not collide on `config.pbtxt.tmp`.
- Build artifacts are intentionally ignored; generate OS-specific binaries locally or in CI rather than committing them.
