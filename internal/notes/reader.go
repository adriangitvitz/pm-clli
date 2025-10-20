package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NoteFrontmatter represents the YAML frontmatter in a dn-tui note
type NoteFrontmatter struct {
	ID      string   `yaml:"id"`
	Created string   `yaml:"created"`
	Links   []string `yaml:"links"`
}

// NoteMetadata contains parsed metadata from a note file
type NoteMetadata struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Path      string
}

// ParseNoteFrontmatter reads a note file and extracts metadata
func ParseNoteFrontmatter(path string) (*NoteMetadata, error) {
	// Read the note file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read note file: %w", err)
	}

	// Parse frontmatter
	frontmatter, err := extractFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	// Parse created timestamp
	var createdAt time.Time
	if frontmatter.Created != "" {
		createdAt, err = time.Parse(time.RFC3339, frontmatter.Created)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created timestamp: %w", err)
		}
	}

	// Get file modification time
	updatedAt, err := GetNoteModTime(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file modification time: %w", err)
	}

	return &NoteMetadata{
		ID:        frontmatter.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Path:      path,
	}, nil
}

// extractFrontmatter extracts and parses YAML frontmatter from markdown content
func extractFrontmatter(content string) (*NoteFrontmatter, error) {
	// Check if content starts with ---
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("note does not contain frontmatter")
	}

	// Find the closing ---
	lines := strings.Split(content, "\n")
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil, fmt.Errorf("frontmatter closing delimiter not found")
	}

	// Extract frontmatter YAML
	frontmatterText := strings.Join(lines[1:endIdx], "\n")

	// Parse YAML
	var fm NoteFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterText), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Ensure links is initialized
	if fm.Links == nil {
		fm.Links = []string{}
	}

	return &fm, nil
}

// GetNoteModTime returns the modification time of a note file
func GetNoteModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to stat note file: %w", err)
	}
	return info.ModTime(), nil
}

// FindNoteByID searches for a note file in the dn-tui notes directory by ID
func FindNoteByID(noteID string) (string, error) {
	// Get the notes directory from environment or use default
	notesDir := os.Getenv("DEBUG_NOTES_DIR")
	if notesDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		notesDir = filepath.Join(home, ".debug-notes")
	}

	// Check if notes directory exists
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		return "", fmt.Errorf("notes directory does not exist: %s", notesDir)
	}

	// Read all notes in the directory
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return "", fmt.Errorf("failed to read notes directory: %w", err)
	}

	// Search for the note with matching ID
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(notesDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Try to extract frontmatter
		fm, err := extractFrontmatter(string(content))
		if err != nil {
			continue
		}

		// Check if ID matches
		if fm.ID == noteID {
			return path, nil
		}
	}

	return "", fmt.Errorf("note with ID %s not found", noteID)
}
