package report

import (
	"fmt"
	"time"
	. "core"
	. "db"
	. "dmpmail"
	. "cfg"
)

func MonthEmail() {
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
