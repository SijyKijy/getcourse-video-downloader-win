package combiner

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "runtime"
)

const (
    BatchSize = 500
)

func CombineFiles(outputFile string) error {
    tsFiles, err := filepath.Glob("parts/*.ts")
    if err != nil {
        return fmt.Errorf("error globbing ts files: %w", err)
    }
    binFiles, err := filepath.Glob("parts/*.bin")
    if err != nil {
        return fmt.Errorf("error globbing bin files: %w", err)
    }
    files := append(tsFiles, binFiles...)

    if len(files) == 0 {
        return fmt.Errorf("no .ts or .bin files found in parts directory")
    }

    sort.Strings(files)

    for _, file := range files {
        if _, err := os.Stat(file); os.IsNotExist(err) {
            return fmt.Errorf("file does not exist: %s", file)
        }
    }

    var batches [][]string
    for i := 0; i < len(files); i += BatchSize {
        end := i + BatchSize
        if end > len(files) {
            end = len(files)
        }
        batches = append(batches, files[i:end])
    }

    intermediateFiles := []string{}

    for idx, batch := range batches {
        listFileName := fmt.Sprintf("ffmpeg_list_batch_%d.txt", idx)
        listFile, err := os.Create(listFileName)
        if (err != nil) {
            return fmt.Errorf("error creating list file: %v", err)
        }

        for _, file := range batch {
            unifiedPath := filepath.ToSlash(file)
            _, err := listFile.WriteString(fmt.Sprintf("file '%s'\n", unifiedPath))
            if err != nil {
                listFile.Close()
                return fmt.Errorf("error writing to list file: %v", err)
            }
        }
        listFile.Close()

        intermediateFile := fmt.Sprintf("intermediate_%d.ts", idx)
        cmd := exec.Command("ffmpeg",
            "-f", "concat",
            "-safe", "0",
            "-i", listFileName,
            "-c", "copy",
            "-y",
            intermediateFile)

        output, err := cmd.CombinedOutput()
        if err != nil {
            return fmt.Errorf("ffmpeg error on batch %d: %v\nOutput: %s", idx, err, string(output))
        }

        intermediateFiles = append(intermediateFiles, intermediateFile)

        os.Remove(listFileName)
    }

    finalListFile, err := os.Create("ffmpeg_final_list.txt")
    if err != nil {
        return fmt.Errorf("error creating final list file: %v", err)
    }

    for _, file := range intermediateFiles {
        unifiedPath := filepath.ToSlash(file)
        _, err := finalListFile.WriteString(fmt.Sprintf("file '%s'\n", unifiedPath))
        if err != nil {
            finalListFile.Close()
            return fmt.Errorf("error writing to final list file: %v", err)
        }
    }
    finalListFile.Close()

    cmd := exec.Command("ffmpeg",
        "-f", "concat",
        "-safe", "0",
        "-i", "ffmpeg_final_list.txt",
        "-c", "copy",
        "-threads", fmt.Sprintf("%d", runtime.NumCPU()),
        "-y",
        outputFile)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ffmpeg final merge error: %v\nOutput: %s", err, string(output))
    }

    for _, file := range intermediateFiles {
        os.Remove(file)
    }
    os.Remove("ffmpeg_final_list.txt")

    return nil
}
