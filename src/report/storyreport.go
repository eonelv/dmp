package report

import (
	. "core"
	. "db"
	. "cfg"
	"fmt"
	"time"
	. "dmpmail"
)

const (
	STROY_NAME_MODLE string = "<a href='http://192.168.0.10/zentaopms/www/index.php?m=story&f=view&storyID=%v'>%v</a>"
	MSG_MODLE string = "<a href='http://192.168.0.10/zentaopms/www/index.php?m=story&f=view&storyID=%v'>%v %v</a>"
	MSG_RTX_MODLE string = "%v%v"
	MSG_RTX_URL_MODLE string = "http://192.168.0.10/zentaopms/www/index.php?m=story&f=view&storyID=%v"
)

/*
  发给具体执行的人，当天有需求需要研发完毕
 */
func PrepareDevelopDeadline() {
	var tableBegin string = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	var tableEnd string = "</table><br>"
	var tableRow string = "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"

	var result string
	var sql string
	sql = "SELECT CASE t10.status WHEN 'closed' THEN t10.finishedBy WHEN 'done' THEN t10.finishedBy ELSE t10.assignedTo END assignedTo, t11.stroyid storyid, t11.storytitle storytitle, t11.openedBy openedBy FROM zt_task t10 INNER JOIN  "
	sql += "(SELECT t1.id as stroyid, t1.title as storytitle, MAX(t0.deadline) storydeadline, t1.assignedTo assignedTo FROM zt_task t0 INNER JOIN zt_story t1 on t0.story = t1.id where t0.project > ? AND t1.status <> 'closed' GROUP BY t1.id, t1.title, t1.assignedTo) t11 "
	sql += "ON t10.story = t11.stroyid WHERE DATE(t11.storydeadline) = ?  ORDER BY t11.stroyid"

	timeBegin := time.Now()
	var timeToday string = timeBegin.Format(TIME_FORMAT_YYYYMMDD)
	rows, err := DBMgr.PreQuery(sql, GetProjectNum(), timeToday)
	if err != nil {
		LogError("查询数据为空")
		return
	}

	var needClosed bool = false
	var mailReceiver string
	var storyID string
	var storyTitle string
	var assignedTo string

	var assignedToStorys map[string]RowSet = make(map[string]RowSet)

	for _, row := range rows {
		storyID = row.GetString("storyid")
		storyTitle = row.GetString("storytitle")
		assignedTo = row.GetString("assignedTo")

		storyName := fmt.Sprintf(STROY_NAME_MODLE, storyID, storyTitle)
		msg := fmt.Sprintf(MSG_MODLE, storyID, timeToday, storyTitle+"需要研发完毕")
		msgRTX := fmt.Sprintf(MSG_RTX_MODLE, timeToday, storyTitle+"需要研发完毕")
		result = tableBegin
		resultDev := tableRow
		resultDev = fmt.Sprintf(resultDev, storyID, storyName, msg)
		result = result + resultDev
		result += tableEnd

		assignedToStorys[storyID] = *row
		mailReceiver = assignedTo + "@eone.com"

		go SendMail(mailReceiver, "DMP你有需求需要今天(" + timeToday + ")研发完毕", result, "html")
		go SendRTX(msgRTX, storyID, assignedTo)

		needClosed = true
	}
	if needClosed {
		//发给对应的策划
		var openedBy string
		for key, row := range assignedToStorys {
			storyID = key
			storyTitle = row.GetString("storytitle")
			openedBy = row.GetString("openedBy")
			mailReceiver = openedBy + "@eone.com"

			msg := fmt.Sprintf(MSG_MODLE, storyID, timeToday, storyTitle+"需要研发完毕")
			result = tableBegin
			resultDev := tableRow
			resultDev = fmt.Sprintf(resultDev, storyID, storyID, msg)
			result = result + resultDev
			result += tableEnd
			go SendMail(mailReceiver, "DMP你有需求需要今天(" + timeToday + ")研发完毕", result, "html")

			msgRTX := fmt.Sprintf(MSG_RTX_MODLE, timeToday, storyTitle+"需要研发完毕")
			go SendRTX(msgRTX, storyID, openedBy)

			//每个需求都通知出版本的人及QA主管
			go SendMail("nagrand@eone.com,peter@eone.com", "DMP今天(" + timeToday + ")有需求要研发完毕", result, "html")
			go SendRTX("今天(" + timeToday + ")有需求要研发完毕", storyID, "nagrand-peter")
		}
	}
}

/*
单独发给具体的相关人， 需求被打回了
 */
func PrepareSendBackStory() {
	var tableBegin string = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	var tableEnd string = "</table><br>"
	var tableRow string = "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"

	var result string
	var sql string
	sql = "SELECT t0.finishedBy finishedBy, t1.id ID, t1.title name from zt_task t0 RIGHT JOIN zt_story t1 ON t0.story = t1.id "
	sql += "INNER JOIN zt_projectStory t2 ON t2.story = t1.id "
	sql += "WHERE t1.product > ? AND t2.project > ? AND t1.status <> 'closed' AND t1.maustage in (8) AND t1.deleted <> '1' ORDER BY t1.id"
	rows, err := DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
	if err != nil {
		LogError("查询数据为空")
	}

	var storyID string
	var storyTitle string
	var assignedTo string
	for _, row := range rows {
		storyID = row.GetString("ID")
		storyTitle = row.GetString("name")
		assignedTo = row.GetString("finishedBy")

		storyName := fmt.Sprintf(STROY_NAME_MODLE, storyID, storyTitle)

		result = tableBegin
		resultDev := tableRow
		resultDev = fmt.Sprintf(resultDev, storyID, storyName, "策划<font color='#ff0000'>验证不通过</font>")
		result = result + resultDev
		result += tableEnd
		go SendMail(assignedTo + "@eone.com", "DMP跟你相关的任务被策划打回了", result, "html")

		msg := fmt.Sprintf(MSG_RTX_MODLE, storyTitle, "被打回了")
		go SendRTX(msg, storyID, assignedTo)
	}
}

func PrepareDevelop() {
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
func PrepareToSend() {
	var isNeedSend bool = false
	var result string
	var tableBegin string = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	var tableEnd string = "</table><br>"
	var tableRow string = "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"
	result = tableBegin
	//验证不通过的需求
	sql := fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product >= ? AND t2.project >= ? AND t1.status <> 'closed' AND t1.maustage in (8) AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ := DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())

	var storyID string
	var storyTitle string
	var assignTo string
	for _,row := range rows {
		storyID = row.GetString("ID")
		storyTitle = row.GetString("name")
		storyName := fmt.Sprintf(STROY_NAME_MODLE, storyID, storyTitle)
		resultBack := tableRow
		resultBack = fmt.Sprintf(resultBack, storyID, storyName, "<font color='#ff0000'>被打回,当前处于验证不通过状态</font>")
		result = result + resultBack

		isNeedSend = true
	}
	result += tableEnd

	result += tableBegin
	//可以测试的需求
	sql = fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product >= ? AND t2.project >= ? AND t1.status <> 'closed' AND t1.maustage in (7) AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ = DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
	for _,row := range rows {
		storyID = row.GetString("ID")
		storyTitle = row.GetString("name")

		storyName := fmt.Sprintf(STROY_NAME_MODLE, storyID, storyTitle)
		resultBack := tableRow
		resultBack = fmt.Sprintf(resultBack, storyID, storyName, "<font color='#0000ff'>验证通过，可以进行测试</font>")
		result = result + resultBack

		isNeedSend = true
	}
	result += tableEnd

	result += tableBegin
	//可以验证的需求
	sql = fmt.Sprintf("%v %v",
		"SELECT t1.id as ID, t1.title as name, t1.maustatus as maustatus, t1.maustage as maustage, t1.assignedTo as assignedTo FROM zt_story t1 INNER JOIN zt_projectStory t2 ON t1.id = t2.story",
		"WHERE t1.product >= ? AND t2.project >= ? AND t1.status <> 'closed' AND t1.maustage not in (4,5,6,7,8) AND t1.stage = 'developed' AND t1.deleted <> '1' ORDER BY t1.id")
	rows, _ = DBMgr.PreQuery(sql, GetProductNum(), GetProjectNum())
	for _,row := range rows {
		storyID = row.GetString("ID")
		storyTitle = row.GetString("name")
		assignTo = row.GetString("assignedTo")

		storyName := fmt.Sprintf(STROY_NAME_MODLE, storyID, storyTitle)
		resultBack := tableRow
		resultBack = fmt.Sprintf(resultBack, storyID, storyName, "<font color='#0000ff'>开发完成，可以进行验证</font>")
		result = result + resultBack

		//通知具体策划，有需求可以验证
		msg := fmt.Sprintf(MSG_MODLE, storyID, storyTitle, "研发完毕，可以进行验证")
		msgRTX := fmt.Sprintf(MSG_RTX_MODLE, storyTitle,"研发完毕，可以进行验证")
		resultSingle := tableBegin
		resultSingleRow := tableRow
		resultSingleRow = fmt.Sprintf(resultSingleRow, storyID, storyName, msg)
		resultSingle = resultSingle + resultSingleRow
		resultSingle += tableEnd
		SendMail(assignTo+"@eone.com", "DMP每日需求提醒", resultSingle, "html")
		go SendRTX(msgRTX, storyID, assignTo)

		isNeedSend = true
	}
	result += tableEnd
	if isNeedSend {
		SendMail(GetMailTo(), "DMP每日需求提醒", result, "html")
		go SendRTX("您有新的邮件，请注意查收", "0", GetRTXTo())
	}
}
