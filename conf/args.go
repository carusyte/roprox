package conf

import (
	"go/build"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Args Global Application Arguments
var Args Arguments

//Arguments arguments struct type
type Arguments struct {
	LogLevel               string  `mapstructure:"log_level"`
	MasterProxyAddr        string  `mapstructure:"master_proxy_addr"`
	HTTPRetry              int     `mapstructure:"http_retry"`
	ScannerPoolSize        int     `mapstructure:"scanner_pool_size"`
	ScannerMaxRetry        int     `mapstructure:"scanner_max_retry"`
	ProbeSize              int     `mapstructure:"probe_size"`
	ProbeInterval          int     `mapstructure:"probe_interval"`
	ProbeTimeout           int     `mapstructure:"probe_timeout"`
	EvictionTimeout        int     `mapstructure:"eviction_timeout"`
	EvictionInterval       int     `mapstructure:"eviction_interval"`
	EvictionScoreThreshold float32 `mapstructure:"eviction_score_threshold"`

	Logging struct {
		LogFilePath string `mapstructure:"log_file_path"`
	}

	DataSource struct {
		UserAgents        string `mapstructure:"user_agents"`
		UserAgentLifespan int    `mapstructure:"user_agent_lifespan"`
	}

	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Schema   string `mapstructure:"schema"`
		UserName string `mapstructure:"user_name"`
		Password string `mapstructure:"password"`
	}
	//TODO logrus log to file
}

func init() {
	setDefaults()
	viper.SetConfigName("roprox") // name of config file (without extension)
	gopath := os.Getenv("GOPATH")
	if "" == gopath {
		gopath = build.Default.GOPATH
	}
	viper.AddConfigPath(filepath.Join(gopath, "bin"))
	viper.AddConfigPath(".") // optionally look for config in the working directory
	// viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if err != nil {
		log.Panicf("config file error: %+v", err)
	}
	err = viper.Unmarshal(&Args)
	if err != nil {
		log.Panicf("config file error: %+v", err)
	}
	// log.Printf("Configuration: %+v", Args)
	//viper.WatchConfig()
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println("Config file changed:", e.Name)
	//})
	checkConfig()
}

func checkConfig() {
	//check if config parameters are valid
}

func setDefaults() {
	Args.LogLevel = "info"
}
