package dddaemon

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.alekc.dev/dddaemon/provider/cloudflare"
	"go.alekc.dev/publicip"
)

type Server struct {
	provider Provider
	ip       string
}

func (s Server) Run() {
	instance := &Server{}
	instance.initProvider().execute()
}

func (s *Server) execute() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		currentIP, err := publicip.Get()
		if err != nil {
			log.Err(err).Msg("cannot obtain public ip")
		}
		// if ip didn't change, we do not need to do anything
		if currentIP == s.ip {
			log.Debug().Msg("ip is the same and doesn't need to be changed")
			continue
		}
		log.Debug().Str("old_ip", s.ip).
			Str("current_ip", currentIP).
			Msg("detected ip change")

		err = s.provider.UpdateRecord(currentIP)
		if err != nil {
			log.Err(err).Msg("could not update dns record")
			continue
		}
		s.ip = currentIP
	}
}
func (s *Server) initProvider() *Server {
	// load required provider
	var err error
	providerName := viper.GetString("provider")
	switch strings.ToLower(providerName) {
	case "cloudflare":
		s.provider, err = cloudflare.NewProvider()
	default:
		log.Fatal().Str("provider_name", providerName).Msg("Unknown provider selected")
	}
	if err != nil {
		log.Fatal().
			Err(err).
			Str("provider_name", providerName).
			Msg("cannot initialize required provider")
	}
	return s
}
