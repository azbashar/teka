package prompt

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// SearchAccounts runs hledger accounts and lets user pick by index or type account
func SearchAccounts(searchTerm, file string) (string, error) {
    var cmdArgs []string
    if file != "" {
        cmdArgs = []string{"accounts", "-f", file, searchTerm}
    } else {
        cmdArgs = []string{"accounts", searchTerm}
    }

    out, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
    if err != nil {
        fmt.Println("Error running hledger accounts:", err)
        return "", err
    }

    results := strings.Split(strings.TrimSpace(string(out)), "\n")
    if len(results) == 0 || (len(results) == 1 && results[0] == "") {
        fmt.Println("No accounts found.")
        return "", nil
    }

    // Print indexed list
    fmt.Println("Accounts found:")
    for i, r := range results {
        fmt.Printf("  %d) %s\n", i+1, r)
    }

    // Ask user to choose
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Select account (type index or full name): ")
    choice, _ := reader.ReadString('\n')
    choice = strings.TrimSpace(choice)

    // Try parsing as index
    num, err := strconv.Atoi(choice)
    if err == nil {
        if num >= 1 && num <= len(results) {
            return results[num-1], nil
        }
    }

    // Otherwise use input as account name
    return choice, nil
}
