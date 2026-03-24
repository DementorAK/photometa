/*
Package photometa provides a unified interface for analyzing image metadata (EXIF, IPTC, XMP), leveraging custom internal parsers with i18n support.

It aggregates various metadata formats into a structured, easy-to-use domain model and exposes this functionality through multiple adapters:
  - CLI: Direct file analysis via terminal arguments.
  - GUI: Graphical interface powered by Fyne.
  - Server: HTTP API for remote analysis.

The project follows Hexagonal Architecture (Ports and Adapters) to keep the core domain logic decoupled from the external interfaces.
*/
package photometa
