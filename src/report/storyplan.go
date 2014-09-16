package report

import (
	. "db"
	. "core"
	. "dmpmail"
	"time"
	"fmt"
)
func SyncStoryPlan() {
	sql := "INSERT INTO t_e_storyplan (storyid, storyname, developEndDate, planEndDate, planCloseDate, productID, projectID) "
	sql += "(SELECT t0.story, t1.title, NULL, NULL, NULL, t0.product, t0.project FROM zt_projectStory t0 INNER JOIN zt_story t1 ON t0.story = t1.id "
	sql += "WHERE t0.story NOT IN (SELECT storyid FROM t_e_storyplan) )"
	re, err := DBMgr.Execute(sql)
	if err != nil {
		LogError(err)
		return
	}
	rowCount, _ := re.RowsAffected()
	LogInfo("sync story plan:", rowCount)
}

func StoryClosePlan() {
	var tableBegin string = "<table border='1'><tr><th>需求ID</th><th width='300'>需求名称</th><th width='300'>描述</th></tr>"
	var tableEnd string = "</table><br>"
	var tableRow string = "<tr align='center'><td>%v</td><td>%v</td><td>%v</td></tr>"

	var result string
	var sql string
	sql = "SELECT CASE t2.status WHEN 'closed' THEN t2.finishedBy WHEN 'done' THEN t2.finishedBy ELSE t2.assignedTo END assignedTo, t0.id id, t0.title title, t0.openedBy openedBy FROM zt_story t0  "
	sql += "INNER JOIN t_e_storyplan t1 ON t0.id = t1.storyid  "
	sql += "INNER JOIN zt_task t2 ON t2.story = t0.id  "
	sql += "WHERE Date(t1.planCloseDate) = ?  "

	timeBegin := time.Now()
	var timeToday string = timeBegin.Format(TIME_FORMAT_YYYYMMDD)
	rows, err := DBMgr.PreQuery(sql, timeToday)
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
		storyID = row.GetString("id")
		storyTitle = row.GetString("title")
		assignedTo = row.GetString("assignedTo")

		storyName := fmt.Sprintf(STROY_NAME_MODLE, storyID, storyTitle)
		msg := fmt.Sprintf(MSG_MODLE, storyID, timeToday, storyTitle+"需要【关闭】")
		msgRTX := fmt.Sprintf(MSG_RTX_MODLE, timeToday, storyTitle+"需要【关闭】")
		result = tableBegin
		resultDev := tableRow
		resultDev = fmt.Sprintf(resultDev, storyID, storyName, msg)
		result = result + resultDev
		result += tableEnd

		assignedToStorys[storyID] = *row
		LogInfo("测试", storyID, row)
		mailReceiver = assignedTo + "@eone.com"

		go SendMail(mailReceiver, "DMP你有需求需要今天(" + timeToday + ")【关闭】", result, "html")
		go SendRTX(msgRTX, storyID, assignedTo)

		needClosed = true
	}
	if needClosed {
		//发给对应的策划
		var openedBy string
		var rtxReceive string
		for key, row := range assignedToStorys {

			storyID = key
			storyTitle = row.GetString("storytitle")
			openedBy = row.GetString("openedBy")
			mailReceiver = openedBy+"@eone.com"
			rtxReceive = openedBy

			msg := fmt.Sprintf(MSG_MODLE, storyID, timeToday, storyTitle+"需要【关闭】")
			result = tableBegin
			resultDev := tableRow
			resultDev = fmt.Sprintf(resultDev, storyID, storyID, msg)
			result = result+resultDev
			result += tableEnd
			go SendMail(mailReceiver, "DMP你有需求需要今天("+timeToday+")【关闭】", result, "html")

			msgRTX := fmt.Sprintf(MSG_RTX_MODLE, timeToday, storyTitle+"需要【关闭】")
			go SendRTX(msgRTX, storyID, openedBy)

			mailReceiver = "nagrand@eone.com,peter@eone.com"
			rtxReceive = "nagrand-peter"
			//每个需求都通知出版本的人及QA主管
			go SendMail(mailReceiver, "DMP今天("+timeToday+")有需求要【关闭】", result, "html")
			go SendRTX("今天("+timeToday+")有需求要【关闭】", storyID, rtxReceive)
		}
	}
}
