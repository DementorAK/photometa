# Photo Metadata Viewer (PhotoMeta)

![Project logo](docs/img/photometa_logo.jpg)

Project supports multiple run modes:

* **CLI Mode**: Analyze a single file or directory.
* **GUI Mode**: A graphical user interface for easy navigation and analysis.
* **Server Mode**: A web server for remote access to the metadata analysis functionality. Includes a built-in **Web Demo Page** at `/demo`.

## 🚀 Features

* **Metadata Extraction**: Read EXIF, IPTC, and XMP data from images.
* **Batch Processing**: Handle multiple files or directories concurrently.
* **Multilingual Support**: Tag names in English, Russian, Ukrainian, German, French, and Spanish.

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
./photometa --locale ru <path-to-image>   # Russian
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
./photometa --locale ru photo.jpg  # analyze with Russian tag names
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

## 📂 Project Structure

This project follows the **Hexagonal Architecture** pattern:

| Component | Description |
|-----------|-------------|
| `cmd/photometa` | Application entry point |
| `internal/adapter` | CLI, GUI, HTTP Server adapters |
| `internal/analyzer` | Core application service |
| `internal/domain` | Domain models and business entities |
| `internal/format` | EXIF, IPTC, XMP metadata parsers |
| `internal/platform` | Logger, locale, and assets |
| `internal/port` | Interface definitions |
| `internal/fake` | Test doubles |
| `integration` | Integration tests |
| `docs` | Documentation |

## Author

### Dmitry Kinash

* 📧 E-mail: [dv.kinash@gmail.com](mailto:dv.kinash@gmail.com)
* 💼 LinkedIn: [dv-kinash](https://www.linkedin.com/in/dv-kinash/)
* 🐙 GitHub: [@DementorAK](https://github.com/DementorAK)

## 📝 License

[MIT License](LICENSE)
