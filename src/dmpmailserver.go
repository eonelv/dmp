package main

import (
	"net/smtp"
	"strings"
	. "db"
	. "core"
	"time"
	"fmt"
	. "cfg"
	"os/exec"
)

const (
	SENDER string = "asbuilder@eone.com"
	PASS string = "123456"
	MAIL_SERVER string = "192.168.0.10:25"
	DURATION int64 = 30 * 60 *1000
	STROY_NAME_MODLE string = "<a href='http://192.168.0.10/zentaopms/www/index.php?m=story&f=view&storyID=%v'>%v</a>"
	MSG_MODLE string = "<a href='http://192.168.0.10/zentaopms/www/index.php?m=story&f=view&storyID=%v'>%v %v</a>"
	MSG_RTX_MODLE string = "%v%v"
	MSG_RTX_URL_MODLE string = "http://192.168.0.10/zentaopms/www/index.php?m=story&f=view&storyID=%v"

)

func main() {
	LoadCfg()
	CreateDBMgr("")
	timer := &Timer{}
	timer.Start(DURATION)
	ch := timer.GetChannel()
	var lastSendTime int
	var isSendWeek bool
	var isSendMonth bool
	go SendRTX("服务器已启动", "0", "liwei")
	for {
		select {
		case <- ch:
			timeNow := time.Now()
			hour := timeNow.Hour()
			weekDay := int(timeNow.Weekday())
			monthDay := timeNow.Day()

			LogInfo("当前时间为")
			if (9 == hour || 14 == hour) && lastSendTime != hour {
				lastSendTime = hour
				go prepareToSend()
				go prepareDevelop()
				go prepareSendBackStory()
				go prepareDevelopDeadline()
			}

			//如果是周六，统计本周关闭的单，可以验证的单，可测试的单，超期的任务，bug列表
			if 6 == weekDay && !isSendWeek{
				isSendWeek = true
				weekEmail()
			} else if isSendWeek && 6 != weekDay {
				isSendWeek = false
			}

			//如果是每个月的第一天，统计上月完成的需求及bug。 按单个人统计(总数)
			if 1 == monthDay && !isSendMonth{
				isSendMonth = true
				monthEmail()
			} else if isSendMonth && 1 != monthDay {
				isSendMonth = false
			}
		}
	}

}

func weekEmail() {

}

func monthEmail() {
	timeBegin := time.Now()
	timeBegin = timeBegin.AddDate(0, -1, 0)
	timeBegin = timeBegin.AddDate(0, 0, -timeBegin.Day()+1)
	du := time.Duration(-((timeBegin.Hour() *60 * 60 + timeBegin.Minute() * 60 + timeBegin.Second())*1e9 + timeBegin.Nanosecond()))
	timeBegin = timeBegin.Add(du)

	timeEnd := time.Now()
	timeEnd = timeEnd.AddDate(0, 0, -timeEnd.Day()+1)
	du = time.Duration(-(((timeEnd.Hour()) *60 * 60 + timeEnd.Minute() * 60 + timeEnd.Second())*1e9 + timeEnd.Nanosecond()))
	timeEnd = timeEnd.Add(du)

	var result string
	var sql string
	sql = "SELECT SUM(t0.estimate) as estimate, SUM(t0.consumed) as consumed, t0.finishedBy as finishedBy, t1.realname as realname "
	sql += "FROM zt_task t0 INNER JOIN zt_user t1 ON t0.finishedBy = t1.account WHERE t0.project > ? AND t0.status = 'closed' AND t0.closedReason = 'done'  AND DATE(t0.finishedDate) >= ? AND DATE(t0.finishedDate) <= ? "
	sql += "GROUP BY t0.finishedBy "
	rows, err := DBMgr.PreQuery(sql, GetProjectNum(), timeBegin.Format("2006-01-02 15:04:05"), timeEnd.Format(TIME_FORMAT_YYYYMMDDHHMMSS))
	if err != nil {
		LogError("查询数据为空")
	}
	result = "<table border='1'><tr><th>名称</th><th>姓名</th><th width='300'>预计完成工作量(H)</th><th width='300'>实际完成工作量(H)</th><th width='100'>修改bug(个)</th></tr>"
	var dataRows map[string]interface {} = make(map[string]interface {})
	for _, row := range rows {
		var dataRow map[string]interface {} = make(map[string]interface {})
		dataRow["finishedBy"] = row.GetString("finishedBy")
		dataRow["estimate"] = row.GetUint64("estimate")
		dataRow["consumed"] = row.GetUint64("consumed")
		dataRow["realname"] = row.GetString("realname")
		dataRows[row.GetString("finishedBy")] = dataRow
	}

	sql = "SELECT COUNT(1) as bugs, t0.resolvedBy as resolvedBy, t1.realname as realname from zt_bug t0 INNER JOIN zt_user t1 ON t0.resolvedBy = t1.account "
	sql += "WHERE t0.product > ? AND t0.project > ? AND DATE(t0.resolvedDate) >= ? AND DATE(t0.resolvedDate) <= ? GROUP BY t0.resolvedBy"
	rows, err = DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum(), timeBegin.Format(TIME_FORMAT_YYYYMMDDHHMMSS), timeEnd.Format(TIME_FORMAT_YYYYMMDDHHMMSS))
	if err != nil {
		LogError("查询数据为空")
	}
	for _, row2 := range rows {
		var dataRow map[string]interface {}
		var ok bool
		dataRow, ok = dataRows[row2.GetString("resolvedBy")].(map[string]interface {})
		if !ok {
			dataRow = make(map[string]interface {})
			dataRow["finishedBy"] = row2.GetString("resolvedBy")
			dataRow["realname"] = row2.GetString("realname")
		}

		dataRow["bugs"] = row2.GetUint64("bugs")
		dataRows[row2.GetString("resolvedBy")] = dataRow
	}

	for _, row1 := range dataRows {
		var tempRow map[string]interface {} = row1.(map[string]interface {})
		resultDev := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultDev = fmt.Sprintf(resultDev, tempRow["finishedBy"], tempRow["realname"],tempRow["estimate"], tempRow["consumed"], tempRow["bugs"])
		result += resultDev
	}
	result += "</table>"
	SendMail(GetMailMgr(), "月工作量汇总", result, "html")
}

/*
  发给具体执行的人，当天有需求需要研发完毕
 */
func prepareDevelopDeadline() {
	var result string
	var sql string
	sql = "SELECT t10.assignedTo assignedTo, t11.stroyid storyid, t11.storytitle storytitle FROM zt_task t10 INNER JOIN  "
	sql += "(SELECT t1.id as stroyid, t1.title as storytitle, MAX(t0.deadline) storydeadline FROM zt_task t0 INNER JOIN zt_story t1 on t0.story = t1.id where t0.project > ? AND t1.status <> 'closed' GROUP BY t1.id, t1.title) t11 "
	sql += "ON t10.story = t11.stroyid WHERE DATE(t11.storydeadline) = ?  ORDER BY t11.stroyid"

	timeBegin := time.Now()
	rows, err := DBMgr.PreQuery(sql, GetProjectNum(), timeBegin.Format(TIME_FORMAT_YYYYMMDD))
	if err != nil {
		LogError("查询数据为空")
	}

	var needClosed bool = false
	for _, row := range rows {
		storyName := fmt.Sprintf(STROY_NAME_MODLE, row.GetString("storyid"), row.GetString("storytitle"))
		msg := fmt.Sprintf(MSG_MODLE, row.GetString("storyid"), timeBegin.Format(TIME_FORMAT_YYYYMMDD), row.GetString("storytitle")+"需要研发完毕")
		msgRTX := fmt.Sprintf(MSG_RTX_MODLE, timeBegin.Format(TIME_FORMAT_YYYYMMDD), row.GetString("storytitle")+"需要研发完毕")
		result = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
		resultDev := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultDev = fmt.Sprintf(resultDev, row.GetString("storyid"), storyName, msg)
		result = result + resultDev
		result += "</table><br>"
		go SendMail(row.GetString("assignedTo") + "@eone.com", "你有需求需要今天(" + timeBegin.Format(TIME_FORMAT_YYYYMMDD) + ")研发完毕", result, "html")
		go SendRTX(msgRTX, row.GetString("storyid"), row.GetString("assignedTo"))
		needClosed = true
	}
	if needClosed {
		go SendMail("nagrand@eone.com", "今天(" + timeBegin.Format(TIME_FORMAT_YYYYMMDD) + ")有需求要研发完毕，留下来加班吧^_^", result, "html")
		go SendRTX("今天(" + timeBegin.Format(TIME_FORMAT_YYYYMMDD) + ")有需求要研发完毕，留下来加班吧^_^", "0", "nagrand")
	}
}

/*
单独发给具体的相关人， 需求被打回了
 */
func prepareSendBackStory() {
	var result string
	var sql string
	sql = "SELECT t0.finishedBy finishedBy, t1.id ID, t1.title name from zt_task t0 RIGHT JOIN zt_story t1 ON t0.story = t1.id "
	sql += "INNER JOIN zt_projectStory t2 ON t2.story = t1.id "
	sql += "WHERE t1.product > ? AND t2.project > ? AND t1.status <> 'closed' AND t1.maustage in (8) AND t1.deleted <> '1' ORDER BY t1.id"
	rows, err := DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
	if err != nil {
		LogError("查询数据为空")
	}

	for _, row := range rows {
		storyName := fmt.Sprintf(STROY_NAME_MODLE, row.GetString("ID"), row.GetString("name"))

		result = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
		resultDev := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultDev = fmt.Sprintf(resultDev, row.GetString("ID"), storyName, "策划<font color='#ff0000'>验证不通过</font>")
		result = result + resultDev
		result += "</table><br>"
		go SendMail(row.GetString("finishedBy") + "@eone.com", "跟你相关的任务被策划打回了", result, "html")

		msg := fmt.Sprintf(MSG_RTX_MODLE, row.GetString("storytitle"), "被打回了")
		go SendRTX(msg, row.GetString("ID"), row.GetString("assignedTo"))
	}
}

func prepareDevelop() {
	var result string
	//可进行开发的需求
	sql := fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product = > AND t2.project = > AND t1.status in ('active') AND t1.stage NOT IN ('developed','tested') AND t1.deleted <> '1' ORDER BY t1.id")
	rows, err := DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
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

/*
发给负责人，dss。carreyzhou，acai，peter。 被打回的需求，可以测试的需求，可以验证的需求
 */
func prepareToSend() {
	var result string

	result = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	//验证不通过的需求
	sql := fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product > ? AND t2.project > ? AND t1.status <> 'closed' AND t1.maustage in (8) AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ := DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
	for _,row := range rows {
		storyName := fmt.Sprintf(STROY_NAME_MODLE, row.GetString("ID"), row.GetString("name"))
		resultBack := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultBack = fmt.Sprintf(resultBack, row.GetString("ID"), storyName, "<font color='#ff0000'>被打回,当前处于验证不通过状态</font>")
		result = result + resultBack
	}
	result += "</table><br>"

	result += "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	//可以测试的需求
	sql = fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product > ? AND t2.project > ? AND t1.status <> 'closed' AND t1.maustage in (7) AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ = DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
	for _,row := range rows {
		storyName := fmt.Sprintf(STROY_NAME_MODLE, row.GetString("ID"), row.GetString("name"))
		resultBack := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultBack = fmt.Sprintf(resultBack, row.GetString("ID"), storyName, "<font color='#0000ff'>验证通过，可以进行测试</font>")
		result = result + resultBack
	}
	result += "</table><br>"

	result += "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	//可以验证的需求
	sql = fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product > ? AND t2.project > ? AND t1.status <> 'closed' AND t1.maustage not in (4,5,6,7,8) AND t1.stage = 'developed' AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ = DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
	for _,row := range rows {
		storyName := fmt.Sprintf(STROY_NAME_MODLE, row.GetString("ID"), row.GetString("name"))
		resultBack := "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
		resultBack = fmt.Sprintf(resultBack, row.GetString("ID"), storyName, "<font color='#0000ff'>开发完成，可以进行验证</font>")
		result = result + resultBack
	}
	result += "</table><br>"
	SendMail(GetMailTo(), "每日需求提醒", result, "html")
	go SendRTX("您有新的邮件，请注意查收", "0", GetRTXTo())
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

func SendRTX(msg string, storyID string,  receiver string) {
	cmd := exec.Command("rtxproxy.cmd", msg, storyID, receiver, "DMP提醒")
	LogInfo("PARAMS", msg, storyID, receiver)
	bytes2, err2 := cmd.Output()
	if err2 == nil {
		if len(bytes2) != 0 {
			LogInfo(string(bytes2))
		}
	} else {
		LogError(err2, string(bytes2))
	}
}
