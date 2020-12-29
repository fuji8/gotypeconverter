/*
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
	"io/ioutil"
	"os"
	"strings"

	"github.com/fuji8/goconvertstruct/internal"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var (
	file     string
	src, dst string
	outFile  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goconvertstruct",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		g := new(internal.Generator)
		g.Init(file)

		data, err := g.Generate(src, dst)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		if outFile == "" {
			outFile = strings.Split(file, ".")[0]
			outFile += "_generated.go"
		}
		err = ioutil.WriteFile(outFile, data, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goconvertstruct.yaml)")

	rootCmd.PersistentFlags().StringVarP(&src, "src", "s", "", "source struct")
	rootCmd.MarkPersistentFlagRequired("src")

	rootCmd.PersistentFlags().StringVarP(&dst, "dst", "d", "", "destination struct")
	rootCmd.MarkPersistentFlagRequired("dst")

	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "input file")
	rootCmd.MarkPersistentFlagRequired("file")

	rootCmd.PersistentFlags().StringVarP(&outFile, "out-file", "o", "", "output file (default is [file]_generated.go)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".goconvertstruct" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".goconvertstruct")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
