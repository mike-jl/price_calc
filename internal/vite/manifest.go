package vite

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type ViteManifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

var (
	manifest     map[string]ViteManifestEntry
	manifestOnce sync.Once
	manifestErr  error
)

func ViteAssetPath(asset string) ViteManifestEntry {
	manifestOnce.Do(func() {
		data, err := os.ReadFile("assets/js/.vite/manifest.json")
		if err != nil {
			manifestErr = fmt.Errorf("failed to read manifest: %w", err)
			return
		}
		err = json.Unmarshal(data, &manifest)
		if err != nil {
			manifestErr = fmt.Errorf("failed to parse manifest: %w", err)
		}
	})

	if manifestErr != nil {
		panic(manifestErr) // or return a fallback / log / etc.
	}

	entry, ok := manifest[asset]
	if !ok {
		panic(fmt.Sprintf("asset %q not found in manifest", asset))
	}

	return entry
}
