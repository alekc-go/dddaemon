package cloudflare

import (
	"errors"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Provider presents functionality for cloudflare operations
type Provider struct {
	CfApi   *cloudflare.API
	Records []string
}

// NewProvider creates a new instance of the cloudflare provider.
func NewProvider() (*Provider, error) {
	var err error
	prv := &Provider{}

	// check for valid config
	records := viper.GetString("CF_RECORDS")
	if records == "" {
		return nil, errors.New("you need to define CF_RECORDS variable")
	}
	prv.Records = strings.Split(records, ",")

	// are we authing with a token?
	if token := viper.GetString("CF_API_TOKEN"); token != "" {
		prv.CfApi, err = cloudflare.NewWithAPIToken(token)
		return prv, err
	}

	// check if we have both email and token for auth
	email := viper.GetString("CF_API_EMAIL")
	key := viper.GetString("CF_API_KEY")
	if email == "" || key == "" {
		return nil, errors.New("you need to define either CF_API_TOKEN, or both CF_API_EMAIL and CF_API_KEY")
	}

	prv.CfApi, err = cloudflare.New(key, email)
	if err != nil {
		return nil, err
	}

	return prv, nil
}

func (p *Provider) UpdateRecord(ip string) error {
	for _, record := range p.Records {
		// todo: cache zone info
		parts := strings.Split(record, ".")
		if len(parts) < 2 {
			return errors.New("invalid record supplied")
		}
		topDomain := strings.Join(parts[len(parts)-2:], ".")
		zone, err := p.CfApi.ZoneIDByName(topDomain)
		if err != nil {
			log.Fatal().Err(err).
				Str("record", record).
				Msg("cannot fetch zone info")
			return err
		}

		// create or update record
		// todo: cache record reference?
		zoneRecord, err := p.CfApi.DNSRecords(zone, cloudflare.DNSRecord{
			Name: record,
		})
		if err != nil {
			log.Fatal().Err(err).
				Str("record", record).
				Msg("cannot fetch zone zonerecord")
			return err
		}

		// if the record doesn't exist, create it
		if len(zoneRecord) == 0 {
			log.Debug().Str("record", record).Msg("record doesn't exist, trying to create it")
			//create the record
			_, err = p.CfApi.CreateDNSRecord(zone, cloudflare.DNSRecord{
				Type:    "A",
				Name:    record,
				Content: ip,
				Proxied: false,
				TTL:     120,
			})
			if err != nil {
				return err
			}
			continue
		}

		// record exists already. Alter it
		log.Debug().Str("record", record).Msg("record exists. Updating it.")
		updatedRecord := zoneRecord[0]
		updatedRecord.Content = ip
		err = p.CfApi.UpdateDNSRecord(zone, updatedRecord.ID, updatedRecord)
		if err != nil {
			log.Debug().Err(err).Str("record", record).Msg("cannot update dns record")
			return err
		}
		log.Debug().
			Str("record", record).
			Msg("record updated. Updating it.")
	}
	return nil
}
