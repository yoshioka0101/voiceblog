package archtest

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/yoshioka0101/voiceblog/backend"

// layerRules はクリーンアーキテクチャの依存規則。
// キーのレイヤー配下のパッケージは、値に列挙した import プレフィックスに依存してはならない。
var layerRules = map[string][]string{
	"internal/entity": {
		modulePath + "/internal/usecase",
		modulePath + "/internal/handler",
		modulePath + "/internal/presenter",
		modulePath + "/internal/infra",
		modulePath + "/internal/repository",
		modulePath + "/internal/di",
		modulePath + "/internal/server",
		modulePath + "/internal/db",
		modulePath + "/internal/api",
		modulePath + "/models",
	},
	"internal/usecase": {
		modulePath + "/internal/handler",
		modulePath + "/internal/presenter",
		modulePath + "/internal/infra",
		modulePath + "/internal/repository",
		modulePath + "/internal/di",
		modulePath + "/internal/server",
		modulePath + "/internal/api",
		modulePath + "/models",
	},
	"internal/presenter": {
		modulePath + "/internal/handler",
		modulePath + "/internal/infra",
		modulePath + "/internal/repository",
		modulePath + "/internal/di",
		modulePath + "/internal/server",
		modulePath + "/models",
	},
	"internal/handler": {
		modulePath + "/internal/infra",
		modulePath + "/internal/repository",
		modulePath + "/models",
	},
}

func TestCleanArchitectureDependencyRules(t *testing.T) {
	backendRoot := filepath.Join("..", "..")

	for layer, forbidden := range layerRules {
		layerDir := filepath.Join(backendRoot, filepath.FromSlash(layer))
		if _, err := os.Stat(layerDir); err != nil {
			t.Fatalf("layer directory not found: %s", layerDir)
		}

		err := filepath.WalkDir(layerDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}

			for _, imp := range file.Imports {
				importPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return err
				}

				for _, prefix := range forbidden {
					if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
						t.Errorf("%s: %s は %s に依存できません（依存規則違反: %s）", layer, path, importPath, prefix)
					}
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", layer, err)
		}
	}
}
