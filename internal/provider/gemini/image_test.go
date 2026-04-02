package gemini

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

func TestSaveImage(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	data := []byte("fake-png-data")
	path, err := p.saveImage(data, "png")
	if err != nil {
		t.Fatalf("saveImage: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("saved to %q, want dir %q", path, dir)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if string(content) != string(data) {
		t.Errorf("file content = %q, want %q", content, data)
	}
}

func TestSaveImage_FilenameFormat(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	path, err := p.saveImage([]byte("data"), "jpg")
	if err != nil {
		t.Fatalf("saveImage: %v", err)
	}

	base := filepath.Base(path)
	if ext := filepath.Ext(base); ext != ".jpg" {
		t.Errorf("extension = %q, want .jpg", ext)
	}
	if base[:6] != "image-" {
		t.Errorf("filename %q does not start with 'image-'", base)
	}
}

func TestGenerate_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Generate(context.Background(), provider.ImageRequest{Prompt: ""})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestEdit_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Edit(context.Background(), provider.EditImageRequest{Prompt: "", ImagePath: "/tmp/img.png"})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestEdit_EmptyImagePath(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Edit(context.Background(), provider.EditImageRequest{Prompt: "edit", ImagePath: ""})
	if err == nil {
		t.Fatal("expected error for empty imagePath")
	}
}

func TestCompose_TooManyReferences(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Compose(context.Background(), provider.ComposeRequest{
		Prompt:          "compose",
		ReferenceImages: []string{"a.png", "b.png", "c.png", "d.png"},
	})
	if err == nil {
		t.Fatal("expected error for >3 reference images")
	}
}

func TestCompose_NoReferences(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Compose(context.Background(), provider.ComposeRequest{
		Prompt:          "compose",
		ReferenceImages: []string{},
	})
	if err == nil {
		t.Fatal("expected error for empty reference images")
	}
}

func TestCompose_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Compose(context.Background(), provider.ComposeRequest{
		Prompt:          "",
		ReferenceImages: []string{"a.png"},
	})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestExtensionFromMIME(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"image/png", "png"},
		{"image/jpeg", "jpg"},
		{"image/webp", "webp"},
		{"image/unknown", "png"},
		{"", "png"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got := extensionFromMIME(tt.mime)
			if got != tt.want {
				t.Errorf("extensionFromMIME(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestMimeFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"photo.png", "image/png"},
		{"photo.webp", "image/webp"},
		{"photo.gif", "image/gif"},
		{"photo.bmp", "image/png"},
		{"no-extension", "image/png"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := mimeFromPath(tt.path)
			if got != tt.want {
				t.Errorf("mimeFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
