package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/azbashar/teka/internal/config"
	"github.com/azbashar/teka/internal/fileselector"
)

func getIncomeStatement(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	value := r.URL.Query().Get("valueMode")
	outputFormat := r.URL.Query().Get("outputFormat")
	period := r.URL.Query().Get("period")
	account := r.URL.Query().Get("account")
	depthStr := r.URL.Query().Get("depth")

	if depthStr != "" {
		depth, err := strconv.Atoi(depthStr)
		if err != nil || depth < 1 {
			http.Error(w, "Invalid depth value. It should be a positive integer.", http.StatusBadRequest)
			return
		}
	}

	cmdArgs := []string{"is"}

	if account != "" {
		cmdArgs = append(cmdArgs, account)
	}

	switch outputFormat {
	case "":
		cmdArgs = append(cmdArgs, "-O", "csv")
	case "csv", "json", "html", "txt":
		cmdArgs = append(cmdArgs, "-O", outputFormat)
	default:
		http.Error(w, "Invalid output format. Allowed values are csv/json/html/txt.", http.StatusBadRequest)
		return
	}

	if startDate != "" {
		cmdArgs = append(cmdArgs, "-b", startDate)
	}
	if endDate != "" {
		cmdArgs = append(cmdArgs, "-e", endDate)
	}

	if value == "then" || value == "now" || value == "end" {
		cmdArgs = append(cmdArgs, "--value="+value+","+config.Cfg.BaseCurrency)
	} else if value != "" {
		http.Error(w, "Invalid value mode. Allowed options are then/now/end.", http.StatusBadRequest)
		return
	}

	if period == "M" || period == "Q" || period == "Y" {
		cmdArgs = append(cmdArgs, "-"+period)
	} else if period != "" {
		http.Error(w, "Invalid period. Allowed options are M/Q/Y.", http.StatusBadRequest)
		return
	}

	if depthStr != "" {
		cmdArgs = append(cmdArgs, "--depth="+depthStr)
	}

	files, expr, err := fileselector.GetRequiredFiles(startDate, endDate, fileArg)
	if err != nil {
		fmt.Println("File selector error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, f := range files {
		cmdArgs = append(cmdArgs, "-f", f)
	}

	if expr != "" {
		cmdArgs = append(cmdArgs, expr)
	}

	is, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println(string(is))
		fmt.Println("Error running hledger:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if outputFormat == "html" {
		// replace invalid utf8 characters with &nbsp; to prevent breaking html rendering
		var b strings.Builder
		for len(is) > 0 {
			r, s := utf8.DecodeRune(is)
			if r == utf8.RuneError && s == 1 {
				b.WriteString("&nbsp;")
				is = is[1:]
			} else {
				b.WriteRune(r)
				is = is[s:]
			}
		}
		is = []byte(b.String())
		w.Header().Set("Content-Type", "text/html")
	} else if outputFormat == "json" {
		w.Header().Set("Content-Type", "application/json")
		var isData map[string]any
		if err := json.Unmarshal(is, &isData); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse hledger output: %v", err), http.StatusInternalServerError)
			return
		}

		// build period-based response
		cbrDates, ok := isData["cbrDates"].([]any)
		if !ok {
			http.Error(w, "invalid cbrDates in hledger output", http.StatusInternalServerError)
			return
		}

		cbrSubreports, ok := isData["cbrSubreports"].([]any)
		if !ok {
			http.Error(w, "invalid cbrSubreports in hledger output", http.StatusInternalServerError)
			return
		}

		cbrTotals, ok := isData["cbrTotals"].(map[string]any)
		if !ok {
			http.Error(w, "invalid cbrTotals in hledger output", http.StatusInternalServerError)
			return
		}
		prrAmountsTotals, _ := cbrTotals["prrAmounts"].([]any)

		var periodReports []map[string]any

		// loop over each period index
		for i, periodRange := range cbrDates {
			rangeArr, ok := periodRange.([]any)
			if !ok || len(rangeArr) < 2 {
				continue
			}
			fromDate := ""
			toDate := ""
			if f, ok := rangeArr[0].(map[string]any); ok {
				fromDate, _ = f["contents"].(string)
			}
			if t, ok := rangeArr[1].(map[string]any); ok {
				toDate, _ = t["contents"].(string)
			}

			var data []map[string]any
			totalAmount := 0.0
			totalCurrency := config.Cfg.BaseCurrency

			// walk through subreports -> prRows -> prrAmounts[i]
			for _, sub := range cbrSubreports {
				subArr, ok := sub.([]any)
				if !ok || len(subArr) < 2 {
					continue
				}
				dataMap, ok := subArr[1].(map[string]any)
				if !ok {
					continue
				}
				prRows, ok := dataMap["prRows"].([]any)
				if !ok {
					continue
				}

				for _, row := range prRows {
					rowMap, ok := row.(map[string]any)
					if !ok {
						continue
					}
					accountName, _ := rowMap["prrName"].(string)
					prrAmounts, ok := rowMap["prrAmounts"].([]any)
					if !ok || i >= len(prrAmounts) {
						continue
					}
					periodAmounts, ok := prrAmounts[i].([]any)
					if !ok || len(periodAmounts) == 0 {
						continue
					}

					amtData, ok := periodAmounts[0].(map[string]any)
					if !ok {
						continue
					}
					amount := 0.0
					currency := config.Cfg.BaseCurrency
					if aq, ok := amtData["aquantity"].(map[string]any); ok {
						amount, _ = aq["floatingPoint"].(float64)
					}
					if comm, ok := amtData["acommodity"].(string); ok {
						currency = comm
					}

					if totalCurrency == config.Cfg.BaseCurrency {
						totalCurrency = currency
					}
					totalAmount += amount

					data = append(data, map[string]any{
						"account":  accountName,
						"amount":   amount,
						"currency": currency,
					})
				}
			}

			// extract period total if available
			periodTotal := map[string]any{
				"amount":   totalAmount,
				"currency": totalCurrency,
			}
			if i < len(prrAmountsTotals) {
				if arr, ok := prrAmountsTotals[i].([]any); ok && len(arr) > 0 {
					if amtData, ok := arr[0].(map[string]any); ok {
						if aq, ok := amtData["aquantity"].(map[string]any); ok {
							if fp, ok := aq["floatingPoint"].(float64); ok {
								periodTotal["amount"] = fp
							}
						}
						if comm, ok := amtData["acommodity"].(string); ok {
							periodTotal["currency"] = comm
						}
					}
				}
			}

			if len(data) > 0 {
				periodReports = append(periodReports, map[string]any{
					"dates": map[string]string{
						"from": fromDate,
						"to":   toDate,
					},
					"total": periodTotal,
					"data":  data,
				})
			}
		}

		resp := periodReports

		json.NewEncoder(w).Encode(resp)
		return
	}

	w.Write(is)
}
