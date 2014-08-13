package main

import (
	"net/smtp"
	"strings"
	. "db"
	. "core"
	"time"
	"fmt"
	. "cfg"
)

const (
	SENDER string = "asbuilder@eone.com"
	PASS string = "123456"
	MAIL_SERVER string = "192.168.0.10:25"
)

func main() {
	LoadCfg()
	CreateDBMgr("")
	timer := &Timer{}
	timer.Start(30 * 60 * 1000)
	ch := timer.GetChannel()
	var lastSendTime int
	for {
		select {
		case <- ch:
			hour := time.Now().Hour()
			LogInfo("当前时间为", hour)
			if (9 == hour || 14 == hour) && lastSendTime != hour {
				lastSendTime = hour
				go prepareToSend()
				go prepareDevelop()
				go prepareSendBackStory()
			}
		}
	}

}

func prepareSendBackStory() {
	var result string
	var sql string
	sql = "SELECT t0.finishedBy finishedBy, t1.id ID, t1.title name from zt_task t0 RIGHT JOIN zt_story t1 ON t0.story = t1.id "
	sql += "INNER JOIN zt_projectStory t2 ON t2.story = t1.id "
	sql += "WHERE t1.product = ? AND t2.project = ? AND t1.status <> 'closed' AND t1.maustage in (8) AND t1.deleted <> '1' ORDER BY t1.id"
	rows, err := DBMgr.PreQuery(sql, 5, 45)
	if err != nil {
		LogError("查询数据为空")
	}

	for _, row := range rows {
		result = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
		resultDev := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultDev = fmt.Sprintf(resultDev, row.GetString("ID"),row.GetString("name"), "策划<font color='#ff0000'>验证不通过</font>")
		result = result + resultDev
		result += "</table><br>"
		go SendMail(row.GetString("finishedBy") + "@eone.com", "跟你相关的任务被策划打回了", result, "html")
	}
}

func prepareDevelop() {
	var result string
	//可进行开发的需求
	sql := fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product = ? AND t2.project = ? AND t1.status in ('active') AND t1.stage NOT IN ('developed','tested') AND t1.deleted <> '1' ORDER BY t1.id")
	rows, err := DBMgr.PreQuery(sql, 5, 45)
	if err != nil {
		fmt.Println("查询数据为空")
		return
	}
	if len(rows) == 0 {
		return
	}
	result = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th><th>手动阶段</th></tr>"
	for _, row := range rows {
		resultDev := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>"
		var strMauStage string
		maustage := row.GetUint64("maustage")
		if maustage == uint64(1) {
			strMauStage = "等待"
		} else {
			strMauStage = "已立项"
		}
		resultDev = fmt.Sprintf(resultDev, row.GetString("ID"),row.GetString("name"), "可以<font color='#00ff00'>进行开发</font>", strMauStage)
		result = result + resultDev
	}
	result += "</table><br>"
	SendMail(GetMailToDEV(), "每日需求提醒-可开发", result, "html")
}

func prepareToSend() {
	var result string

	result = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	//验证不通过的需求
	sql := fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product = ? AND t2.project = ? AND t1.status <> 'closed' AND t1.maustage in (8) AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ := DBMgr.PreQuery(sql, 5, 45)
	for _,row := range rows {
		resultBack := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultBack = fmt.Sprintf(resultBack, row.GetString("ID"), row.GetString("name"), "<font color='#ff0000'>被打回,当前处于验证不通过状态</font>")
		result = result + resultBack
	}
	result += "</table><br>"

	result += "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	//可以测试的需求
	sql = fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product = ? AND t2.project = ? AND t1.status <> 'closed' AND t1.maustage in (7) AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ = DBMgr.PreQuery(sql, 5, 45)
	for _,row := range rows {
		resultBack := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultBack = fmt.Sprintf(resultBack, row.GetString("ID"), row.GetString("name"), "<font color='#0000ff'>验证通过，可以进行测试</font>")
		result = result + resultBack
	}
	result += "</table><br>"

	result += "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	//可以验证的需求
	sql = fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product = ? AND t2.project = ? AND t1.status <> 'closed' AND t1.maustage not in (4,5,6,7,8) AND t1.stage = 'developed' AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ = DBMgr.PreQuery(sql, 5, 45)
	for _,row := range rows {
		resultBack := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultBack = fmt.Sprintf(resultBack, row.GetString("ID"), row.GetString("name"), "<font color='#0000ff'>开发完成，可以进行验证</font>")
		result = result + resultBack
	}
	result += "</table><br>"
	SendMail(GetMailTo(), "每日需求提醒", result, "html")
}

func SendMail(to, subject, body, mailtype string) error {
	fmt.Println("mailto", to)
	hp := strings.Split(MAIL_SERVER, ":")
	auth := smtp.PlainAuth("", SENDER, PASS, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + SENDER + "<" + SENDER + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(MAIL_SERVER, auth, SENDER, send_to, msg)
	return err
}
