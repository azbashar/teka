package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/A-Bashar/Teka-Finance/internal/fileselector"
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
			cmdArgs = append(cmdArgs, "--value="+value)
	} else if value != "" {
		http.Error(w, "Invalid value mode. Allowed options are then/now/end.",http.StatusBadRequest)
		return
	}

	if period == "M" || period == "Q" || period == "Y" {
		cmdArgs = append(cmdArgs, "-"+period)
	} else if period != "" {
		http.Error(w, "Invalid period. Allowed options are M/Q/Y.",http.StatusBadRequest)
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
	for _,f := range files {
		cmdArgs = append(cmdArgs,"-f", f)
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
		var isData map[string]interface{}
		if err := json.Unmarshal(is, &isData); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse hledger output: %v", err), http.StatusInternalServerError)
			return
		}

		cbrSubreports, ok := isData["cbrSubreports"].([]interface{})
		if !ok || len(cbrSubreports) == 0 {
			http.Error(w, "invalid cbrSubreports in hledger output", http.StatusInternalServerError)
			return
		}

		var items []IncomeReportItem
		totalAmount := 0.0
		totalCurrency := "USD"

		for _, sub := range cbrSubreports {
			subArr, ok := sub.([]interface{})
			if !ok || len(subArr) < 2 {
				continue
			}

			data, ok := subArr[1].(map[string]interface{})
			if !ok {
				continue
			}
			prRows, ok := data["prRows"].([]interface{})
			if !ok {
				continue
			}

			for _, row := range prRows {
				rowMap, ok := row.(map[string]interface{})
				if !ok {
					continue
				}

				accountName, _ := rowMap["prrName"].(string)
				prrTotal, ok := rowMap["prrTotal"].([]interface{})
				if !ok || len(prrTotal) == 0 {
					continue
				}
				amountData, ok := prrTotal[0].(map[string]interface{})
				if !ok {
					continue
				}

				amount := 0.0
				currency := "USD"
				if aq, ok := amountData["aquantity"].(map[string]interface{}); ok {
					amount, _ = aq["floatingPoint"].(float64)
				}
				if comm, ok := amountData["acommodity"].(string); ok {
					currency = comm
				}

				if totalCurrency == "USD" {
					totalCurrency = currency
				}
				totalAmount += amount
				items = append(items, IncomeReportItem{
					Account:  accountName,
					Amount:   amount,
					Currency: currency,
				})
			}
		}

		response := IncomeReportResponse{
			IncomeData: items,
		}
		response.Total.Amount = totalAmount
		response.Total.Currency = totalCurrency

		json.NewEncoder(w).Encode(response)
		return
	}
	w.Write(is)
}