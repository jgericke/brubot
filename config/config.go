package config

import (
	"time"

	"github.com/spf13/viper"
)

// Parameters for target parameters
type Parameters struct {
	Config  *viper.Viper
	Global  *viper.Viper
	Target  *viper.Viper
	Sources *viper.Viper
}

// GlobalConfig maps to global config stanza
type GlobalConfig struct {
	DB struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Name     string `mapstructure:"name"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"db"`
	UserAgent string `mapstructure:"userAgent"`
}

// TargetConfig maps to target config stanza
type TargetConfig struct {
	UseGlobals bool `mapstructure:"useGlobals"`
	Auth       struct {
		URL            string            `mapstructure:"url"`
		Parameters     map[string]string `mapstructure:"parameters"`
		PasswordEncode bool              `mapstructure:"passwordEncode"`
		Method         string            `mapstructure:"method"`
		UserAgent      string            `mapstructure:"userAgent"`
		ErrorMsg       string            `mapstructure:"errorMsg"`
		Timeout        time.Duration     `mapstructure:"timeout"`
		Headers        map[string]string `mapstructure:"headers"`
	} `mapstructure:"auth"`
	Client struct {
		UserAgent           string            `mapstructure:"userAgent"`
		IgnoreRobots        bool              `mapstructure:"ignoreRobots"`
		EnableCache         bool              `mapstructure:"enableCache"`
		CacheDir            string            `mapstructure:"cacheDir"`
		DialTimeout         time.Duration     `mapstructure:"dialTimeout"`
		TLSHandShakeTimeout time.Duration     `mapstructure:"tlsHandShakeTimeout"`
		URLs                map[string]string `mapstructure:"urls"`
		Parser              struct {
			Login       map[string]string `mapstructure:"login"`
			Fixtures    map[string]string `mapstructure:"fixtures"`
			Predictions map[string]string `mapstructure:"predictions"`
		} `mapstructure:"parser"`
	} `mapstructure:"client"`
}

// SourcesConfig holds settings for every defined source (s1..sX)
// Sources slice unmarshals from sources.endpoints list
type SourcesConfig struct {
	Sources []struct {
		Name       string  `mapstructure:"name"`
		Tournament string  `mapstructure:"tournament"`
		Weight     float64 `mapstructure:"weight"`
		UseGlobals bool    `mapstructure:"useGlobals"`
		Client     struct {
			UserAgent           string            `mapstructure:"userAgent"`
			IgnoreRobots        bool              `mapstructure:"ignoreRobots"`
			EnableCache         bool              `mapstructure:"enableCache"`
			CacheDir            string            `mapstructure:"cacheDir"`
			DialTimeout         time.Duration     `mapstructure:"dialTimeout"`
			TLSHandShakeTimeout time.Duration     `mapstructure:"tlsHandShakeTimeout"`
			URLs                map[string]string `mapstructure:"urls"`
			Parser              struct {
				Predictions map[string]string `mapstructure:"predictions"`
			} `mapstructure:"parser"`
		} `mapstructure:"client"`
	} `mapstructure:"endpoints"`
}

// Init parses config file
func (p *Parameters) Init() error {

	p.Config = viper.New()
	p.Config.SetConfigType("yaml")
	p.Config.SetConfigName("config")
	p.Config.AddConfigPath(".")

	if err := p.Config.ReadInConfig(); err != nil {
		return err
	}
	return nil

}

// ParseConfig parses and populates all config parameters
// unmarshalling config parameters for globals, targets and sources
// into their respective Config structs
func (p *Parameters) ParseConfig(g *GlobalConfig, t *TargetConfig, s *SourcesConfig) error {

	var err error

	p.Global = p.Config.Sub("global")

	err = p.Global.Unmarshal(&g)
	if err != nil {
		return err
	}

	p.Target = p.Config.Sub("target")

	err = p.Target.Unmarshal(&t)
	if err != nil {
		return err
	}

	p.Sources = p.Config.Sub("sources")

	err = p.Sources.Unmarshal(&s)
	if err != nil {
		return err
	}

	return nil

}
