/*
Package domain defines the core business entities and rules for the PhotoMeta application.

It contains pure Go structs that represent the "Ubiquitous Language" of the project, independent of any external framework or library.

# Core Entities

  - ImageFile: Represents a single analyzed image file, encapsulating its path, name, and extracted metadata.
  - Metadata: A comprehensive structure that organizes various metadata types (EXIF, IPTC, XMP) into logical groups.

# Logic Groups

Metadata is categorized into:
  - Shooting: Camera settings like Exposure Time, Aperture, ISO.
  - Photo: Physical properties like Dimensions, Orientation, Color Space.
  - DateTime: Temporal and spatial information (Date Taken, GPS Coordinates).
  - Equipment: Hardware details (Camera Model, Lens, Software).
  - Author: Copyright and credit information.
  - Other: Any additional or unrecognized tags.
*/
package domain
