package fonts

import (
	"testing"
)

func TestFontManager(t *testing.T) {
	mgr := NewManager()
	fonts := mgr.ListFonts()
	t.Logf("Discovered %d fonts", len(fonts))
	for _, f := range fonts {
		t.Logf(" - [%s] %s (id: %s)", f.Category, f.Name, f.ID)
	}

	if len(fonts) == 0 {
		t.Log("No system fonts found automatically, will require upload/fallback")
		return
	}

	f, err := mgr.GetFont(fonts[0].ID)
	if err != nil {
		t.Fatalf("Failed to retrieve font %s: %v", fonts[0].ID, err)
	}
	t.Logf("Retrieved font %s, UnitsPerEm: %d", f.FontName, f.UnitsPerEm)
}
