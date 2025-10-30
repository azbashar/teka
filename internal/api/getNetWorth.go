package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/azbashar/teka/internal/config"
	"github.com/azbashar/teka/internal/fileselector"
)

type NetWorth struct {
	Date     string  `json:"date"`
	Networth float64 `json:"networth"`
	Currency string  `json:"currency"`
}

func getNetWorth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	// Prepare hledger command: use balance sheet
	cmdArgs := []string{
		"bs",      // balance sheet
		"-V",      // value in default currency
		"--daily", // daily snapshot
		"-O", "json",
	}
	if startDate != "" {
		cmdArgs = append(cmdArgs, "--begin", startDate)
	}
	if endDate != "" {
		cmdArgs = append(cmdArgs, "--end", endDate)
	}

	// Add file args from fileselector
	files, expr, err := fileselector.GetRequiredFiles(startDate, endDate, fileArg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, f := range files {
		cmdArgs = append(cmdArgs, "-f", f)
	}
	if expr != "" {
		cmdArgs = append(cmdArgs, expr)
	}

	cmd := exec.Command("hledger", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		http.Error(w, "hledger error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var bsData map[string]interface{}
	if err := json.Unmarshal(output, &bsData); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse hledger output: %v", err), http.StatusInternalServerError)
		return
	}

	cbrDates, ok := bsData["cbrDates"].([]interface{})
	if !ok {
		http.Error(w, "invalid cbrDates in hledger output", http.StatusInternalServerError)
		return
	}

	cbrSubreports, ok := bsData["cbrSubreports"].([]interface{})
	if !ok {
		http.Error(w, "invalid cbrSubreports in hledger output", http.StatusInternalServerError)
		return
	}

	// Helper to get totals safely
	getTotals := func(name string) []map[string]interface{} {
		for _, sub := range cbrSubreports {
			subArr, ok := sub.([]interface{})
			if !ok || len(subArr) < 2 {
				continue
			}
			subName, ok := subArr[0].(string)
			if !ok || !strings.EqualFold(subName, name) {
				continue
			}
			data, ok := subArr[1].(map[string]interface{})
			if !ok {
				continue
			}
			prTotals, ok := data["prTotals"].(map[string]interface{})
			if !ok {
				continue
			}
			prrAmounts, ok := prTotals["prrAmounts"].([]interface{})
			if !ok || len(prrAmounts) == 0 {
				continue
			}
			var result []map[string]interface{}
			for _, period := range prrAmounts {
				periodSlice, ok := period.([]interface{})
				if !ok || len(periodSlice) == 0 {
					result = append(result, nil)
					continue
				}
				amount, ok := periodSlice[0].(map[string]interface{})
				if !ok {
					result = append(result, nil)
					continue
				}
				result = append(result, amount)
			}
			return result
		}
		return nil
	}

	assetsTotals := getTotals("Assets")
	liabilitiesTotals := getTotals("Liabilities")

	var results []NetWorth
	for i, dateRange := range cbrDates {
		drSlice, ok := dateRange.([]interface{})
		if !ok || len(drSlice) == 0 {
			continue
		}
		lastDateObj, ok := drSlice[len(drSlice)-1].(map[string]interface{})
		if !ok {
			continue
		}
		dateStr, ok := lastDateObj["contents"].(string)
		if !ok {
			continue
		}

		assetVal, currency := 0.0, config.Cfg.BaseCurrency

		if i < len(assetsTotals) && assetsTotals[i] != nil {
			aquantity, ok := assetsTotals[i]["aquantity"].(map[string]interface{})
			if ok {
				assetVal, _ = aquantity["floatingPoint"].(float64)
			}
			if comm, ok := assetsTotals[i]["acommodity"].(string); ok {
				currency = comm
			}
		}

		liabVal := 0.0
		if i < len(liabilitiesTotals) && liabilitiesTotals[i] != nil {
			aquantity, ok := liabilitiesTotals[i]["aquantity"].(map[string]interface{})
			if ok {
				liabVal, _ = aquantity["floatingPoint"].(float64)
			}
		}

		results = append(results, NetWorth{
			Date:     dateStr,
			Networth: assetVal - liabVal,
			Currency: currency,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
