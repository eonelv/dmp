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
	RTXTo string
	MailTo string
	MailToDEV string
	MailMgr string
	ProductNum int
	ProjectNum int
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
	srvCfg.RTXTo = serverCfg["RTX_TO"]
	srvCfg.MailToDEV = serverCfg["MAIL_DEV"]
	srvCfg.isDebug, _ = strconv.ParseBool(serverCfg["IS_DEBUG"])
	srvCfg.MailMgr = serverCfg["MAIL_MGR"]
	srvCfg.ProductNum, _ = strconv.Atoi(serverCfg["PRODUCT_NUM"])
	srvCfg.ProjectNum, _ = strconv.Atoi(serverCfg["PROJECT_NUM"])

	return true, nil
}

func GetMailMgr() string {
	return srvCfg.MailMgr
}

func GetMailTo() string {
	return srvCfg.MailTo
}

func GetRTXTo() string {
	return srvCfg.RTXTo
}

func GetMailToDEV() string {
	return srvCfg.MailToDEV
}

func GetProductNum() int {
	return srvCfg.ProductNum
}

func GetProjectNum() int {
	return srvCfg.ProjectNum
}

func IsDebug() bool {
	return srvCfg.isDebug
}


