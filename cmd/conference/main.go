package main

import (
	"bytes"
	"os"

	"github.com/HunterGooD/webrtcGolangServer/internal/room"
	"github.com/HunterGooD/webrtcGolangServer/internal/server"
	"github.com/HunterGooD/webrtcGolangServer/internal/util"
	"github.com/gobuffalo/packr/v2"
	"github.com/joho/godotenv"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		err := loadEnv()
		if err != nil {
			// Load default setting
		}
	}

	roomManager := room.NewRoomManager()
	wsServer := server.NewSFUServer(roomManager.HandleNewWebSocket)
	sslCert := os.Getenv("cert")
	sslKey := os.Getenv("key")
	bindAddress := os.Getenv("HOST")

	config := server.DefaultConfig()
	config.Host = bindAddress
	if bindAddress == "" {
		config.Host = "127.0.0.1"
	}
	config.Port = port
	if port == "" {
		config.Port = "3000"
	}
	config.CertFile = sslCert
	config.KeyFile = sslKey
	wsServer.Bind(config)
}

func loadEnv() error {
	envFile := packr.New("env", ".env")
	buff, err := envFile.Find(".env")
	if err != nil {
		util.Errorf(err.Error())
		return err
	}

	if envVar, err := godotenv.Parse(bytes.NewReader(buff)); err != nil {
		util.Errorf(err.Error())
		return err
	} else {
		for key, val := range envVar {
			os.Setenv(key, val)
		}
	}
	return nil
}
