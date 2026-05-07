package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var reloadStatic bool
var host string
var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the out/ folder and watch for changes",
	Run: func(cmd *cobra.Command, args []string) {
		// Initial build
		fmt.Println("Performing initial build...")
		statusCmd.Run(cmd, args)
		buildCmd.Run(cmd, args)

		// Setup watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		// SSE Setup
		clients := make(map[chan bool]bool)
		clientsMu := sync.Mutex{}

		notifyClients := func() {
			clientsMu.Lock()
			defer clientsMu.Unlock()

			fmt.Printf("Notifying %d clients...\n", len(clients))
			for client := range clients {
				select {
				case client <- true:
				default:
				}
			}
		}

		go func() {
			var timer *time.Timer
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					// Ignore Chmod events
					if event.Op&fsnotify.Chmod == fsnotify.Chmod {
						continue
					}

					// Filter for relevant files
					ext := filepath.Ext(event.Name)
					isNote := ext == ".note"
					isTemplate := false
					if reloadStatic {
						if strings.HasPrefix(event.Name, "tpl/") {
							isTemplate = true
						}
					}

					if !isNote && !isTemplate {
						// If it's a directory creation, watch it
						if event.Op&fsnotify.Create == fsnotify.Create {
							info, err := os.Stat(event.Name)
							if err == nil && info.IsDir() {
								addRecursive(watcher, event.Name)
							}
						}
						continue
					}

					// Rebuild logic with debouncing
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(200*time.Millisecond, func() {
						fmt.Printf("\n[%s] Change detected in %s. Rebuilding...\n", time.Now().Format("15:04:05"), event.Name)
						buildCmd.Run(cmd, args)
						notifyClients()
					})

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()

		// Watch src/ for .note files
		addRecursive(watcher, "src")

		if reloadStatic {
			addRecursive(watcher, "tpl")
		}

		// Start server
		addr := fmt.Sprintf("%s:%d", host, port)
		fmt.Printf("\nServing out/ on http://%s\n", addr)
		fmt.Println("Press Ctrl+C to stop")

		mux := http.NewServeMux()

		// SSE Endpoint
		mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			client := make(chan bool)
			clientsMu.Lock()
			clients[client] = true
			clientsMu.Unlock()

			defer func() {
				clientsMu.Lock()
				delete(clients, client)
				clientsMu.Unlock()
			}()

			for {
				select {
				case <-client:
					fmt.Fprintf(w, "data: reload\n\n")
					w.(http.Flusher).Flush()
				case <-r.Context().Done():
					return
				}
			}
		})

		// File Server with injection
		fileServer := http.FileServer(http.Dir("out"))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			ext := filepath.Ext(r.URL.Path)
			if ext == ".html" || ext == "" {
				path := filepath.Join("out", r.URL.Path)
				if ext == "" {
					path = filepath.Join(path, "index.html")
				}

				content, err := os.ReadFile(path)
				if err == nil {
					html := string(content)
					script := `
<script>
  const ev = new EventSource('/sse');
  ev.onmessage = (e) => {
    if (e.data === 'reload') {
      location.reload();
    }
  };
</script>
</body>`
					html = strings.Replace(html, "</body>", script, 1)
					w.Header().Set("Content-Type", "text/html")
					fmt.Fprint(w, html)
					return
				}
			}
			fileServer.ServeHTTP(w, r)
		})

		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	},
}

func init() {
	serveCmd.Flags().BoolVar(&reloadStatic, "reload-static", false, "Watch tpl/ folder as well")
	serveCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to serve on")
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve on")
}

func addRecursive(watcher *fsnotify.Watcher, path string) {
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(p)
		}
		return nil
	})
	if err != nil {
		log.Printf("error walking %s: %v", path, err)
	}
}
