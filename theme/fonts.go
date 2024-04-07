package theme

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/text"
)

// LoadBuiltin loads builtin fonts from a provided dir and the Gio builtin Go font collection.
func LoadBuiltin(fontDir string) []font.FontFace {
	var fonts []font.FontFace
	fonts = append(fonts, gofont.Collection()...)

	entries, err := os.ReadDir(fontDir)
	if err != nil {
		log.Printf("loading fonts from dir failed: %w", err)
		return fonts
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if filepath.Ext(filename) != ".ttf" {
			continue
		}
		ttfData, err := os.ReadFile(filepath.Join(fontDir, filename))
		if err != nil {
			log.Printf("loading fonts from dir failed: %w", err)
			continue
		}

		face, err := loadFont(ttfData)
		if err != nil {
			panic(err)
		}
		fonts = append(fonts, *face)

	}

	for _, f := range fonts {
		log.Printf("loaded builtin font face: %s, style: %s, weight: %s", f.Font.Typeface, f.Font.Style, f.Font.Weight)
	}

	return fonts
}

func loadFont(ttf []byte) (*font.FontFace, error) {
	faces, err := opentype.ParseCollection(ttf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	return &text.FontFace{
		Font: faces[0].Font,
		Face: faces[0].Face,
	}, nil
}
