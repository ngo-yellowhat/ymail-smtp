package cmd

import (
	"fmt"
	"log"
	"time"

	"yellowsmtp/internal/inbound"

	"github.com/emersion/go-smtp"
	"github.com/spf13/cobra"
)

var domain string
var port int
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		be := &inbound.Backend{
		}
		s := smtp.NewServer(be)

		s.Addr = fmt.Sprintf(":%d", port)
		s.Domain = fmt.Sprintf("%s", domain)
		s.WriteTimeout = 5 * time.Minute
		s.ReadTimeout = 10 * time.Second
		s.AllowInsecureAuth = true

		log.Printf("[  SERVER  ] SMTP start on %s%s", s.Domain, s.Addr)
		if err := s.ListenAndServe(); err != nil {
			log.Printf("[  ERROR  ] SMTP server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&port, "port", "p", 25, "Set port for SMTP server")
	startCmd.Flags().StringVarP(&domain, "domain", "d", "localhost", "Set domain for SMTP server")
}
