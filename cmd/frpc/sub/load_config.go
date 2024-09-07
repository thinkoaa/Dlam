package sub

import (
	"fmt"
	"io"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

type DNATMapping struct {
	LocalPort  int `toml:"localPort"`
	RemotePort int `toml:"remotePort"`
}

type DNAT struct {
	RemoteIP string        `toml:"remoteIP"`
	Mappings []DNATMapping `toml:"mappings"`
}

type FOFA struct {
	Switch      string `toml:"switch"`
	APIUrl      string `toml:"apiUrl"`
	Email       string `toml:"email"`
	Key         string `toml:"key"`
	QueryString string `toml:"queryString"`
	ResultSize  int    `toml:"resultSize"`
}

type QUAKE struct {
	Switch      string `toml:"switch"`
	APIUrl      string `toml:"apiUrl"`
	Key         string `toml:"key"`
	QueryString string `toml:"queryString"`
	ResultSize  int    `toml:"resultSize"`
}

type HUNTER struct {
	Switch      string `toml:"switch"`
	APIUrl      string `toml:"apiUrl"`
	Key         string `toml:"key"`
	QueryString string `toml:"queryString"`
	ResultSize  int    `toml:"resultSize"`
}

type Config struct {
	MyPortList      []int    `toml:"myPortList"`
	InternetAddress []string `toml:"whoseInternetAddress"`
	DNATS           []DNAT   `toml:"dnats"`
	FOFA            FOFA     `toml:"FOFA"`
	QUAKE           QUAKE    `toml:"QUAKE"`
	HUNTER          HUNTER   `toml:"HUNTER"`
}

func LoadConfig() Config {
	file, err := os.ReadFile("config.toml")
	if err != nil {
		fmt.Printf("config.toml导入失败[%s]\n", err)
		os.Exit(1)
	}
	var config Config
	err = toml.Unmarshal(file, &config)
	if err != nil {
		fmt.Printf("config.toml配置项错误[%s]\n", err)
		os.Exit(1)
	}
	return config
}
func WriteConfig(config *Config) {

	copyFile("config.toml", "config.toml.backup")
	newContent, err := toml.Marshal(&config)
	if err != nil {
		fmt.Printf("Error marshaling TOML: %v\n", err)
		return
	}

	if err := os.WriteFile("config.toml", newContent, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

}

func copyFile(sourceFile string, destinationFile string) {

	src, err := os.Open(sourceFile)
	if err != nil {
		fmt.Printf("Error opening source file: %v\n", err)
		return
	}
	defer src.Close()

	dst, err := os.Create(destinationFile)
	if err != nil {
		fmt.Printf("Error creating destination file: %v\n", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		fmt.Printf("Error copying file: %v\n", err)
		return
	}

}
