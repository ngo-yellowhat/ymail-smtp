package cmd

import (
	"fmt"
	"log"
	"time"

	"ysmtp/internal/inbound"
	"ysmtp/internal/outbound"

	"github.com/emersion/go-smtp"
	"github.com/spf13/cobra"
)

var (
	heloDomain   string
	port         int
	writeTimeout time.Duration
	readTimeout  time.Duration
	dkimDomain   string
	dkimSelector string
	dkimKeyPath  string
	startCmd     = &cobra.Command{
		Use:   "start",
		Short: "Start SMTP server",
		Run: func(cmd *cobra.Command, args []string) {
			sender := &outbound.Sender{
				HeloDomain:   heloDomain,
				DKIMdomain:   dkimDomain,
				DKIMselector: dkimSelector,
				DKIMkey:      dkimKeyPath,
			}
			be := inbound.NewBackend(sender, dkimDomain)
			s := smtp.NewServer(be)

			s.Addr = fmt.Sprintf(":%d", port)
			s.Domain = fmt.Sprintf("%s", heloDomain)
			s.WriteTimeout = writeTimeout
			s.ReadTimeout = readTimeout
			s.AllowInsecureAuth = true

			log.Printf("[  SERVER  ] SMTP start on %s%s", s.Domain, s.Addr)
			if err := s.ListenAndServe(); err != nil {
				log.Printf("[  ERROR  ] SMTP server: %v", err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&port, "port", "p", 25, "Set port for SMTP server")
	startCmd.Flags().StringVar(&heloDomain, "helo-domain", "mail.yellowhat.cz", "Set helo-domain for SMTP server")
	startCmd.Flags().StringVar(&dkimDomain, "dkim-domain", "yellowhat.cz", "Set DKIM domain")
	startCmd.Flags().StringVar(&dkimKeyPath, "dkim-key", ".", "Set path to DKIM key")
	startCmd.Flags().StringVar(&dkimSelector, "dkim-selector", "", "Set DKIM selector")
	startCmd.Flags().DurationVar(&writeTimeout, "timeout-write", 30*time.Second, "Set write timeout")
	startCmd.Flags().DurationVar(&readTimeout, "timeout-read", 30*time.Second, "Set read timeout")
}
