package pipeline

import (
	"io/fs"
	"log"
	"path/filepath"
	"sync"

	"github.com/djblackett/bootdev-hackathon/ai"
	"github.com/djblackett/bootdev-hackathon/heuristics"
	"github.com/djblackett/bootdev-hackathon/utils"
)

type Job struct {
	Path     string
	Content  string
	Filename string // heuristic or final
}

// ProcessFiles processes files in the given directory using a worker pool pattern
func ProcessFiles(dir string, N, M int, dryRun bool, aiClient ai.Client, extract func(string) string) {
	filesCh := make(chan string) // paths
	aiCh := make(chan Job)       // low‑conf jobs
	resultsCh := make(chan Job)  // final filenames

	wgExt, wgAI := &sync.WaitGroup{}, &sync.WaitGroup{}

	// 1. discovery goroutine
	go func() {
		filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				filesCh <- p
			}
			return nil
		})
		close(filesCh)
	}()

	// 2. extractor + heuristic workers
	for i := 0; i < N; i++ {
		wgExt.Add(1)
		go func() {
			defer wgExt.Done()
			for path := range filesCh {
				content := extract(path) // pdf/txt/etc.
				name, conf := heuristics.Score(content)
				if conf >= 70 {
					resultsCh <- Job{Path: path, Filename: name}
				} else {
					aiCh <- Job{Path: path, Content: content}
				}
			}
		}()
	}

	// 3. close aiCh after extractors finish
	go func() {
		wgExt.Wait()
		close(aiCh)
	}()

	// 4. AI worker pool (rate‑limited)
	limiter := make(chan struct{}, M) // M concurrent AI calls
	for j := 0; j < M; j++ {
		wgAI.Add(1)
		go func() {
			defer wgAI.Done()
			for job := range aiCh {
				limiter <- struct{}{} // back‑pressure
				name, _ := aiClient.SuggestFilename(job.Content)
				<-limiter
				resultsCh <- Job{Path: job.Path, Filename: utils.Sanitize(name)}
			}
		}()
	}

	// 5. close results when AI done
	go func() {
		wgAI.Wait()
		close(resultsCh)
	}()

	// 6. consumer - rename or dry‑run
	for job := range resultsCh {
		if dryRun {
			log.Printf("[DRY] %s → %s\n", job.Path, job.Filename)
		} else {
			utils.RenameFile(job.Path, job.Filename)
		}
	}
}
