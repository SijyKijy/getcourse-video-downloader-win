package combiner

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
)

func CombineFiles(outputFile string) error {
    listFile, err := os.Create("ffmpeg_list_all.txt")
    if err != nil {
        return fmt.Errorf("error creating list file: %v", err)
    }
    defer func(name string) {
        err := os.Remove(name)
        if err != nil {
            log.Printf("Error removing file: %v", err)
        }
    }(listFile.Name())

    tsFiles, _ := filepath.Glob("parts/*.ts")
    binFiles, _ := filepath.Glob("parts/*.bin")
    files := append(tsFiles, binFiles...)

    sort.Strings(files)

    for _, file := range files {
        if _, err := os.Stat(file); os.IsNotExist(err) {
            return fmt.Errorf("file not exist: %s", file)
        }
    }

    for _, file := range files {
        unifiedPath := filepath.ToSlash(file)
        _, err := listFile.WriteString(fmt.Sprintf("file '%s'\n", unifiedPath))
        if err != nil {
            return fmt.Errorf("error writing to list file: %v", err)
        }
    }
    err = listFile.Close()
    if err != nil {
        log.Printf("Error closing file: %v", err)
    }

    cmd := exec.Command("ffmpeg",
        "-f", "concat",
        "-safe", "0",
        "-i", listFile.Name(),
        "-c", "copy",
        outputFile)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ffmpeg error: %v\nOutput: %s", err, string(output))
    }

    return nil
}
