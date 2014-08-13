package cfg

import (
	"os"
	"bufio"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"fmt"
)

type ServerConfig struct {
	MailTo string
	MailToDEV string
	isDebug bool
}

var srvCfg ServerConfig
var serverCfg map[string]string
func LoadCfg() (bool, error) {
	userFile := "dmpmailserver.cfg"

	file,err := os.OpenFile(userFile, os.O_RDONLY, os.ModeAppend)

	if err != nil {
		return false, err
	}
	defer file.Close()

	serverCfg = make(map[string]string)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// expression match
		lineArray := strings.Split(line, "=")
		fmt.Println("解析配置文件", line)
		if len(lineArray) >= 2 {
			serverCfg[lineArray[0]] = lineArray[1]
		}
	}
	srvCfg = ServerConfig{}
	srvCfg.MailTo = serverCfg["MAIL_TO"]
	srvCfg.MailToDEV = serverCfg["MAIL_DEV"]
	srvCfg.isDebug, _ = strconv.ParseBool(serverCfg["IS_DEBUG"])

	return true, nil
}

func GetMailTo() string {
	return srvCfg.MailTo
}

func GetMailToDEV() string {
	return srvCfg.MailToDEV
}

func IsDebug() bool {
	return srvCfg.isDebug
}


