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

	"gopkg.in/yaml.v3"
)

type Request struct {
	ServiceName      string `json:"service_name"`
	Tag              string `json:"tag"`
	ComposeFilename  string `json:"compose_filename"`
}

type DockerCompose struct {
	Version string              `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Image string `yaml:"image"`
	Build interface{} `yaml:"build"`
}

func isValidServiceName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

func isValidTag(tag string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, tag)
	return matched
}

func updateImageTag(composePath string, newTag string) error {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return err
	}

	var compose DockerCompose
	err = yaml.Unmarshal(data, &compose)
	if err != nil {
		return err
	}

	for name, service := range compose.Services {
		if service.Image != "" {
			parts := strings.Split(service.Image, ":")
			if len(parts) > 1 {
				service.Image = parts[0] + ":" + newTag
			} else {
				service.Image = service.Image + ":" + newTag
			}
			compose.Services[name] = service
		}
	}

	newData, err := yaml.Marshal(compose)
	if err != nil {
		return err
	}

	return os.WriteFile(composePath, newData, 0644)
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

	cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Dir = servicePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "Error running docker compose: "+err.Error()+"\n"+string(output), http.StatusInternalServerError)
		return
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
