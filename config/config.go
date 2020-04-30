package config

import (
	"time"

	"github.com/spf13/viper"
)

// Parameters for target parameters
type Parameters struct {
	Config *viper.Viper
	Global *viper.Viper
	Target *viper.Viper
}

// GlobalConfig represents global config
type GlobalConfig struct {
	UserAgent string `mapstructure:"userAgent"`
}

// TargetConfig represents target params
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
			Login    map[string]string `mapstructure:"login"`
			Fixtures map[string]string `mapstructure:"fixtures"`
		} `mapstructure:"parser"`
	} `mapstructure:"client"`
}

// Init reads main config
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

// ParseConfig populates all config parameters
func (p *Parameters) ParseConfig(g *GlobalConfig, t *TargetConfig) error {

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

	return nil

}
