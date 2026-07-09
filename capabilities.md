# TriGen: AI Developer Context & Capabilities

This document provides a highly structured overview of the **TriGen** application to bring other AI models (like ChatGPT) instantly up to speed on the codebase, architecture, quirks, and exact capabilities.

---

## 🚀 Core Stack & Specifications
* **Language**: Go 1.24 (strictly backward-compatible with 1.24+ runtime compilers).
* **GUI Engine**: Fyne v2 (desktop GUI, cross-compiled using Docker).
* **Network state**: 100% offline. Statically compiled, zero external telemetry or cloud dependencies.

---

## 🛠️ Complete Feature Capabilities

### 1. Triton Text Proto Parser & Generator
* **Self-Contained Tokenizer/Parser**: (`internal/parser/parser.go`) Tokenizes and parses Triton Inference Server `config.pbtxt` files into a Go struct. It handles space-separated lists, comments (`#`), and duplicates (repeated fields like `input` or `output` are parsed into slices).
* **Schema-Compliant Generator**: (`internal/generator/generator.go`) Outputs valid protobuf text syntax conforming strictly to Triton's parameter schemas.

### 2. Dual-Mode Workspaces
* **File Editor Mode**: Actively opens a standalone `.pbtxt` configuration file. The toolbar switches to save directly to that file path, and repository-specific sidebar panels are disabled.
* **Folder Repository Mode**: Opens a model directory tree. Scans existing version folders, unlocks version/binary copying, and exports zipped repositories.
* **State Persistence**: On startup (`main.go`), loads the user's last directory or file path and automatically restores the active view.

### 3. Triton Version Policy & TVIs
* **TVI Computation**: Converts semantic versions (`major.minor.patch`) to Triton Version Integers (TVIs) via the math formula:
  $$\text{TVI} = 10000 \times \text{major} + 100 \times \text{minor} + \text{patch}$$
* **Version Management**: Scans the directory for numeric folders, copies raw model binaries (like `model.onnx` or `model.pt`) into their computed TVI folders, and lists versions in the GUI.

### 4. ZIP Exporter
* **Recursive Packaging**: Archives the entire model folder recursively into a deployable `.zip` file.
* **Cross-Platform Normalizer**: (`internal/exporter/exporter.go`) Sanitizes model names, protects against directory loops (saving the output inside the zipped source directory), and converts Windows backslashes (`\`) to standard zip slashes (`/`).

---

## 🎨 Advanced UI Implementations & Fyne Workarounds

### 1. Dynamic Window Width Adjustments
* To prevent layout cramming, `e.adjustWindowSize()` dynamically changes the window width:
  * **Base width**: `1024px` (Sidebar + standard Form).
  * **Split forms**: Adds `+200px` for multi-pane layouts (`Inputs`, `Outputs`, `Instance Groups`, `Warmup`, `Parameters`).
  * **Live Preview**: Adds `+450px` when the live text-proto preview pane is open.
  * **Split Offsets**: Recalculated dynamically on resize so that Sidebar (`200px`), Forms (`450px`), and Preview (`450px`) stay at fixed sizes.

### 2. Sleek Custom Layouts & Thinned Splitters
* **Thinned Splitters**: Fyne HSplit splitters default to `16px` thickness. TriGen overrides the splitter's `theme.SizeNamePadding` setting to `3`, shrinking the splitter to `6px` with a `1.5px` handle.
* **Theme Isolation**: The splitter children are wrapped back in standard `ThemeOverrides` so their margins, buttons, and input paddings remain normal.
* **Tight Row Heights**: Sidebar button labels are made of `canvas.Text` primitives (which have zero default margins) and arranged using a custom layout engine (`tightVBoxLayout`) forcing exactly `1px` spacing between tab rows.

### 3. Checkbox Dirty-State Bypass Hook
* Fyne triggers checkbox state change callbacks upon programmatical initialization (making new tabs mark files as dirty). TriGen bypasses this by setting checkboxes to `nil` change handlers first, checking their initial values, and only then registering the `OnChanged` listener callback.

---

## 📂 Sitemap & Main Modules

```
trigen/
├── cmd/
│   └── app/
│       └── main.go       # Launcher, ID preferences, and recent session restore
├── internal/
│   ├── model/            # Config structs and CloneConfig() deep cloner
│   ├── parser/           # Lexer tokenizer and protobuf text parser
│   ├── generator/        # Triton schema-compliant config generator
│   ├── validator/        # Real-time warning constraints checkers
│   ├── state/            # State machine (dirty indicator, error mapping)
│   ├── fileio/           # Atomic file reads and writes
│   ├── exporter/         # TVI math, copying binaries, and ZIP packaging
│   ├── templates/        # Built-in PyTorch, TRT, ONNX, and Python LLM configurations
│   └── ui/               # Editor grid shell, custom layouts, and forms
├── releases/             # Force-tracked pre-compiled binaries for Windows & Ubuntu
└── SDD.md                # Software Design Document
```
