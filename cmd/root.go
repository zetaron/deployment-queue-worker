package cmd

// Copyright ©2016 Fabian Stegemann
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/adjust/rmq"
)

var cfgFile string
var version = "1.0.0"
var workerImageName string
var deploymentContainerNameTemplate *template.Template
var deploymentIDTemplate *template.Template
var secretsVolumeNameTemplate *template.Template
var cacheVolumeNameTemplate *template.Template

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "queue-worker",
	Short: "Process queue entries.",
	Long:  `Starts an instance of the worker image with mounted project environment secrets and a deployment cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("deployment_id_template", "deployment-{{.deployment.id}}")
		viper.SetDefault("cache_volume_name_template", "{{.repository.name}}-{{.deployment.environment}}-deployment-{{.deployment.id}}")
		viper.SetDefault("secrets_volume_name_template", "{{.repository.name}}-{{.deployment.environment}}-secrets")
		viper.SetDefault("redis.url", "redis:6379")
		viper.SetDefault("redis.database", 1)
		viper.SetDefault("worker.count", 1)

		log.Infof("Starting queue-worker version %s", version)

		deploymentIDTemplate = template.Must(template.New("deployment-id").Parse(viper.GetString("deployment_id_template")))
		cacheVolumeNameTemplate = template.Must(template.New("cache-volume-name-template").Parse(viper.GetString("cache_volume_name_template")))
		secretsVolumeNameTemplate = template.Must(template.New("secrets-volume-name").Parse(viper.GetString("secrets_volume_name_template")))

		connection := rmq.OpenConnection(
			"queue-worker",
			"tcp",
			viper.GetString("redis.url"),
			viper.GetInt("redis.database"),
		)
		log.WithFields(log.Fields{
			"redis-url":      viper.GetString("redis.url"),
			"redis-database": viper.GetInt("redis.database"),
		}).Info("Connected to redis")

		queue := connection.OpenQueue("deployment_events")
		queue.StartConsuming(viper.GetInt("worker.count"), time.Second)
		log.Info("Opened queue")

		for i := 0; i < viper.GetInt("worker.count"); i++ {
			name := fmt.Sprintf("consumer-%d-on-database-%d", i, viper.GetInt("redis.database"))
			queue.AddConsumer(name, &DeploymentEventConsumer{
				name: name,
			})

			log.WithFields(log.Fields{
				"name":           name,
				"worker-count":   viper.GetInt("worker.count"),
				"redis-database": viper.GetInt("redis.database"),
				"redis-url":      viper.GetString("redis.url"),
			}).Info("Started worker ", i)
		}

		// while true
		select {}
	},
}

//Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings
	// Cobra supports Persistent Flags which if defined here will be global for your application

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/cutter/queue-worker.yaml)")

	// Cobra also supports local flags which will only run when this action is called directly
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// Read in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	// allow for nested environment variables
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetConfigName("queue-worker") // name of config file (without extension)
	viper.AddConfigPath("/etc/cutter")  // adding home directory as first search path
	viper.AutomaticEnv()                // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
