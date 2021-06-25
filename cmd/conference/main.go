package main

import (
	"bytes"
	"os"

	"github.com/HunterGooD/webrtcGolangServer/util"
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
}

func loadEnv() error {
	envFile := packr.New("env", ".env")
	buff, err := envFile.Find(".env")
	if err != nil {
		util.Errorf("%v", err)
		return err
	}

	if envVar, err := godotenv.Parse(bytes.NewReader(buff)); err != nil {
		util.Errorf("%v", err)
		return err
	} else {
		for key, val := range envVar {
			os.Setenv(key, val)
		}
	}
	return nil
}
