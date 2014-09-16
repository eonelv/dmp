package main

import (
	. "db"
	. "core"
	"time"
	. "cfg"
	. "dmpmail"
	. "report"
)

const (
	DURATION int64 = 30 * 60 *1000
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
				go SyncStoryPlan()

				go PrepareToSend()
				go PrepareDevelop()
				go PrepareSendBackStory()
				go PrepareDevelopDeadline()
				go StoryClosePlan()
			}

			//如果是周六，统计本周关闭的单，可以验证的单，可测试的单，超期的任务，bug列表
			if 6 == weekDay && !isSendWeek{
				isSendWeek = true
				WeekEmail()
			} else if isSendWeek && 6 != weekDay {
				isSendWeek = false
			}

			//如果是每个月的第一天，统计上月完成的需求及bug。 按单个人统计(总数)
			if 1 == monthDay && !isSendMonth{
				isSendMonth = true
				MonthEmail()
			} else if isSendMonth && 1 != monthDay {
				isSendMonth = false
			}
		}
	}

}


