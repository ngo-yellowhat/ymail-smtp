package cmd

import (
	"log"
	"time"

	"yellowsmtp/internal/inbound"

	"github.com/emersion/go-smtp"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		be := &inbound.Backend{}
		s := smtp.NewServer(be)

		s.Addr = ":2525"
		s.Domain = "localhost"
		s.WriteTimeout = 5 * time.Minute
		s.ReadTimeout = 10 * time.Second
		s.AllowInsecureAuth = true

		if err := s.ListenAndServe(); err != nil {
			log.Printf("[  SERVER  ] SMTP start on %s:%s", s.Domain, s.Addr)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
