package fonts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/jaredwarren/Text2SVG/pkg/converter"
)

// FontInfo describes an available font.
type FontInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"` // "System", "Uploaded", "Local"
}

// Manager handles font storage, scanning, and retrieval.
type Manager struct {
	mu           sync.RWMutex
	fontData     map[string][]byte
	fontCache    map[string]*converter.LoadedFont
	fontMetadata map[string]FontInfo
}

// NewManager creates a new FontManager and scans available system/local fonts.
func NewManager() *Manager {
	m := &Manager{
		fontData:     make(map[string][]byte),
		fontCache:    make(map[string]*converter.LoadedFont),
		fontMetadata: make(map[string]FontInfo),
	}
	m.scanInitialFonts()
	return m
}

// scanInitialFonts scans local ./fonts folder and system fonts.
func (m *Manager) scanInitialFonts() {
	// 1. Check local ./fonts directory
	if entries, err := os.ReadDir("./fonts"); err == nil {
		for _, e := range entries {
			if !e.IsDir() && isFontFile(e.Name()) {
				path := filepath.Join("./fonts", e.Name())
				if data, err := os.ReadFile(path); err == nil {
					name := cleanFontName(e.Name())
					m.RegisterFont(name, data, "Local")
				}
			}
		}
	}

	// 2. Check common system font directories
	searchDirs := []string{
		"/System/Library/Fonts/Supplemental",
		"/Library/Fonts",
		"/System/Library/Fonts",
		"C:\\Windows\\Fonts",
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	}

	// Well-known popular starter fonts to check first
	priorityFonts := []string{
		"Arial.ttf",
		"Helvetica.ttf",
		"Times New Roman.ttf",
		"Courier New.ttf",
		"Verdana.ttf",
		"Georgia.ttf",
		"Trebuchet MS.ttf",
		"Impact.ttf",
		"Apple Chancery.ttf",
		"Brush Script.ttf",
	}

	for _, dir := range searchDirs {
		for _, pf := range priorityFonts {
			path := filepath.Join(dir, pf)
			if _, ok := m.fontMetadata[cleanFontName(pf)]; !ok {
				if data, err := os.ReadFile(path); err == nil {
					m.RegisterFont(cleanFontName(pf), data, "System")
				}
			}
		}
	}

	// Also scan any other TTF/OTF fonts in those directories up to a reasonable count
	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if len(m.fontMetadata) >= 40 {
				break
			}
			if !e.IsDir() && isFontFile(e.Name()) {
				name := cleanFontName(e.Name())
				if _, exists := m.fontMetadata[name]; !exists {
					path := filepath.Join(dir, e.Name())
					if data, err := os.ReadFile(path); err == nil {
						m.RegisterFont(name, data, "System")
					}
				}
			}
		}
	}
}

func isFontFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".ttf" || ext == ".otf" || ext == ".woff"
}

func cleanFontName(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return strings.TrimSpace(name)
}

// RegisterFont stores a font into memory.
func (m *Manager) RegisterFont(name string, data []byte, category string) (*converter.LoadedFont, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	loaded, err := converter.ParseFont(data, name)
	if err != nil {
		return nil, fmt.Errorf("failed to register font %s: %w", name, err)
	}

	id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	m.fontData[id] = data
	m.fontCache[id] = loaded
	m.fontMetadata[id] = FontInfo{
		ID:       id,
		Name:     name,
		Category: category,
	}

	return loaded, nil
}

// GetFont retrieves a loaded font by ID or name.
func (m *Manager) GetFont(idOrName string) (*converter.LoadedFont, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lookup := strings.ToLower(strings.ReplaceAll(idOrName, " ", "-"))
	if lf, ok := m.fontCache[lookup]; ok {
		return lf, nil
	}

	// Try fuzzy matching against metadata
	for id, meta := range m.fontMetadata {
		if strings.EqualFold(meta.Name, idOrName) || strings.EqualFold(id, lookup) {
			return m.fontCache[id], nil
		}
	}

	// If none found, return the first available font
	for _, lf := range m.fontCache {
		return lf, nil
	}

	return nil, fmt.Errorf("font '%s' not found and no default fonts loaded", idOrName)
}

// ListFonts returns sorted list of all available fonts.
func (m *Manager) ListFonts() []FontInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []FontInfo
	for _, info := range m.fontMetadata {
		list = append(list, info)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Category != list[j].Category {
			return list[i].Category < list[j].Category
		}
		return list[i].Name < list[j].Name
	})

	return list
}
