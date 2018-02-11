package stocks

import (
	 "github.com/PuerkitoBio/goquery"
	"cquant/comm"
	"time"
	"fmt"
	"os"
	"strings"
	"github.com/datochan/gcom/logger"
	"strconv"
	"github.com/kniren/gota/dataframe"
	"github.com/kniren/gota/series"
	"github.com/datochan/gcom/utils"
	"github.com/datochan/gcom/cnet"
	"github.com/datochan/gcom/bytes"
	"cquant/comm/akqj01"
)

func reportDateList() []string {
	dateList := []string{"1998-06-30", "1998-12-31", "1999-06-30", "1999-12-31", "2000-06-30", "2000-12-31",
						 "2001-06-30", "2001-09-30", "2001-12-31",}

	today := time.Now()
	for idx:=2002; idx < today.Year(); idx++ {
		dateList = append(dateList, fmt.Sprintf("%d-03-31", idx))
		dateList = append(dateList, fmt.Sprintf("%d-06-30", idx))
		dateList = append(dateList, fmt.Sprintf("%d-09-30", idx))
		dateList = append(dateList, fmt.Sprintf("%d-12-31", idx))
	}

	// 超过3月份的就需要更新一季度
	target := fmt.Sprintf("%d0331", today.Year())
	if utils.StrToDate(target).Before(today)  {
		dateList = append(dateList, fmt.Sprintf("%d-03-31", today.Year()))
	}

	target = fmt.Sprintf("%d0630", today.Year())
	if utils.StrToDate(target).Before(today)  {
		dateList = append(dateList, fmt.Sprintf("%d-06-30", today.Year()))
	}

	target = fmt.Sprintf("%d0930", today.Year())
	if utils.StrToDate(target).Before(today)  {
		dateList = append(dateList, fmt.Sprintf("%d-09-30", today.Year()))
	}

	if today.Month() == 4 {
		// 当年的4月份也得更新年报
		dateList = append(dateList, fmt.Sprintf("%d-12-31", today.Year()-1))
	}

	return dateList
}

func reportTime(timeUrl, formatStr, rtFilePath, dateStr string) {
	curPage := 1
	maxPage := 0
	var dateList [][]string

	logger.Info(fmt.Sprintf("\t开始更新 %s 的财报披露时间...", dateStr))

	for {
		// 每次都得重新计算sessionId否则403
		session := akqj01.New10JQKASession()
		session.UpdateServerTime(timeUrl)
		strCookie := session.Encode()

		htmlCnt := cnet.HttpRequest(fmt.Sprintf(formatStr, dateStr, curPage), "",
			fmt.Sprintf("v=%s", strCookie), "", "")
		content := bytes.BytesToString(htmlCnt)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			logger.Fatal(err.Error())
		}

		// 获取当前页的数据信息
		doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
			codeStr := strings.TrimSpace(s.Find("td").Eq(1).Text())    // 股票代码
			first := strings.TrimSpace(s.Find("td").Eq(3).Text())      // 首次预约时间
			changed := strings.TrimSpace(s.Find("td").Eq(4).Text())    // 变更时间
			act := strings.TrimSpace(s.Find("td").Eq(5).Text())        // 实际披露时间

			first = strings.Replace(first, "-", "", -1)
			first = strings.Replace(first, "00000000", "", -1)
			changed = strings.Replace(changed, "-", "", -1)
			act = strings.Replace(act, "-", "", -1)

			dateList = append(dateList, []string{codeStr, first, changed, act})
		})

		if 1 == curPage {
			spanText := doc.Find("div.m-page.J-ajax-page span").Text()
			lastPage := strings.Split(spanText, "/")
			maxPage, _ = strconv.Atoi(lastPage[1])
		}

		if curPage > maxPage { break }

		logger.Info(fmt.Sprintf("\t\t已处理完 %s 财报的第 %d 页 数据, 共计 %d 页..", dateStr, curPage, maxPage))

		curPage ++
	}

	rtListDF := dataframe.LoadRecords(dateList, dataframe.DetectTypes(false), dataframe.DefaultType(series.String))
	rtListDF.SetNames("code", "first", "change", "act")

	sortedDf := rtListDF.Arrange(dataframe.Sort("code"))

	utils.WriteCSV(rtFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, &sortedDf)
	logger.Info(fmt.Sprintf("\t %s 的财报披露时间更新结束...", dateStr))
}

/**
 * 更新财报披露时间
 * 回算时可以更让回测结果更准确
 */
func ReportTime(configure *comm.Configure, dateList []string) {
	var resultDateList []string

	if len(dateList) > 0 {
		today := time.Now()
		for _, dateItem := range dateList {
			targetItem, _ := time.Parse("2006-01-02", dateItem)

			if today.Year() >= targetItem.Year() {
				q1 := fmt.Sprintf("%d0331", targetItem.Year())
				q2 := fmt.Sprintf("%d0630", targetItem.Year())
				q3 := fmt.Sprintf("%d0930", targetItem.Year())

				if utils.StrToDate(q3).Before(targetItem) || utils.StrToDate(q3).Equal(targetItem) {
					resultDateList = append(resultDateList, fmt.Sprintf("%d-09-30", targetItem.Year()))

				} else if  utils.StrToDate(q2).Before(targetItem) || utils.StrToDate(q2).Equal(targetItem) {
					resultDateList = append(resultDateList, fmt.Sprintf("%d-06-30", targetItem.Year()))

				} else if utils.StrToDate(q1).Before(targetItem) || utils.StrToDate(q1).Equal(targetItem) {
					resultDateList = append(resultDateList, fmt.Sprintf("%d-03-31", targetItem.Year()))
				} else {
					resultDateList = append(resultDateList, fmt.Sprintf("%d-12-31", targetItem.Year()-1))
				}

				if today.Year() == targetItem.Year() && today.Month() == 4 {
					// 当年的4月份也得更新年报
					resultDateList = append(resultDateList, fmt.Sprintf("%d-12-31", today.Year()-1))
				}

			} else {
				logger.Fatal("指定的年报披露日期有误,不能超过当前日期!")
				return
			}
		}

	} else {
		resultDateList = reportDateList()
	}

	for _, dateItem := range resultDateList {
		rtFilePath := fmt.Sprintf("%s%srt%s.csv", configure.App.DataPath, configure.Thsi.Files.ReportTime,
			strings.Replace(dateItem, "-", "", -1))

		reportTime(configure.Thsi.Urls.ServerTime, configure.Thsi.Urls.ReportTime, rtFilePath, dateItem)
	}
}
