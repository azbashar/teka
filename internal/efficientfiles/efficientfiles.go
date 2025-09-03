package efficientfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/A-Bashar/Teka-Finance/internal/config" // adjust import path to where your config package lives
)

func GetRootDir() string {
	return config.Cfg.EfficientFileStructure.FilesRoot
}

func GetMainFile() string {
	return filepath.Join(GetRootDir(), "main.journal")
}

func GetConfigFile() string {
	return filepath.Join(GetRootDir(), "config.journal")
}

// GetRequiredFiles returns a string with all year journal files covering the date range.
func GetRequiredFiles(start, end string) (string, error) {
	startDate, err := time.Parse("2006-01-02", start)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", end)
	if err != nil {
		return "", fmt.Errorf("invalid end date: %w", err)
	}

	if endDate.Before(startDate) {
		return "", errors.New("end date is before start date")
	}

	startYear := startDate.Year()
	endYear := endDate.Year()

	var parts []string
	for year := startYear; year <= endYear; year++ {
		path := filepath.Join(GetRootDir(), fmt.Sprintf("%d/%d.journal", year, year))
		parts = append(parts, "-f "+path)
	}
	return strings.Join(parts, " "), nil
}

// GetCurrentFile returns the appropriate file path for a given date.
func GetCurrentFile(date string) (string, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("invalid date: %w", err)
	}
	year := d.Year()
	month := int(d.Month())

	yearDir := filepath.Join(GetRootDir(), fmt.Sprintf("%d", year))

	// Month file: yearM<month>
	monthFile := filepath.Join(yearDir, fmt.Sprintf("%dM%d.journal", year, month))
	if fileExists(monthFile) {
		return monthFile, nil
	}

	// Quarter file: yearQ<quarter>
	quarter := (month-1)/3 + 1
	quarterFile := filepath.Join(yearDir, fmt.Sprintf("%dQ%d.journal", year, quarter))
	if fileExists(quarterFile) {
		return quarterFile, nil
	}

	// Half file: yearH1 or yearH2
	half := 1
	if month > 6 {
		half = 2
	}
	halfFile := filepath.Join(yearDir, fmt.Sprintf("%dH%d.journal", year, half))
	if fileExists(halfFile) {
		return halfFile, nil
	}

	// Year file: year.journal
	yearFile := filepath.Join(yearDir, fmt.Sprintf("%d.journal", year))
	if fileExists(yearFile) {
		return yearFile, nil
	}

	return "", fmt.Errorf("file not found for year %d. initialize year first", year)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
