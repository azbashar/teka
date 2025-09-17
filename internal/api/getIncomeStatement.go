package api

import (
	"fmt"
	"net/http"
	"os/exec"
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

	cmdArgs := []string{"is"}

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
	}
	w.Write(is)
}