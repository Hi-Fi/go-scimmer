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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hi-fi/go-scimmer/pkg/ldap"
	"github.com/hi-fi/go-scimmer/pkg/model"
	"github.com/hi-fi/go-scimmer/pkg/scim"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string
var ldapConfig ldap.Config
var scimConfig scim.Config
var modelConfig model.IDMap
var cmdConfig Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-scimmer",
	Short: "Sync identities from one system to another",
	Long: `Sync identities from one system to another. This utility is needed
when wanting to sync from system that doesn't offer SCIM at all or allows
if only to preauthorized systems.

Currently supports LDAP as a source. SCIM target needs to support token
authentication. `,
	Run: func(cmd *cobra.Command, args []string) {
		if cmdConfig.JSONLogging {
			log.SetFormatter(&log.JSONFormatter{})
		}
		logLevel, err := log.ParseLevel(cmdConfig.LogLevel)
		if err != nil {
			log.Infof("Incorrect log level (%s) given. Setting loglevel to info", cmdConfig.LogLevel)
			logLevel = log.InfoLevel
		}
		log.SetLevel(logLevel)
		users, groups := ldapConfig.LoadUsersAndGroups()
		modelConfig.EnrichUsersAndGroupsWithScimIDs(users, groups)
		scimConfig.SyncIdentities(users, groups, &modelConfig)
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

	// Variables also from environment

	viper.SetEnvPrefix("scimmer")
	viper.AutomaticEnv()

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-scimmer.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVar(&cmdConfig.JSONLogging, "json_logging", false, "Log in JSON format")
	rootCmd.Flags().StringVar(&cmdConfig.LogLevel, "log_level", log.InfoLevel.String(), "Logging level. Possible values panic, fatal, error, warn, info, debug, trace.")
	rootCmd.Flags().StringVar(&ldapConfig.Host, "ldap_host", "localhost", "LDAP server host")
	rootCmd.Flags().IntVar(&ldapConfig.Port, "ldap_port", 389, "LDAP server host")
	rootCmd.Flags().StringVar(&ldapConfig.UserDN, "ldap_user", "", "User (DN) to bind to LDAP server. Needs only read righs.")
	rootCmd.Flags().StringVar(&ldapConfig.Password, "ldap_password", "", "Password for LDAP user")
	rootCmd.Flags().StringVar(&ldapConfig.ActiveUserAttribute, "ldap_activeuser", "", "Attribute that tells if user is active or not. If empty, all users considered as active")
	rootCmd.Flags().StringVar(&ldapConfig.ActiveUserValue, "ldap_activeuservalue", "", "Value that means that user is active. Only valid if active user attribute is defined")
	rootCmd.Flags().StringVar(&ldapConfig.GroupFilter, "ldap_groupfilter", "(|(objectclass=posixGroup)(objectclass=group)(objectclass=groupOfNames)(objectclass=groupOfUniqueNames))", "Filter for LDAP groups")
	rootCmd.Flags().StringVar(&ldapConfig.UserFilter, "ldap_userfilter", "(|(objectclass=user)(objectclass=person)(objectclass=inetOrgPerson)(objectclass=organizationalPerson))", "Filter for LDAP users")
	rootCmd.Flags().StringVar(&ldapConfig.BaseDN, "ldap_basedn", "dc=example,dc=org", "BaseDN to use in search")
	rootCmd.Flags().StringVar(&ldapConfig.UsernameAttribute, "ldap_username_attribute", "uid", "Attribute to be used as username in the external system")
	rootCmd.Flags().StringVar(&modelConfig.FilePath, "output", "scim_id_map.yaml", "File to write internal and external id mapping to. Note that this file needs to be kept safe to allow updates to objects.")
	rootCmd.Flags().StringVar(&scimConfig.EndpointURL, "scim_endpoint", "", "SCIM server endpoint")
	rootCmd.Flags().StringVar(&scimConfig.Token, "scim_token", "", "Authentication (Bearer) token to SCIM endpoint")
	rootCmd.Flags().BoolVar(&scimConfig.DryRun, "scim_dryrun", true, "Execute dry run that just prints out the messages that would have been sent to server")
	rootCmd.Flags().BoolVar(&scimConfig.BulkSupported, "scim_bulk_supported", false, "SCIM endpoint bulk support. If not supported, objects are synced one by one on parallel.")

	// Updating flags from viper (environment variables)
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			rootCmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
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

		// Search config in home directory with name ".go-scimmer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".go-scimmer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
