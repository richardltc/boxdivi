/*
Package cmd ...
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	be "richardmace.co.uk/boxwallet/cmd/cli/cmd/bend"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops your chosen coin's daemon server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("  ____          __          __   _ _      _   \n |  _ \\         \\ \\        / /  | | |    | |  \n | |_) | _____  _\\ \\  /\\  / /_ _| | | ___| |_ \n |  _ < / _ \\ \\/ /\\ \\/  \\/ / _` | | |/ _ \\ __|\n | |_) | (_) >  <  \\  /\\  / (_| | | |  __/ |_ \n |____/ \\___/_/\\_\\  \\/  \\/ \\__,_|_|_|\\___|\\__| v" + be.CBWAppVersion + "\n                                              \n                                               ")

		apw, err := be.GetAppWorkingFolder()
		if err != nil {
			log.Fatal("Unable to GetAppWorkingFolder: " + err.Error())
		}

		// Make sure the config file exists, and if not, force user to use "coin" command first.
		if _, err := os.Stat(apw + be.CConfFile + be.CConfFileExt); os.IsNotExist(err) {
			log.Fatal("Unable to determine coin type. Please run " + be.CAppFilename + " coin first")
		}

		cliConf, err := be.GetConfigStruct("", true)
		if err != nil {
			log.Fatal("Unable to GetCLIConfStruct " + err.Error())
		}

		sCoinDaemonName, err := be.GetCoinDaemonFilename(be.APPTCLI, cliConf.ProjectType)
		if err != nil {
			log.Fatal("Unable to GetCoinDaemonFilename " + err.Error())
		}

		// If we're running locally and the coin daemon is not currently running, inform the user.
		if cliConf.ServerIP != "127.0.0.1" {
			bIsRunning, _, _ := be.IsCoinDaemonRunning(cliConf.ProjectType)
			if !bIsRunning {
				log.Fatal("The coin server " + sCoinDaemonName + " is not currently running.")
			}
		}

		switch cliConf.ProjectType {
		case be.PTScala:
			resp, err := be.StopDaemonMonero(&cliConf)
			if err != nil {
				log.Fatal("Unable to StopDaemon " + err.Error())
			}
			fmt.Println(resp.Status)
			fmt.Println("daemon stopping")
		default:
			fmt.Println("Stopping the " + sCoinDaemonName + " server...")
			_, err := be.StopDaemon(&cliConf)
			if err != nil {
				log.Fatal("Unable to StopDaemon " + err.Error())
			}
			for i := 0; i < 600; i++ {
				bStillRunning, _, _ := be.IsCoinDaemonRunning(cliConf.ProjectType)
				if bStillRunning {
					_, _ = be.StopDaemon(&cliConf)
					fmt.Printf("\r" + "Waiting for " + sCoinDaemonName + " to stop. This could take a long time on slower devices... " + strconv.Itoa(i+1))
					time.Sleep(1 * time.Second)
				} else {
					fmt.Println("\n" + sCoinDaemonName + " server stopped.")
					break
				}
			}
		}

		// sAppCLIName, err := gwc.GetAppCLIName() // e.g. GoDivi CLI
		// if err != nil {
		// 	log.Fatal("Unable to GetAppCLIName " + err.Error())
		// }

		// sAppFileCLIName, err := gwc.GetAppFileName(gwc.APPTCLI)
		// if err != nil {
		// 	log.Fatal("Unable to GetAppFileCLIName " + err.Error())
		// }
		// sCoinDaemonFile, err := gwc.GetCoinDaemonFilename()
		// if err != nil {
		// 	log.Fatal("Unable to GetCoinDaemonFilename " + err.Error())
		// }

		// // Check to make sure we're installed
		// if !gwc.IsGoWalletInstalled() {
		// 	log.Fatal(sAppCLIName + ` doesn't appear to be installed yet. Please run "` + sAppFileCLIName + ` install" first`)
		// }

		// err = gwc.StopCoinDaemon().
		// if err != nil {
		// 	log.Fatal(`Unable to stop ` + sCoinDaemonFile + ` server:` + err.Error())
		// }
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
