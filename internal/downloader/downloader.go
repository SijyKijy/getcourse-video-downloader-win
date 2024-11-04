package downloader

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/schollz/progressbar/v3"
)

const (
	MaxConcurrentDownloads = 5    // Maximum number of concurrent downloads
	MaxRetryAttempts       = 3    // Maximum number of retry attempts on failure
	RetryDelay             = 2 * time.Second // Delay between retry attempts
)

var client = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

func DownloadPlaylist(playlistURL string) ([]string, error) {
    resp, err := client.Get(playlistURL)
    if err != nil {
        return nil, fmt.Errorf("error fetching playlist: %w", err)
    }
    defer func() {
        if err := resp.Body.Close(); err != nil {
            log.Printf("Error closing response body: %v", err)
        }
    }()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading playlist body: %w", err)
    }

    var playlist []string
    scanner := bufio.NewScanner(strings.NewReader(string(body)))
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "http") {
            playlist = append(playlist, line)
        } else if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
            baseURL, err := url.Parse(playlistURL)
            if err != nil {
                log.Printf("Error parsing base URL: %v", err)
                continue
            }
            relativeURL, err := url.Parse(line)
            if err != nil {
                log.Printf("Error parsing relative URL '%s': %v", line, err)
                continue
            }
            absoluteURL := baseURL.ResolveReference(relativeURL)
            playlist = append(playlist, absoluteURL.String())
        }
    }

    if len(playlist) == 0 {
        scanner = bufio.NewScanner(strings.NewReader(string(body)))
        var lastLine string
        for scanner.Scan() {
            lastLine = scanner.Text()
        }
        if strings.HasPrefix(lastLine, "http") {
            return DownloadPlaylist(lastLine)
        }
        return nil, fmt.Errorf("no valid URLs found in the playlist")
    }

    return playlist, scanner.Err()
}

func DownloadFiles(playlist []string) error {
    if len(playlist) < 2 {
        return fmt.Errorf("the playlist contains insufficient files to download")
    }

    bar := progressbar.NewOptions(len(playlist)-1,
        progressbar.OptionEnableColorCodes(true),
        progressbar.OptionShowCount(),
        progressbar.OptionSetWidth(50),
        progressbar.OptionSetDescription("Downloading"),
        progressbar.OptionSetWriter(os.Stdout),
        progressbar.OptionSetTheme(progressbar.Theme{
            Saucer:        "[green]=[reset]",
            SaucerHead:    "[green]>[reset]",
            SaucerPadding: " ",
            BarStart:      "[",
            BarEnd:        "]",
        }),
    )

    var wg sync.WaitGroup
    sem := make(chan struct{}, MaxConcurrentDownloads)
    errChan := make(chan error, len(playlist)-1)

    for i := 1; i < len(playlist); i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()

            sem <- struct{}{}
            defer func() { <-sem }()

            urlString := playlist[i]
            fileName := filepath.Join("parts", fmt.Sprintf("file_%05d.ts", i-1))

            var err error
            for attempt := 1; attempt <= MaxRetryAttempts; attempt++ {
                err = downloadFile(urlString, fileName)
                if err == nil {
                    break
                }
                log.Printf("Error downloading %s (attempt %d/%d): %v", urlString, attempt, MaxRetryAttempts, err)
                time.Sleep(RetryDelay)
            }
            if err != nil {
                errChan <- fmt.Errorf("error downloading file %d: %w", i, err)
                return
            }

            if err := bar.Add(1); err != nil {
                log.Printf("Error updating progress bar: %v", err)
            }
        }(i)
    }

    wg.Wait()
    close(errChan)

    var combinedError string
    for err := range errChan {
        combinedError += err.Error() + "\n"
    }
    if combinedError != "" {
        return fmt.Errorf("errors occurred while downloading files:\n%s", combinedError)
    }

    fmt.Println()
    return nil
}

func downloadFile(urlString, fileName string) error {
    resp, err := client.Get(urlString)
    if err != nil {
        return fmt.Errorf("error fetching URL %s: %w", urlString, err)
    }
    defer func() {
        if err := resp.Body.Close(); err != nil {
            log.Printf("Error closing response body for %s: %v", urlString, err)
        }
    }()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("non-200 HTTP status: %s", resp.Status)
    }

    out, err := os.Create(fileName)
    if err != nil {
        return fmt.Errorf("error creating file %s: %w", fileName, err)
    }
    defer func() {
        if err := out.Close(); err != nil {
            log.Printf("Error closing file %s: %v", fileName, err)
        }
    }()

    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return fmt.Errorf("error writing to file %s: %w", fileName, err)
    }

    return nil
}