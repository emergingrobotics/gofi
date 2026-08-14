package clients

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/unifi-go/gofi/utilities/internal/config"
)

const (
	ouiDatabaseURL      = "https://standards-oui.ieee.org/oui/oui.txt"
	ouiMaxAgeDays       = 30
	ouiFileName         = "oui.txt"
	ouiHexMarker        = "(hex)"
	ouiMACOctetCount    = 3
	ouiMinHexLineParts  = 2
	ouiDownloadTimeout  = 60 * time.Second
	ouiHTTPUserAgent    = "gofimac/1.0"
	ouiMaxDownloadBytes = 20 * 1024 * 1024
)

// OUIDatabase maps 3-octet MAC prefixes to manufacturer names because IEEE assigns
// the first 24 bits of a MAC address as the Organizationally Unique Identifier.
type OUIDatabase struct {
	entries map[string]string
}

// Lookup extracts the 3-octet OUI prefix from a full MAC address and returns
// the IEEE-registered manufacturer, or "unknown" if the prefix is not in the database.
func (database *OUIDatabase) Lookup(macAddress string) string {
	normalized := strings.ToLower(strings.ReplaceAll(macAddress, "-", ":"))
	parts := strings.SplitN(normalized, ":", ouiMACOctetCount+1)
	if len(parts) < ouiMACOctetCount {
		return "unknown"
	}
	prefix := strings.Join(parts[:ouiMACOctetCount], ":")
	if manufacturer, ok := database.entries[prefix]; ok {
		return manufacturer
	}
	return "unknown"
}

// LoadOUIDatabase downloads the IEEE OUI file if missing or stale (>30 days) before parsing.
func LoadOUIDatabase() (*OUIDatabase, error) {
	databasePath, err := ouiDatabasePath()
	if err != nil {
		return nil, err
	}

	if err := ensureOUIFreshness(databasePath); err != nil {
		return nil, err
	}

	file, err := os.Open(databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open OUI database: %w", err)
	}
	defer file.Close()

	return ParseOUIDatabase(file)
}

// ParseOUIDatabase extracts manufacturer names from lines containing "(hex)" markers
// in the IEEE OUI text file, normalizing MAC prefixes from dash-separated to colon-separated.
func ParseOUIDatabase(reader io.Reader) (*OUIDatabase, error) {
	scanner := bufio.NewScanner(reader)
	entries := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, ouiHexMarker) {
			continue
		}

		// Format: "AA-BB-CC   (hex)\t\tManufacturer Name"
		hexIndex := strings.Index(line, ouiHexMarker)
		macPart := strings.TrimSpace(line[:hexIndex])
		manufacturerPart := strings.TrimSpace(line[hexIndex+len(ouiHexMarker):])

		if macPart == "" || manufacturerPart == "" {
			continue
		}

		// Normalize MAC prefix from "AA-BB-CC" to "aa:bb:cc"
		normalized := strings.ToLower(strings.ReplaceAll(macPart, "-", ":"))
		entries[normalized] = manufacturerPart
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse OUI database: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("OUI database contains no entries")
	}

	return &OUIDatabase{entries: entries}, nil
}

func ouiDatabasePath() (string, error) {
	dataDir, err := config.DataDir("gofi")
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, ouiFileName), nil
}

func ensureOUIFreshness(databasePath string) error {
	fileInfo, err := os.Stat(databasePath)
	fileExists := err == nil

	if fileExists {
		age := time.Since(fileInfo.ModTime())
		if age < ouiMaxAgeDays*24*time.Hour {
			return nil
		}
		fmt.Fprintf(os.Stderr, "OUI database is %d days old, updating...\n", int(age.Hours()/24))
	} else {
		fmt.Fprintf(os.Stderr, "OUI database not found, downloading...\n")
	}

	downloadErr := downloadOUIDatabase(databasePath)
	if downloadErr != nil {
		if fileExists {
			fmt.Fprintf(os.Stderr, "Warning: OUI download failed (%v), using cached data\n", downloadErr)
			return nil
		}
		return fmt.Errorf("OUI download failed and no cached data available: %w", downloadErr)
	}

	fmt.Fprintf(os.Stderr, "OUI database updated successfully\n")
	return nil
}

func downloadOUIDatabase(databasePath string) error {
	directory := filepath.Dir(databasePath)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("failed to create OUI data directory: %w", err)
	}

	httpClient := &http.Client{
		Timeout: ouiDownloadTimeout,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	request, err := http.NewRequest(http.MethodGet, ouiDatabaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create OUI download request: %w", err)
	}
	request.Header.Set("User-Agent", ouiHTTPUserAgent)

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to download OUI database: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("OUI download returned status %d", response.StatusCode)
	}

	file, err := os.CreateTemp(directory, "oui-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary OUI file: %w", err)
	}
	temporaryPath := file.Name()

	limitedReader := io.LimitReader(response.Body, ouiMaxDownloadBytes)
	bytesWritten, err := io.Copy(file, limitedReader)
	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(temporaryPath)
		return fmt.Errorf("failed to write OUI database: %w", err)
	}

	if bytesWritten == 0 {
		os.Remove(temporaryPath)
		return fmt.Errorf("downloaded OUI database is empty")
	}

	if err := os.Rename(temporaryPath, databasePath); err != nil {
		os.Remove(temporaryPath)
		return fmt.Errorf("failed to install OUI database: %w", err)
	}

	return nil
}
