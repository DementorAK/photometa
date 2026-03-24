package gui

import (
	"sort"
	"strings"
)

type TreeNode struct {
	ID           string
	ParentID     string
	IsBranch     bool
	Level        int
	Expanded     bool
	Name         string
	Info         string
	Count        int      // For branches: property count
	Data         []string // For branches: children IDs
	FilteredData []string
	OrigKey      string
	IconName     string
}

type TreeModel struct {
	nodes         map[string]*TreeNode
	roots         []string
	filteredRoots []string
	isFiltered    bool
}

func NewTreeModel() *TreeModel {
	return &TreeModel{
		nodes: make(map[string]*TreeNode),
		roots: []string{},
	}
}

func (t *TreeModel) AddNode(node *TreeNode) {
	t.nodes[node.ID] = node
	if node.ParentID == "" {
		t.roots = append(t.roots, node.ID)
	}
}

func (t *TreeModel) Clear() {
	t.nodes = make(map[string]*TreeNode)
	t.roots = []string{}
	t.filteredRoots = []string{}
	t.isFiltered = false
}

func (t *TreeModel) SortRoots() {
	sort.Strings(t.roots)
}

func (t *TreeModel) Filter(text string) {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		t.isFiltered = false
		t.filteredRoots = []string{}
		return
	}

	t.isFiltered = true
	t.filteredRoots = []string{}

	for _, rootID := range t.roots {
		node := t.nodes[rootID]
		var matchingChildren []string

		for _, childID := range node.Data {
			childNode, ok := t.nodes[childID]
			if !ok {
				continue
			}

			// Match by Original Key (English) OR localized Name
			origLower := strings.ToLower(childNode.OrigKey)
			nameLower := strings.ToLower(childNode.Name)

			if strings.Contains(origLower, text) || strings.Contains(nameLower, text) {
				matchingChildren = append(matchingChildren, childID)
			}
		}

		if len(matchingChildren) > 0 {
			t.filteredRoots = append(t.filteredRoots, rootID)
			node.FilteredData = matchingChildren
		}
	}
}
