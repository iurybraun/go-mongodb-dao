package config

import (
	"gopkg.in/ini.v1"
	"fmt"
	"github.com/c-jimin/codetech/utils/logger"
	"strconv"
	"github.com/c-jimin/codetech/utils/args"
)

type MysqlConfig struct {
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
	Database string `ini:"database"`
	Options  string `ini:"options"`
}

type MongoDBConfig struct {
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
	Database string `ini:"database"`
	Options  string `ini:"options"`
}

var (
	MysqlUri                string
	MysqlUriWithoutDBName   string
	MysqlPWD                string
	MysqlDBName             string
	MongoDBUri              string
	MongoDBName             string
	MongoDBPWD              string
	SessionGCTime                   = 600
	OfflineTime             float64 = 3600
	HttpPort                        = 8000
	Debug                           = true
	FileSystemAllowTypeList         = []string{"image/*"}
	FileSystemAllowSize     int64   = 10 * 1024 * 1024 //单位：Byte
)

func init() {
	my := &MysqlConfig{
		Host:     "localhost",
		Port:     "3306",
		Username: "root",
		Password: "123456",
		Database: "codetech",
		Options:  "charset=utf8&parseTime=true&loc=Local",
	}

	mon := &MongoDBConfig{
		Host:     "localhost",
		Port:     "27017",
		Username: "",
		Password: "",
		Database: "codetech",
		Options:  "gssapiServiceName=mongodb",
	}

	if args.ConfigFilePath == "" {
		MysqlUri = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", my.Username, my.Password, my.Host, my.Port, my.Database, my.Options)
		MongoDBUri = fmt.Sprintf("mongodb://%s:%s@%s:%s/?%s", mon.Username, mon.Password, mon.Host, mon.Port, mon.Options)
		MysqlUriWithoutDBName = fmt.Sprintf("%s:%s@tcp(%s:%s)/", my.Username, my.Password, my.Host, my.Port)
		MysqlDBName = my.Database
		MongoDBName = mon.Database
		MysqlPWD = my.Password
		MongoDBPWD = mon.Password
		return
	}
	cfg, err := ini.Load(args.ConfigFilePath)
	if err != nil {
		logger.Failed("加载配置文件失败", err)
	}
	cfg.BlockMode = false
	// 配置Mysql

	if cfg.Section("Mysql").HasKey("host") {
		my.Host = cfg.Section("Mysql").Key("host").Value()
	}
	if cfg.Section("Mysql").HasKey("port") {
		my.Port = cfg.Section("Mysql").Key("port").Value()
	}
	if cfg.Section("Mysql").HasKey("username") {
		my.Username = cfg.Section("Mysql").Key("username").Value()
	}
	if cfg.Section("Mysql").HasKey("password") {
		my.Password = cfg.Section("Mysql").Key("password").Value()
		MysqlPWD = my.Password
	}
	if cfg.Section("Mysql").HasKey("database") {
		my.Database = cfg.Section("Mysql").Key("database").Value()
		MysqlDBName = my.Database
	}
	MysqlUri = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", my.Username, my.Password, my.Host, my.Port, my.Database, my.Options)
	MysqlUriWithoutDBName = fmt.Sprintf("%s:%s@tcp(%s:%s)/", my.Username, my.Password, my.Host, my.Port)

	if cfg.Section("MongoDB").HasKey("host") {
		mon.Host = cfg.Section("MongoDB").Key("host").Value()
	}
	if cfg.Section("MongoDB").HasKey("port") {
		mon.Port = cfg.Section("MongoDB").Key("port").Value()
	}
	if cfg.Section("MongoDB").HasKey("username") {
		mon.Username = cfg.Section("MongoDB").Key("username").Value()
	}
	if cfg.Section("MongoDB").HasKey("password") {
		mon.Password = cfg.Section("MongoDB").Key("password").Value()
		MongoDBPWD = mon.Password
	}
	if cfg.Section("MongoDB").HasKey("database") {
		mon.Database = cfg.Section("MongoDB").Key("database").Value()
	}
	MongoDBUri = fmt.Sprintf("mongodb://%s:%s@%s:%s/?%s", mon.Username, mon.Password, mon.Host, mon.Port, mon.Options)
	MongoDBName = mon.Database

	if cfg.Section("settings").HasKey("sessionGCTime") {
		SessionGCTime, err = strconv.Atoi(cfg.Section("settings").Key("sessionGCTime").Value())
		if err != nil {
			logger.Failed("sessionGCTime配置错误", err)
		}
	}

	if cfg.Section("settings").HasKey("offlineTime") {
		OfflineTime, err = strconv.ParseFloat(cfg.Section("settings").Key("offlineTime").Value(), 64)
		if err != nil {
			logger.Failed("offlineTime配置错误", err)
		}
	}
	if cfg.Section("settings").HasKey("httpPort") {
		HttpPort, err = strconv.Atoi(cfg.Section("settings").Key("httpPort").Value())
		if err != nil {
			logger.Failed("httpPort配置错误", err)
		}
	}
	if cfg.Section("settings").HasKey("debug") {
		Debug, err = strconv.ParseBool(cfg.Section("settings").Key("debug").Value())
		if err != nil {
			logger.Failed("debug配置错误", err)
		}
	}
}
