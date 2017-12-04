// Copyright © 2017 Anduin Transactions Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/anduintransaction/fakes3/api"
	"github.com/anduintransaction/fakes3/config"
	"github.com/anduintransaction/fakes3/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the fake s3 server",
	Long:  "Run the fake s3 server",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.ReadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read config file, the error is: %s\n", err)
			os.Exit(1)
		}
		setupLogger(config)
		runServer(config)
	},
}

func setupLogger(config *config.Config) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	switch config.Logging.Level {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	}
	switch config.Logging.Output {
	case "stdout":
		logrus.SetOutput(os.Stdout)
	case "stderr":
		logrus.SetOutput(os.Stderr)
	default:
		output, err := os.Create(config.Logging.Output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot create log file %s\n, error: %s", config.Logging.Output, err)
			os.Exit(1)
		}
		logrus.SetOutput(output)
	}
}

func runServer(config *config.Config) {
	apiServer := server.NewHTTPServer(config.S3ApiServer.HTTP)
	apiServer.Start(api.NewServer(config).Mux)
	err := apiServer.Wait()
	if err != nil {
		logrus.Error(err)
	}
}

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("s3ApiAddr", "b", ":8000", "Listening address for s3 api server")
	serverCmd.Flags().StringP("s3DataFolder", "d", "/data/fakes3", "Data folder for s3")
	serverCmd.Flags().StringP("s3AdvertisedAddr", "a", "", "Advertised address, for prepending to some response. If empty then this value will be calculated from Host Header")
	viper.BindPFlag("s3ApiServer.http.addr", serverCmd.Flags().Lookup("s3ApiAddr"))
	viper.BindPFlag("s3ApiServer.dataFolder", serverCmd.Flags().Lookup("s3DataFolder"))
	viper.BindPFlag("s3ApiServer.advertisedAddr", serverCmd.Flags().Lookup("s3AdvertisedAddr"))
}
