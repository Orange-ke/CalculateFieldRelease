package conf

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

var AppConfig *Config

type Config struct {
	Port                  string
	ReadBufferSize        int
	WriteBufferSize       int
	Mode                  string
	PhaseTemperatureFile  string
	PhysicalParameterFile string
	NozzleConfigFile      string
	CasterHomePath        string
}

func Init() {
	file, err := ini.Load("conf/config.ini")
	if err != nil {
		log.Fatalln("配置文件读取错误，请检查文件路径: ", err)
		return
	}

	loadCfg(file)
}

func loadCfg(file *ini.File) {
	AppConfig = &Config{
		Port:                  file.Section("app").Key("Port").MustString(":9000"),
		ReadBufferSize:        file.Section("app").Key("ReadBufferSize").MustInt(1024),
		WriteBufferSize:       file.Section("app").Key("WriteBufferSize").MustInt(1024),
		Mode:                  file.Section("app").Key("Mode").MustString("debug"),
		PhaseTemperatureFile:  file.Section("app").Key("PhaseTemperatureFile").MustString(""),
		PhysicalParameterFile: file.Section("app").Key("PhysicalParameterFile").MustString(""),
		NozzleConfigFile:      file.Section("app").Key("NozzleConfigFile").MustString(""),
		CasterHomePath:        file.Section("app").Key("CasterHomePath").MustString(""),
	}

	log.Info("配置参数：", AppConfig)
}
