// @title Customer Gateway Service API
// @version 1.0
// @description HTTP API Gateway for Stockanalyzr - manages customer operations with JWT authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@stockanalyzr.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "customer-gw-service",
	Short: "customer-gw-service manages customer gateway operations for stockanalyzr",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
