package cmd

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/A-Bashar/Teka-Finance/frontend"
	"github.com/A-Bashar/Teka-Finance/internal/api"
	"github.com/spf13/cobra"
)


var port string

var serveCmd = &cobra.Command{
	Use: "serve",
	Short: "Start teka web app in headless mode",
	Run: func (cmd *cobra.Command, args []string)  {
		files, _ := fs.Sub(&frontend.Files,"out")
		fs := http.FileServer(http.FS(files))
		http.Handle("/",fs)
		fileArg = rootCmd.Flag("file").Value.String()
		mainFileArg = rootCmd.Flag("mainfile").Value.String()
		api.InitAPI(fileArg, mainFileArg)
		fmt.Printf("Server started on http://localhost:%s\n", port)
		http.ListenAndServe(fmt.Sprintf(":%s",port), nil)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&port,"port","p","8080","Port of the server to listen to.")
}