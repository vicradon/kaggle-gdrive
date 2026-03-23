package main

import (
	"context"
	"flag"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var (
	fromPath = flag.String("from", "", "Local source path (file or directory)")
	toPath   = flag.String("to", "", `Google Drive destination (e.g. "/drive/My Drive/PetCrib")`)
)

func main() {
	flag.Parse()

	if *fromPath == "" || *toPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: gdrive -from <local-path> -to <drive-path>")
		os.Exit(1)
	}

	saKey := os.Getenv("GDRIVE_SA_KEY")
	if saKey == "" {
		fmt.Fprintln(os.Stderr, "error: GDRIVE_SA_KEY environment variable is not set")
		os.Exit(1)
	}

	ctx := context.Background()

	conf, err := google.JWTConfigFromJSON([]byte(saKey), drive.DriveScope)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to parse service account key: %v\n", err)
		os.Exit(1)
	}

	svc, err := drive.NewService(ctx, option.WithHTTPClient(conf.Client(ctx)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to create Drive service: %v\n", err)
		os.Exit(1)
	}

	drivePath := normaliseDrivePath(*toPath)

	destFolderID, err := ensureFolderPath(ctx, svc, drivePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to resolve destination folder: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(*fromPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot stat source path: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		if err := uploadDir(ctx, svc, *fromPath, destFolderID); err != nil {
			fmt.Fprintf(os.Stderr, "error: upload failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if _, err := uploadFile(ctx, svc, *fromPath, destFolderID); err != nil {
			fmt.Fprintf(os.Stderr, "error: upload failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("✓ Upload complete.")
}

func normaliseDrivePath(p string) string {
	prefixes := []string{
		"/drive/My Drive/",
		"/drive/My Drive",
		"My Drive/",
		"My Drive",
	}
	for _, pfx := range prefixes {
		if strings.HasPrefix(p, pfx) {
			return strings.TrimPrefix(p, pfx)
		}
	}
	return strings.TrimPrefix(p, "/")
}

func ensureFolderPath(ctx context.Context, svc *drive.Service, path string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	parentID := "root"

	for _, part := range parts {
		if part == "" {
			continue
		}
		id, err := findOrCreateFolder(ctx, svc, part, parentID)
		if err != nil {
			return "", fmt.Errorf("folder %q: %w", part, err)
		}
		parentID = id
	}
	return parentID, nil
}

func findOrCreateFolder(ctx context.Context, svc *drive.Service, name, parentID string) (string, error) {
	q := fmt.Sprintf(
		"name = %q and mimeType = 'application/vnd.google-apps.folder' and %q in parents and trashed = false",
		name, parentID,
	)
	res, err := svc.Files.List().
		Q(q).
		Fields("files(id, name)").
		Context(ctx).
		Do()
	if err != nil {
		return "", err
	}
	if len(res.Files) > 0 {
		return res.Files[0].Id, nil
	}

	f := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}
	created, err := svc.Files.Create(f).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", err
	}
	fmt.Printf("  📁 Created folder: %s\n", name)
	return created.Id, nil
}

func uploadDir(ctx context.Context, svc *drive.Service, localDir, parentID string) error {
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		localPath := filepath.Join(localDir, entry.Name())
		if entry.IsDir() {
			subID, err := findOrCreateFolder(ctx, svc, entry.Name(), parentID)
			if err != nil {
				return err
			}
			if err := uploadDir(ctx, svc, localPath, subID); err != nil {
				return err
			}
		} else {
			if _, err := uploadFile(ctx, svc, localPath, parentID); err != nil {
				return err
			}
		}
	}
	return nil
}

func uploadFile(ctx context.Context, svc *drive.Service, localPath, parentID string) (*drive.File, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mimeType := detectMIME(localPath)
	name := filepath.Base(localPath)

	meta := &drive.File{
		Name:    name,
		Parents: []string{parentID},
	}

	fmt.Printf("  ⬆  Uploading: %s  (%s)\n", name, mimeType)

	uploaded, err := svc.Files.Create(meta).
		Media(f, googleapi.ContentType(mimeType)).
		Fields("id, name, size").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("upload %q: %w", name, err)
	}

	fmt.Printf("     ✓ %s  [id: %s, %d bytes]\n", uploaded.Name, uploaded.Id, uploaded.Size)
	return uploaded, nil
}

func detectMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}
