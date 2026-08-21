package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Request struct {
	ServiceName     string `json:"service_name"`
	Tag             string `json:"tag"`
	ComposeFilename string `json:"compose_filename"`
}

func isValidServiceName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

func isValidTag(tag string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, tag)
	return matched
}

func backupFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	backupPath := path + ".backup"
	return os.WriteFile(backupPath, data, 0644)
}

func updateImageTag(composePath string, newTag string) error {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return err
	}

	err = backupFile(composePath)
	if err != nil {
		fmt.Printf("Warning: could not create backup: %s\n", err)
	}

	lines := strings.Split(string(data), "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "image:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				imageValue := strings.TrimSpace(parts[1])
				imageParts := strings.Split(imageValue, ":")
				if len(imageParts) > 1 {
					newImageValue := imageParts[0] + ":" + newTag
					indent := strings.Repeat(" ", len(line)-len(trimmed))
					result = append(result, indent+"image: "+newImageValue)
				} else {
					newImageValue := imageValue + ":" + newTag
					indent := strings.Repeat(" ", len(line)-len(trimmed))
					result = append(result, indent+"image: "+newImageValue)
				}
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}

	newData := strings.Join(result, "\n")
	return os.WriteFile(composePath, []byte(newData), 0644)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	if !isValidServiceName(req.ServiceName) {
		http.Error(w, "Invalid service name. Only alphanumeric, underscore, and dash allowed", http.StatusBadRequest)
		return
	}

	if req.Tag == "" {
		http.Error(w, "Tag is required", http.StatusBadRequest)
		return
	}

	if !isValidTag(req.Tag) {
		http.Error(w, "Invalid tag. Only alphanumeric, dot, underscore, and dash allowed", http.StatusBadRequest)
		return
	}

	baseFolder := os.Getenv("BASE_FOLDER")
	if baseFolder == "" {
		baseFolder = "/home/ajii/manifest"
	}

	servicePath := filepath.Join(baseFolder, req.ServiceName)

	var composePath string
	if req.ComposeFilename != "" {
		composePath = filepath.Join(servicePath, req.ComposeFilename)
	} else {
		composePath = filepath.Join(servicePath, "docker-compose.yaml")
	}

	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		http.Error(w, "Service or docker-compose file not found", http.StatusNotFound)
		return
	}

	err = updateImageTag(composePath, req.Tag)
	if err != nil {
		http.Error(w, "Error updating image tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var output []byte
	var execErr error

	fmt.Printf("Trying: docker compose -f %s up -d\n", composePath)
	cmd1 := exec.Command("sh", "-c", "docker compose -f "+composePath+" up -d")
	cmd1.Dir = servicePath
	output1, err1 := cmd1.CombinedOutput()
	fmt.Printf("Output: %s\n", string(output1))

	if err1 == nil {
		output = output1
		execErr = err1
		fmt.Println("✓ Success with docker compose -f")
	} else {
		fmt.Printf("✗ Failed, trying: docker-compose -f %s up -d\n", composePath)
		cmd2 := exec.Command("sh", "-c", "docker-compose -f "+composePath+" up -d")
		cmd2.Dir = servicePath
		output2, err2 := cmd2.CombinedOutput()
		fmt.Printf("Output: %s\n", string(output2))

		if err2 == nil {
			output = output2
			execErr = err2
			fmt.Println("✓ Success with docker-compose -f")
		} else {
			fmt.Printf("✗ Failed, trying: docker compose up -d in %s\n", servicePath)
			cmd3 := exec.Command("sh", "-c", "docker compose up -d")
			cmd3.Dir = servicePath
			output, execErr = cmd3.CombinedOutput()
			fmt.Printf("Output: %s\n", string(output))

			if execErr == nil {
				fmt.Println("✓ Success with docker compose up -d")
			} else {
				fmt.Println("✗ All commands failed")
				http.Error(w, "Error running docker compose: "+execErr.Error()+"\n"+string(output), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Service deployed",
		"output":  string(output),
	})
}

func main() {
	http.HandleFunc("/deploy", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
