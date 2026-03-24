//go:build gui

package gui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/platform/assets"
	"github.com/DementorAK/photometa/internal/platform/locale"
	"github.com/DementorAK/photometa/internal/port"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	service port.ImageAnalyzer
	app     fyne.App
	window  fyne.Window

	// UI widgets
	listTree     *widget.Tree
	pathEntry    *widget.Entry
	filterEntry  *widget.Entry
	statusLabel  *widget.Label
	localeSelect *widget.Select

	btnFolder      *widget.Button
	btnAnalyze     *widget.Button
	btnCollapseAll *widget.Button
	btnExpandAll   *widget.Button
	filterLabel    *widget.Label

	// Tree data
	tree *TreeModel

	currentPath string
	iconCache   map[string]fyne.Resource
}

func NewGUI(service port.ImageAnalyzer) *GUI {
	a := app.NewWithID("photometa")
	w := a.NewWindow(locale.T("Photo Metadata Viewer"))
	return &GUI{
		service:   service,
		app:       a,
		window:    w,
		tree:      NewTreeModel(),
		iconCache: make(map[string]fyne.Resource),
	}
}

func (g *GUI) Start() {
	g.setupUI()
	g.window.Resize(fyne.NewSize(800, 600))
	g.window.ShowAndRun()
}

func (g *GUI) setupUI() {
	g.statusLabel = widget.NewLabel(locale.T("Ready"))
	g.pathEntry = widget.NewEntry()
	g.pathEntry.SetPlaceHolder(locale.T("Select a folder or enter path..."))

	g.filterEntry = widget.NewEntry()
	g.filterEntry.SetPlaceHolder(locale.T("Filter by property name..."))
	g.filterEntry.OnChanged = func(text string) {
		g.tree.Filter(text)
		g.listTree.Refresh()
		if g.tree.isFiltered {
			for _, root := range g.tree.filteredRoots {
				g.listTree.OpenBranch(root)
			}
		}
	}

	g.listTree = widget.NewTree(
		// ChildUIDs
		func(id string) []string {
			if id == "" {
				if g.tree.isFiltered {
					return g.tree.filteredRoots
				}
				return g.tree.roots
			} else if node, ok := g.tree.nodes[id]; ok {
				if g.tree.isFiltered {
					return node.FilteredData
				}
				return node.Data
			}
			return []string{}
		},
		// IsBranch
		func(id string) bool {
			if id == "" {
				return true
			} else if node, ok := g.tree.nodes[id]; ok {
				return node.IsBranch
			}
			return false
		},
		// CreateNode
		func(branch bool) fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Label"),
				widget.NewLabel("Info"),
			)
		},
		// UpdateNode
		func(id string, branch bool, obj fyne.CanvasObject) {
			if id == "" {
				return
			}
			box := obj.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)
			info := box.Objects[2].(*widget.Label)

			node, ok := g.tree.nodes[id]
			if !ok {
				return
			}

			// Set icon
			res := theme.FileIcon()
			if node.IconName != "" {
				if cached, ok := g.iconCache[node.IconName]; ok {
					res = cached
				} else {
					svg, err := assets.GetIcon(node.IconName)
					if err == nil {
						res = fyne.NewStaticResource(node.IconName+".svg", svg)
						g.iconCache[node.IconName] = res
					}
				}
			}

			if branch {
				if node.IconName == "" {
					icon.SetResource(theme.DocumentIcon())
				} else {
					icon.SetResource(res)
				}
			} else {
				icon.SetResource(res)
			}
			label.SetText(node.Name)
			info.SetText(node.Info)
		},
	)

	for _, nodeId := range g.tree.roots {
		node := g.tree.nodes[nodeId]
		if node.IsBranch {
			g.listTree.OpenBranch(node.ID)
		}
	}

	// Buttons
	g.btnFolder = widget.NewButton(locale.T("Select Folder"), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			g.currentPath = uri.Path()
			g.pathEntry.SetText(g.currentPath)
		}, g.window)
	})

	g.btnAnalyze = widget.NewButton(locale.T("Analyze"), func() {
		path := g.pathEntry.Text
		if path == "" {
			g.statusLabel.SetText(locale.T("Please select or enter a folder path"))
			return
		}
		g.loadDirectory(path)
	})

	g.btnCollapseAll = widget.NewButton(locale.T("Collapse All"), func() {
		for id, node := range g.tree.nodes {
			if node.IsBranch {
				g.listTree.CloseBranch(id)
			}
		}
	})

	g.btnExpandAll = widget.NewButton(locale.T("Expand All"), func() {
		for id, node := range g.tree.nodes {
			if node.IsBranch {
				g.listTree.OpenBranch(id)
			}
		}
	})

	// Locale dropdown
	locales := locale.GetLocales()
	localeOptions := make([]string, len(locales))
	localeCodes := make([]string, len(locales))
	selectedIdx := 0
	for i, loc := range locales {
		localeOptions[i] = loc.Description
		localeCodes[i] = loc.Code
		if loc.Code == locale.Locale() {
			selectedIdx = i
		}
	}

	g.localeSelect = widget.NewSelect(localeOptions, func(selected string) {
		for i, opt := range localeOptions {
			if opt == selected {
				locale.SetLocale(localeCodes[i])
				g.refreshUI()
				// Re-analyze if data is loaded
				if g.currentPath != "" {
					g.loadDirectory(g.currentPath)
				}
				break
			}
		}
	})
	g.filterLabel = widget.NewLabel(locale.T("Filter:"))
	g.localeSelect.SetSelectedIndex(selectedIdx)

	topRow := container.NewBorder(
		nil, nil,
		g.btnFolder,
		container.NewHBox(g.btnAnalyze, g.localeSelect),
		g.pathEntry,
	)

	controlRow := container.NewBorder(
		nil, nil,
		container.NewHBox(g.btnCollapseAll, g.btnExpandAll),
		nil,
		container.NewBorder(nil, nil, g.filterLabel, nil, g.filterEntry),
	)

	header := container.NewVBox(
		topRow,
		controlRow,
	)

	// Wrap the tree with TreeTheme to enforce custom hover/selection colors
	treeContainer := container.NewThemeOverride(g.listTree, &TreeTheme{Theme: &AppTheme{}})

	content := container.NewBorder(
		header,
		g.statusLabel,
		nil, nil,
		treeContainer,
	)

	g.window.SetContent(content)
	g.app.Settings().SetTheme(&AppTheme{})
}

func (g *GUI) refreshUI() {
	g.window.SetTitle(locale.T("Photo Metadata Viewer"))
	g.statusLabel.SetText(locale.T("Ready"))
	g.pathEntry.SetPlaceHolder(locale.T("Select a folder or enter path..."))
	g.filterEntry.SetPlaceHolder(locale.T("Filter by property name..."))
	g.btnFolder.SetText(locale.T("Select Folder"))
	g.btnAnalyze.SetText(locale.T("Analyze"))
	g.btnCollapseAll.SetText(locale.T("Collapse All"))
	g.btnExpandAll.SetText(locale.T("Expand All"))
	g.filterLabel.SetText(locale.T("Filter:"))
}

func (g *GUI) loadDirectory(path string) {
	g.statusLabel.SetText(locale.T("Scanning ") + path + "...")
	g.currentPath = path

	go func() {
		imgs, err := g.service.ScanDirectory(context.Background(), path)

		fyne.Do(func() {
			if err != nil {
				g.statusLabel.SetText(locale.T("Error: ") + err.Error())
				return
			}

			g.updateData(imgs)
			g.statusLabel.SetText(fmt.Sprintf(locale.T("Found %d images"), len(imgs)))
		})
	}()
}

func (g *GUI) updateData(imgs []domain.ImageFile) {
	// Rebuild tree structure
	g.tree.Clear()

	for _, img := range imgs {
		fileID := img.Name
		props := g.buildProperties(img)

		// Create Node for File
		node := &TreeNode{
			ID:       fileID,
			IsBranch: true,
			Count:    len(props),
			Level:    0,
			Expanded: false,
			Name:     fileID,
			Info:     fmt.Sprintf("%d "+locale.T("properties"), len(props)),
			Data:     []string{},
			IconName: "format_" + img.Metadata.Format,
		}

		// Create Nodes for Properties
		for _, p := range props {
			propID := fileID + "::" + p.Name
			node.Data = append(node.Data, propID)

			iconName := "group_other"
			switch strings.ToLower(p.Group) {
			case "file":
				iconName = "group_file"
			case "shooting":
				iconName = "group_shooting"
			case "photo":
				iconName = "group_photo"
			case "date/location", "location", "datetime":
				iconName = "group_location"
			case "equipment":
				iconName = "group_equipment"
			case "author":
				iconName = "group_author"
			}

			g.tree.AddNode(&TreeNode{
				ID:       propID,
				ParentID: fileID,
				IsBranch: false,
				Level:    1,
				Name:     "(" + p.TagType + ") " + p.Name,
				Info:     "",
				OrigKey:  p.OrigKey,
				IconName: iconName,
			})
		}
		g.tree.AddNode(node)
	}

	g.tree.SortRoots()

	g.listTree.Refresh()
}

// dateFormat is the unified date format for all displayed dates.
const dateFormat = "02 Jan 2006 15:04:05"

type Property struct {
	Group   string
	TagType string
	OrigKey string
	Name    string
}

// addProp appends a formatted "key: value" property to the slice.
// Zero values of each type are skipped: nil, empty strings, zero numbers, zero dates.
// Non-zero values are formatted using standard format strings for their types.
func addProp(props []Property, group, tagType, origKey string, value any) []Property {
	if value == nil {
		return props
	}

	var s string
	switch val := value.(type) {
	case string:
		if val == "" || val == "<nil>" {
			return props
		}
		s = val
	case int:
		if val == 0 {
			return props
		}
		s = fmt.Sprintf("%d", val)
	case int64:
		if val == 0 {
			return props
		}
		s = fmt.Sprintf("%d", val)
	case float64:
		if val == 0 {
			return props
		}
		s = fmt.Sprintf("%.1f", val)
	case time.Time:
		if val.IsZero() {
			return props
		}
		s = val.Format(dateFormat)
	case *time.Time:
		if val == nil || val.IsZero() {
			return props
		}
		s = val.Format(dateFormat)
	default:
		s = fmt.Sprintf("%v", val)
		if s == "" || s == "<nil>" {
			return props
		}
	}

	return append(props, Property{
		Group:   group,
		TagType: tagType,
		OrigKey: origKey,
		Name:    fmt.Sprintf("%s: %s", locale.T(origKey), s),
	})
}

func (g *GUI) buildProperties(img domain.ImageFile) []Property {
	var props []Property

	props = addProp(props, "File", "File", "Size", fmt.Sprintf("%d "+locale.T("bytes"), img.Metadata.FileSize))
	props = addProp(props, "File", "File", "Format", strings.ToUpper(img.Metadata.Format))

	for _, t := range img.Metadata.Tags {
		props = addProp(props, t.Group, t.Type, t.Name, t.Value)
	}

	return props
}
