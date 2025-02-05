package config

import (
	"encoding/base32"
	"encoding/base64"
	"errors"
	"net/url"

	"github.com/caarlos0/env/v9"
	"github.com/rs/zerolog/log"
)

const (
	DefaultShamirKeyCount = 7
)

var Config struct {
	Mode                    string `envDefault:"dev"`
	DbUrl                   string
	KongUrl                 string
	RedisUrl                string
	NotificationUrl         string
	EmailWhitelist          []string
	ValidateEmailWhitelist  []string
	EmailServerNoReplyUrl   url.URL `env:"EMAIL_SERVER_NO_REPLY_URL"`
	EmailDomain             string
	EmailDev                string `envDefault:"dev@danta.tech"`
	ShamirFeature           bool   `envDefault:"true"`
	Standalone              bool
	VerificationCodeExpires int    `envDefault:"10"`
	SiteName                string `envDefault:"Open Tree Hole"`
	EnableRegisterQuestions bool   `envDefault:"false"`
}

var FileConfig struct {
	IdentifierSalt     string `env:"IDENTIFIER_SALT,file" envDefault:"/var/run/secrets/identifier_salt" default:""`
	ProvisionKey       string `env:"PROVISION_KEY,file" envDefault:"/var/run/secrets/provision_key" default:""`
	RegisterApikeySeed string `env:"REGISTER_APIKEY_SEED,file" envDefault:"/var/run/secrets/register_apikey_seed" default:""`
	KongToken          string `env:"KONG_TOKEN,file" envDefault:"/var/run/secrets/kong_token" default:""`
}

var DecryptedIdentifierSalt []byte
var RegisterApikeySecret string

func InitConfig() {
	var err error
	err = env.ParseWithOptions(&Config, env.Options{UseFieldNameByDefault: true})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	if Config.Mode == "production" {
		var innerErr error
		if Config.DbUrl == "" {
			innerErr = errors.Join(innerErr, errors.New("db url not set"))
		}
		if Config.EmailServerNoReplyUrl.String() == "" {
			innerErr = errors.Join(innerErr, errors.New("email server no reply url not set"))
		}
		if Config.EmailDomain == "" {
			innerErr = errors.Join(innerErr, errors.New("email domain not set"))
		}
		if innerErr != nil {
			log.Fatal().Err(innerErr).Send()
		}
	}

	log.Info().Any("config", Config).Send()

	initFileConfig()

	if FileConfig.IdentifierSalt == "" {
		if Config.Mode == "production" {
			log.Fatal().Msg("identifier salt not set")
		} else {
			DecryptedIdentifierSalt = []byte("123456")
		}
	} else {
		DecryptedIdentifierSalt, err = base64.StdEncoding.DecodeString(FileConfig.IdentifierSalt)
		if err != nil {
			log.Fatal().Err(err).Msg("decode identifier salt error")
		}
	}

	RegisterApikeySecret = base32.StdEncoding.EncodeToString([]byte(FileConfig.RegisterApikeySeed))
}

func initFileConfig() {
	err := env.Parse(&FileConfig)
	if err != nil {
		var e *env.AggregateError
		if errors.As(err, &e) {
			for _, innerErr := range e.Errors {
				var loadFileContentError env.LoadFileContentError
				if !errors.As(innerErr, &loadFileContentError) {
					log.Fatal().Err(err).Msg("init file config error")
				}
			}
		}
	}
}
