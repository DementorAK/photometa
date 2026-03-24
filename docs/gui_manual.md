# PhotoMeta GUI User Manual

This guide explains how to use the graphical interface of PhotoMeta to analyze and browse image metadata.

## 🚀 Starting the GUI

Ensure you have built the application with GUI support (see [Installation](../README.md#building-with-gui-support)).

Run the application with the `gui` command:

```bash
./photometa gui
```

## 🖥 The Interface

The main window is divided into several areas:

1.  **Top Control Bar**:
    *   **Select Folder**: Opens a system dialog to choose a directory containing images.
    *   **Path Entry**: Displays the currently selected path. You can also manually type a path here.
    *   **Analyze**: Starts the scanning process for the selected directory.
    *   **Language Dropdown**: Select the display language for metadata tag names (English, Русский, Українська, Deutsch, Français, Español).

2.  **Filter Bar**:
    *   **Filter Input**: Type here to search for specific metadata properties (e.g., "ISO", "Date", "Canon").
    *   **Collapse All / Expand All**: Quickly manage the tree view expansion state.

3.  **Main View (Tree/List)**:
    *   The central area displays a hierarchical view of found images and their metadata.
    *   **Root Nodes**: Represent image files found in the directory.
    *   **Child Nodes**: Represent grouped metadata properties (e.g., Shooting, Exif).
    *   **Leaf Nodes**: Specific property values (e.g., `ISO: 100`, `F-Number: 2.8`).

4.  **Status Bar**:
    *   Located at the bottom, displays messages about scanning progress ("Scanning...", "Found X images") or errors.

## 🛠 How to Use

### Analyzing a Directory

1.  Click **Select Folder**.
2.  Navigate to your folder with images and select it.
3.  Click **Analyze**.
4.  Wait for the scan to complete. The status bar will update with the count of processed images.

### Browsing Metadata

*   New items will appear in the main list as **Files**.
*   Click on a file name to **expand** it and see its metadata categories.
*   Click on a category (e.g., "Shooting Params") to see individual tags.

### Searching/Filtering

If you are looking for specific photos (e.g., taken with a specific camera):

1.  Type the keyword in the **Filter** box (e.g., "Sony").
2.  The list will automatically update to show only files containing that metadata tag. Matches are highlighted or isolated depending on the view mode.

### Changing Language

1.  Use the **Language Dropdown** in the top-right area (next to the Analyze button) to select a display language.
2.  Metadata tag names will update to the selected language. If data has already been loaded, the view refreshes automatically.
