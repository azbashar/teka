package fileselector

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/A-Bashar/Teka-Finance/internal/config"
)

func GetRootDir() string {
	return config.Cfg.EfficientFileStructure.FilesRoot
}

func GetConfigFile() string {
	return filepath.Join(GetRootDir(), "config.journal")
}

func GetMainFile(file, mainFile string) (string, error) {
	if mainFile != "" {
			return mainFile, nil
		}
	if file != "" {
			return file, nil
		}
	if !config.Cfg.EfficientFileStructure.Enabled {
		file = os.Getenv("LEDGER_FILE")
			if file == "" {
				fmt.Println("No ledger file specified. Use --file flag or set LEDGER_FILE environment variable.")
				return "", errors.New("no ledger file specified")
			}
		return file, nil
	}
	return filepath.Join(GetRootDir(), "main.journal"), nil
}

func GetRequiredFiles(start, end, file string) ([]string, string, error) {
	if file != "" {
		return []string{file}, "", nil
	}
	if !config.Cfg.EfficientFileStructure.Enabled {
		file = os.Getenv("LEDGER_FILE")
			if file == "" {
				fmt.Println("No ledger file specified. Use --file flag or set LEDGER_FILE environment variable.")
				return []string{}, "", errors.New("no ledger file specified")
			}
		return []string{file}, "", nil
	}

	if end == "" || start == "" {
		files, err := os.ReadDir(GetRootDir())
		if err != nil {
			return []string{}, "", err
		}
		
		var minYear, maxYear int
		first := true

		for _, f := range files {
			if f.IsDir() {
				year, err := strconv.Atoi(f.Name())
				if err != nil {
					// skip folders that aren’t numeric years
					continue
				}
				if first {
					minYear, maxYear = year, year
					first = false
				} else {
					if year < minYear {
						minYear = year
					}
					if year > maxYear {
						maxYear = year
					}
				}
			}
		}

		if first {
			return []string{}, "", errors.New("no valid year folders found")
		} else {
			if start == "" {
				start = strconv.Itoa(minYear)+"-01-01"
			}
			if end == "" {
				end = strconv.Itoa(maxYear)+"-12-31"
			}
		}
	}

	startDate, err := time.Parse("2006-01-02", start)
	if err != nil {
		return []string{}, "", fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", end)
	if err != nil {
		return []string{}, "", fmt.Errorf("invalid end date: %w", err)
	}

	if endDate.Before(startDate) {
		return []string{}, "", errors.New("end date is before start date")
	}

	startYear := startDate.Year()
	endYear := endDate.Year()

	var parts []string
	for year := startYear; year <= endYear; year++ {
		path := filepath.Join(GetRootDir(), fmt.Sprintf("%d/%d.journal", year, year))
		parts = append(parts, path)
	}
	return parts, fmt.Sprintf("expr:tag:clopen=%v or not tag:clopen", startYear), nil
}

// GetCurrentFile returns the appropriate file path for a given date.
func GetCurrentFile(date, file string) (string, error) {
	if file != "" {
			return file, nil
		}
	if !config.Cfg.EfficientFileStructure.Enabled {
		file = os.Getenv("LEDGER_FILE")
			if file == "" {
				fmt.Println("No ledger file specified. Use --file flag or set LEDGER_FILE environment variable.")
				return "", errors.New("no ledger file specified")
			}
		return file, nil
	}
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
	var quarter int = int(math.Ceil(float64(month) / 3.0))
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
