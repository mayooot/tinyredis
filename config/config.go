package config

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var Configures *Config
var (
	defaultHost     = "0.0.0.0"
	defaultPort     = 6379
	defaultLogDir   = "./"
	defaultLogLevel = "info"
	defaultShardNum = 1024
)

type Config struct {
	ConfFile string
	Host     string
	Port     int
	LogDir   string
	LogLevel string
	ShardNum int
}
type CfgError struct {
	message string
}

func (err *CfgError) Error() string {
	return err.message
}
func flagInit(cfg *Config) {
	flag.StringVar(&(cfg.ConfFile), "config", "", "Appoint a config file: such as /etc/redis.conf")
	flag.StringVar(&(cfg.Host), "host", defaultHost, "Bind host ip: default is 127.0.0.1")
	flag.IntVar(&(cfg.Port), "port", defaultPort, "Bind a listening port: default is 6379")
	flag.StringVar(&(cfg.LogDir), "logdir", defaultLogDir, "Set log directory: default is /tmp")
	flag.StringVar(&(cfg.LogLevel), "loglevel", defaultLogLevel, "Set log level.go: default is info")
}
func Setup() (*Config, error) {

	cfg := &Config{
		Host:     defaultHost,
		Port:     defaultPort,
		LogDir:   defaultLogDir,
		LogLevel: defaultLogLevel,
		ShardNum: defaultShardNum,
	}

	flagInit(cfg)
	flag.Parse()
	if cfg.ConfFile != "" {
		if err := cfg.Parse(cfg.ConfFile); err != nil {
			return nil, err
		}
	} else {
		if ip := net.ParseIP(cfg.Host); ip == nil {
			ipErr := &CfgError{
				message: fmt.Sprintf("Given ip address %s is invalid", cfg.Host),
			}
			return nil, ipErr
		}
		if cfg.Port <= 1024 || cfg.Port >= 65535 {
			portErr := &CfgError{
				message: fmt.Sprintf("Listening port should between 1024 and 65535, but %d is given.", cfg.Port),
			}
			return nil, portErr
		}
	}
	Configures = cfg
	return cfg, nil
}

func (cfg *Config) Parse(cfgFile string) error {
	fl, err := os.Open(cfgFile)
	if err != nil {
		return err
	}

	defer func() {
		err := fl.Close()
		if err != nil {
			fmt.Printf("Close config file error: %s \n", err.Error())
		}
	}()

	reader := bufio.NewReader(fl)
	for {
		line, ioErr := reader.ReadString('\n')
		if ioErr != nil && ioErr != io.EOF {
			return ioErr
		}

		if len(line) > 0 && line[0] == '#' {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			cfgName := strings.ToLower(fields[0])
			if cfgName == "host" {
				if ip := net.ParseIP(fields[1]); ip == nil {
					ipErr := &CfgError{
						message: fmt.Sprintf("Given ip address %s is invalid", cfg.Host),
					}
					return ipErr
				}
				cfg.Host = fields[1]
			} else if cfgName == "port" {
				port, err := strconv.Atoi(fields[1])
				if err != nil {
					return err
				}
				if port <= 1024 || port >= 65535 {
					portErr := &CfgError{
						message: fmt.Sprintf("Listening port should between 1024 and 65535, but %d is given.", port),
					}
					return portErr
				}
				cfg.Port = port
			} else if cfgName == "logdir" {
				cfg.LogDir = strings.ToLower(fields[1])
			} else if cfgName == "loglevel" {
				cfg.LogLevel = strings.ToLower(fields[1])
			} else if cfgName == "shardnum" {
				cfg.ShardNum, err = strconv.Atoi(fields[1])
				if err != nil {
					fmt.Println("ShardNum should be a number. Get: ", fields[1])
					panic(err)
				}
			}
		}
		if ioErr == io.EOF {
			break
		}
	}
	return nil
}
