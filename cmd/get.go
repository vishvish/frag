// Copyright © 2019 Vish Vishvanath <vishvish@users.noreply.github.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"encoding/json"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		s, retrieveError := retrieveItem()
		if retrieveError != nil {
			log.Println("Please login first.")
		} else {
			log.Println("Token found in keychain.")
			var t oauth2.Token
			err := json.Unmarshal(s.Data, &t)
			if err != nil {
				log.Println("Please login first.")
			}
			client := conf.Client(ctx, &t)

			resp, err := client.Get("https://api.freeagent.com/v2/projects")
			if err != nil {
				log.Fatal(err)
			} else {
				log.Println(color.CyanString("API call successful"))
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				bodyBytes, err2 := ioutil.ReadAll(resp.Body)
				if err2 != nil {
					log.Fatal(err2)
				} else {
					bodyString := string(bodyBytes)
					log.Println(color.GreenString(bodyString))
				}
			}

		}

		// The HTTP Client returned by conf.Client will refresh the token as necessary.
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
