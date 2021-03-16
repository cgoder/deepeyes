package gss

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	//Default www static file dir
	DefaultHTTPDir = "web"
)

// var config ConfigST
// var server ServerST

func (svr *ServerST) LoadConfig() {
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath("./")

	viper.SetDefault("Debug", "false")
	viper.SetDefault("LogLevel", "debug")

	viper.SetDefault("HTTPPort", ":8084")
	viper.SetDefault("HTTPDir", DefaultHTTPDir)
	viper.SetDefault("HTTPLogin", "admin")
	viper.SetDefault("HTTPPassword", "admin")
	viper.SetDefault("HTTPDebug", "true")
	viper.SetDefault("HTTPDemo", "true")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Errorln("init config error:", err)
		panic("init config error")
	}
	// log.Infoln("load config ok")

	err = viper.Unmarshal(&svr)
	if err != nil {
		log.Errorln("init config unmarshal error:", err)
		panic("init config unmarshal error")
	}

	if svr.Conf.Debug {
		log.Infof("config :", JsonFormat(&svr))
	}

	// init log
	level, _ := log.ParseLevel(svr.Conf.LogLevel)
	log.SetLevel(level)
	if !svr.Conf.Debug {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(false)

	// InitDB(config.DB)
	// config.MOD = strings.ToUpper(config.MOD)
}

func (svr *ServerST) SaveConfig() error {
	if err := viper.SafeWriteConfig(); err != nil {
		log.Errorln("save config fail!")
		return errors.New("save config fail!")
	}
	return nil
}

//ServerHTTPDir
func (svr *ServerST) ServerHTTPDir() string {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	if filepath.Clean(svr.Conf.HTTPDir) == "." {
		return DefaultHTTPDir
	}
	return filepath.Clean(svr.Conf.HTTPDir)
}

//ServerHTTPDebug read debug options
func (svr *ServerST) ServerHTTPDebug() bool {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	return svr.Conf.HTTPDebug
}

//ServerLogLevel read debug options
func (svr *ServerST) ServerLogLevel() string {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	return svr.Conf.LogLevel
}

//ServerHTTPDemo read demo options
func (svr *ServerST) ServerHTTPDemo() bool {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	return svr.Conf.HTTPDemo
}

//ServerHTTPLogin read Login options
func (svr *ServerST) ServerHTTPLogin() string {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	return svr.Conf.HTTPLogin
}

//ServerHTTPPassword read Password options
func (svr *ServerST) ServerHTTPPassword() string {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	return svr.Conf.HTTPPassword
}

//ServerHTTPPort read HTTP Port options
func (svr *ServerST) ServerHTTPPort() string {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	return svr.Conf.HTTPPort
}

//ServerRTSPPort read HTTP Port options
func (svr *ServerST) ServerRTSPPort() string {
	svr.mutex.RLock()
	defer svr.mutex.RUnlock()
	return svr.Conf.RTSPPort
}
