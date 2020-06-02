package main

import (
	"errors"
	"github.com/spf13/viper"
	"github.com/taalhach/restcron/server"
	"log"
)

const (
	cfgPath     = "."
	cfgFile= "config"
	cfgFileType     = "json"
)

type _config struct {
	Database struct{
		Url string
		User_name string
		Password string
 	}
	Server struct{
		Port string
	}
}

func main() {
	cfg,err:=readConfig()
	if err!=nil{
		log.Fatal(err)
	}
	s, err := server.NewServer(cfg.Database.Url, cfg.Database.User_name, cfg.Database.Password)
	if err != nil {
		log.Fatal(err)
	}
	s.RunServer(cfg.Server.Port)
}
func readConfig() (*_config,error) {
	viper.AddConfigPath(cfgPath)
	viper.SetConfigType(cfgFileType)
	viper.SetConfigName(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil,err
	}
	var cfg _config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil,err
	}
	if cfg.Database.Url==""{
		return nil,errors.New("database Url missing in config")
	}
	if cfg.Database.User_name==""{
		return nil,errors.New("database user_name missing in config")
	}
	if cfg.Server.Port==""{
		return nil,errors.New("server port missing in config")
	}
	return &cfg,nil
}