package entities

import (
	"path/filepath"
	"strings"
)

type Language struct {
	Name       string
	Color      string
	Extensions []string
}

var LanguageCatalog = map[string]Language{
	"go": {
		Name:       "Go",
		Color:      "80",
		Extensions: []string{".go", ".mod", ".sum", ".tmpl", ".gohtml"},
	},
	"python": {
		Name:       "Python",
		Color:      "68",
		Extensions: []string{".py", ".pyw", ".pyi"},
	},
	"rust": {
		Name:       "Rust",
		Color:      "180",
		Extensions: []string{".rs"},
	},
	"javascript": {
		Name:       "JavaScript",
		Color:      "220",
		Extensions: []string{".js", ".mjs", ".cjs", ".jsx"},
	},
	"typescript": {
		Name:       "TypeScript",
		Color:      "33",
		Extensions: []string{".ts", ".tsx", ".mts", ".cts"},
	},
	"java": {
		Name:       "Java",
		Color:      "136",
		Extensions: []string{".java"},
	},
	"cpp": {
		Name:       "C++",
		Color:      "205",
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx", ".cppm"},
	},
	"c": {
		Name:       "C",
		Color:      "75",
		Extensions: []string{".c", ".h"},
	},
	"csharp": {
		Name:       "C#",
		Color:      "34",
		Extensions: []string{".cs", ".csx"},
	},
	"ruby": {
		Name:       "Ruby",
		Color:      "124",
		Extensions: []string{".rb", ".rake", ".gemspec", ".rbw"},
	},
	"php": {
		Name:       "PHP",
		Color:      "97",
		Extensions: []string{".php", ".phtml", ".phps"},
	},
	"swift": {
		Name:       "Swift",
		Color:      "202",
		Extensions: []string{".swift"},
	},
	"kotlin": {
		Name:       "Kotlin",
		Color:      "141",
		Extensions: []string{".kt", ".kts"},
	},
	"scala": {
		Name:       "Scala",
		Color:      "161",
		Extensions: []string{".scala", ".sc", ".sbt"},
	},
	"dart": {
		Name:       "Dart",
		Color:      "43",
		Extensions: []string{".dart"},
	},
	"lua": {
		Name:       "Lua",
		Color:      "26",
		Extensions: []string{".lua"},
	},
	"haskell": {
		Name:       "Haskell",
		Color:      "97",
		Extensions: []string{".hs", ".lhs"},
	},
	"elixir": {
		Name:       "Elixir",
		Color:      "97",
		Extensions: []string{".ex", ".exs"},
	},
	"erlang": {
		Name:       "Erlang",
		Color:      "160",
		Extensions: []string{".erl", ".hrl"},
	},
	"clojure": {
		Name:       "Clojure",
		Color:      "42",
		Extensions: []string{".clj", ".cljs", ".cljc", ".edn"},
	},
	"julia": {
		Name:       "Julia",
		Color:      "133",
		Extensions: []string{".jl"},
	},
	"r": {
		Name:       "R",
		Color:      "37",
		Extensions: []string{".r", ".R", ".rmd"},
	},
	"matlab": {
		Name:       "MATLAB",
		Color:      "130",
		Extensions: []string{".m"},
	},
	"zig": {
		Name:       "Zig",
		Color:      "178",
		Extensions: []string{".zig"},
	},
	"nim": {
		Name:       "Nim",
		Color:      "220",
		Extensions: []string{".nim", ".nims"},
	},
	"ocaml": {
		Name:       "OCaml",
		Color:      "163",
		Extensions: []string{".ml", ".mli"},
	},
	"fsharp": {
		Name:       "F#",
		Color:      "74",
		Extensions: []string{".fs", ".fsx", ".fsi"},
	},
	"assembly": {
		Name:       "Assembly",
		Color:      "105",
		Extensions: []string{".asm", ".s", ".S"},
	},
	"shell": {
		Name:       "Shell",
		Color:      "149",
		Extensions: []string{".sh", ".bash", ".zsh", ".fish"},
	},
	"powershell": {
		Name:       "PowerShell",
		Color:      "74",
		Extensions: []string{".ps1", ".psm1", ".psd1"},
	},
	"html": {
		Name:       "HTML",
		Color:      "196",
		Extensions: []string{".html", ".htm", ".xhtml"},
	},
	"css": {
		Name:       "CSS",
		Color:      "61",
		Extensions: []string{".css", ".scss", ".sass", ".less"},
	},
	"json": {
		Name:       "JSON",
		Color:      "220",
		Extensions: []string{".json"},
	},
	"yaml": {
		Name:       "YAML",
		Color:      "160",
		Extensions: []string{".yaml", ".yml"},
	},
	"toml": {
		Name:       "TOML",
		Color:      "136",
		Extensions: []string{".toml"},
	},
	"xml": {
		Name:       "XML",
		Color:      "131",
		Extensions: []string{".xml"},
	},
	"markdown": {
		Name:       "Markdown",
		Color:      "255",
		Extensions: []string{".md", ".mdx", ".markdown"},
	},
	"dockerfile": {
		Name:       "Dockerfile",
		Color:      "33",
		Extensions: []string{"Dockerfile", ".dockerfile"},
	},
	"terraform": {
		Name:       "Terraform",
		Color:      "99",
		Extensions: []string{".tf", ".tfvars"},
	},
	"protobuf": {
		Name:       "Protobuf",
		Color:      "196",
		Extensions: []string{".proto"},
	},
	"graphql": {
		Name:       "GraphQL",
		Color:      "200",
		Extensions: []string{".graphql", ".gql"},
	},
	"vue": {
		Name:       "Vue",
		Color:      "42",
		Extensions: []string{".vue"},
	},
	"svelte": {
		Name:       "Svelte",
		Color:      "196",
		Extensions: []string{".svelte"},
	},
	"astro": {
		Name:       "Astro",
		Color:      "202",
		Extensions: []string{".astro"},
	},
}

var defaultLanguage = Language{
	Name:       "Unknown",
	Color:      "6",
	Extensions: nil,
}

func DetectLanguage(files []string) Language {
	counts := make(map[string]int)

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		base := filepath.Base(file)

		for id, lang := range LanguageCatalog {
			for _, le := range lang.Extensions {
				if strings.ToLower(le) == ext || strings.ToLower(le) == base {
					counts[id]++
				}
			}
		}
	}

	bestID := ""
	bestCount := 0
	for id, count := range counts {
		if count > bestCount {
			bestID = id
			bestCount = count
		}
	}

	if bestID == "" {
		return defaultLanguage
	}

	lang := LanguageCatalog[bestID]
	return lang
}
