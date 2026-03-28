# Photo Metadata Viewer (PhotoMeta)

Project supports multiple run modes:

* **CLI Mode**: Analyze a single file or directory.
* **GUI Mode**: A graphical user interface for easy navigation and analysis.
* **Server Mode**: A web server for remote access to the metadata analysis functionality. Includes a built-in **Web Demo Page** at `/demo`.

## 🚀 Features

* **Metadata Extraction**: Read EXIF, IPTC, and XMP data from images.
* **Batch Processing**: Handle multiple files or directories concurrently.
* **Multilingual Support**: Tag names in English, Ukrainian, Russian, German, French, and Spanish.

## 📦 Installation

To build PhotoMeta from source, you will need **[Go](https://go.dev/dl/)** (version 1.21+) installed.

For detailed instructions on setting up the environment (especially for GUI support), cloning the repository, and compilation options, please refer to our **[Developer Guide](CONTRIBUTING.md#%F0%9F%8F%97-build-and-run)**.

### Quick Build (CLI & Server)

If you already have Go installed and just want the CLI tool:

```bash
git clone https://github.com/DementorAK/photometa.git
cd photometa
go build ./cmd/photometa
```

## 🛠 Usage

**Analyze a single file:**

```bash
./photometa --path <path-to-image>
./photometa -p <path-to-image>
./photometa <path-to-image>
```

**Process a directory:**

```bash
./photometa --path <path-to-directory>
./photometa -p <path-to-directory>
./photometa <path-to-directory>
```

**Run GUI:**

```bash
./photometa gui
./photometa g
```

**Set display language:**

```bash
./photometa --locale ua <path-to-image>   # Ukrainian
./photometa -l de <path-to-image>          # German
```

**List available languages:**

```bash
./photometa --locale
```

**Run Server (default port is 8080):**

```bash
./photometa server [--port <port>]
./photometa s [--port <port>]
```

**Standard Input (Piping):**

PhotoMeta can read image data directly from standard input.

* **Linux / macOS:**

    ```bash
    cat <path> | ./photometa
    ```

* **Windows (Command Prompt):**

    ```cmd
    type <path> | photometa.exe
    ```

* **Windows (PowerShell):**

    ```powershell
    Get-Content <path> -AsByteStream | ./photometa.exe
    ```

**Examples:**

```bash
./photometa ~/photos           # analyze all files in the directory
./photometa server --port 2233 # start server on port 2233
./photometa gui                # working in a graphical interface
./photometa --locale ua photo.jpg  # analyze with Ukrainian tag names
./photometa --locale           # list available languages
```

**Version:**

```bash
./photometa version          # Show version (human-readable)
./photometa version --json   # Show version as JSON (for CI/CD)
```

## 📚 Documentation

Detailed documentation is available in the `docs/` directory:

*   **[GUI Manual](docs/gui_manual.md)**: How to use the graphical interface.
*   **[API Reference](docs/api.md)**: HTTP Server API documentation.
*   **[Architecture Overview](docs/architecture.md)**: Internal design and structure of the project.
*   **[Test Images](integration/testdata/README.md)**: Sample images used for testing metadata extraction.

## 📂 Project Structure

This project follows the **Hexagonal Architecture** (Ports and Adapters) pattern.

#### 🏗️ Architecture

| Layer | Path | Description |
|-----------|-----------|-------------|
| **Core** | `internal/domain` | Domain models and business entities (Entities) |
| **Logic** | `internal/analyzer` | Core application logic and use cases (Services) |
| **Ports** | `internal/port` | Interfaces defining how the core interacts with the world |
| **Adapters** | `internal/adapter` | Implementations of CLI, GUI, and HTTP Server |
| **Support** | `internal/format` | Technical metadata parsers (EXIF, XMP, IPTC) |
| **Support** | `internal/platform` | Infrastructure code like logger and locale |

#### 🛠️ Project Tooling

*   `.github/` — CI/CD workflows and automation.
*   `integration/` — Cross-layer integration tests and `testdata/`.
*   `cmd/` — Application entry points (main packages).
*   `docs/` — Project documentation and branding assets.
*   `.golangci.yml` & `.goreleaser.yml` — Code quality and release manifests.

## 🚀 Releases

Pre-built binaries are available on the [Releases](https://github.com/DementorAK/photometa/releases) page:

| Type | Platforms | Architectures |
|------|-----------|---------------|
| **CLI** | Windows, macOS, Linux | amd64, arm64 |
| **GUI** | Windows, Linux | amd64 |

## Author

### Dmitry Kinash

* 📧 E-mail: [dv.kinash@gmail.com](mailto:dv.kinash@gmail.com)
* 💼 LinkedIn: [dv-kinash](https://www.linkedin.com/in/dv-kinash/)
* 🐙 GitHub: [@DementorAK](https://github.com/DementorAK)

## 📝 License

[MIT License](LICENSE)
