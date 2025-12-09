package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/matzehuels/stacktower/pkg/dag"
	pkgio "github.com/matzehuels/stacktower/pkg/io"
	"github.com/matzehuels/stacktower/pkg/source"
	"github.com/matzehuels/stacktower/pkg/source/javascript"
	"github.com/matzehuels/stacktower/pkg/source/php"
	"github.com/matzehuels/stacktower/pkg/source/python"
	"github.com/matzehuels/stacktower/pkg/source/ruby"
	"github.com/matzehuels/stacktower/pkg/source/rust"
	"github.com/spf13/cobra"
)

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run a web server to serve dependency graphs.",
		Long:  `Starts a web server that provides an API to generate and view dependency graphs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context())
		},
	}

	return cmd
}

func runServer(ctx context.Context) error {
	http.Handle("/api/dependencies", dependenciesHandler(ctx))
	http.HandleFunc("/api/render", renderHandler)

	// Redirect root to dependencies.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dependencies.html", http.StatusMovedPermanently)
			return
		}
		// Fallthrough to the file server for other static assets
		http.FileServer(http.Dir("blogpost")).ServeHTTP(w, r)
	})

	fmt.Println("Starting server on http://localhost:8080")
	fmt.Println("Serving UI from the 'blogpost' directory.")
	return http.ListenAndServe(":8080", nil)
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	tmpfile, err := os.CreateTemp("blogpost/tmp", "render-*.json")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		http.Error(w, "Error creating temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(body); err != nil {
		log.Printf("Error writing to temp file: %v", err)
		http.Error(w, "Error writing to temp file", http.StatusInternalServerError)
		return
	}
	if err := tmpfile.Close(); err != nil {
		log.Printf("Error closing temp file: %v", err)
		http.Error(w, "Error closing temp file", http.StatusInternalServerError)
		return
	}

	outputFile := tmpfile.Name() + ".svg"
	defer os.Remove(outputFile)

	executable, err := os.Executable()
	if err != nil {
		log.Printf("Error finding executable: %v", err)
		http.Error(w, "Error finding executable", http.StatusInternalServerError)
		return
	}

	cmd := exec.Command(
		executable,
		"render",
		tmpfile.Name(),
		"-t", "tower",
		"--style", "handdrawn",
		"--width", "982",
		"--height", "500",
		"--ordering", "optimal",
		"--merge",
		"--randomize",
		"-o", outputFile,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Error running stacktower render: %v\n%s", err, stderr.String())
		http.Error(w, "Error running stacktower render", http.StatusInternalServerError)
		return
	}

	svgData, err := os.ReadFile(outputFile)
	if err != nil {
		log.Printf("Error reading svg file: %v", err)
		http.Error(w, "Error reading svg file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(svgData)
}

var parserFactories = map[string]func() (source.Parser, error){
	"pypi":       func() (source.Parser, error) { return python.NewParser(source.DefaultCacheTTL) },
	"crates":     func() (source.Parser, error) { return rust.NewParser(source.DefaultCacheTTL) },
	"npm":        func() (source.Parser, error) { return javascript.NewParser(source.DefaultCacheTTL) },
	"rubygems":   func() (source.Parser, error) { return ruby.NewParser(source.DefaultCacheTTL) },
	"packagist":  func() (source.Parser, error) { return php.NewParser(source.DefaultCacheTTL) },
	// "github" would need a different handling as it's not a simple package parser
}


func dependenciesHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceType := r.URL.Query().Get("source")
		pkgName := r.URL.Query().Get("id")

		if sourceType == "" || pkgName == "" {
			http.Error(w, "Missing 'source' or 'id' query parameter", http.StatusBadRequest)
			return
		}

		factory, ok := parserFactories[sourceType]
		if !ok {
			http.Error(w, fmt.Sprintf("Source '%s' not supported", sourceType), http.StatusBadRequest)
			return
		}

		p, err := factory()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating parser: %v", err), http.StatusInternalServerError)
			return
		}

		// Simplified opts for now
		opts := &parseOpts{maxDepth: 10, maxNodes: 500}

		graph, err := runParseForServer(r.Context(), p, pkgName, opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing dependencies: %v", err), http.StatusInternalServerError)
			return
		}
		
		// Now we will render as json
		w.Header().Set("Content-Type", "application/json")
		if err := pkgio.WriteJSON(graph, w); err != nil {
			http.Error(w, fmt.Sprintf("Error writing json output: %v", err), http.StatusInternalServerError)
			return
		}
	})
}


func runParseForServer(ctx context.Context, p source.Parser, pkg string, opts *parseOpts) (*dag.DAG, error) {
	// This function is an adaptation of runParse from parse.go
	// We can't use the logger from the command context here easily, so we use a default one for now.
	
	// No metadata providers for now to keep it simple
	var providers []source.MetadataProvider

	srcOpts := source.Options{
		MaxDepth:          opts.maxDepth,
		MaxNodes:          opts.maxNodes,
		MetadataProviders: providers,
		Refresh:           opts.refresh,
		CacheTTL:          source.DefaultCacheTTL,
	}

	g, err := p.Parse(ctx, pkg, srcOpts)
	if err != nil {
		return nil, err
	}

	return g, nil
}
