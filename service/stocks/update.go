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

	if today.Month() > 3 { dateList = append(dateList, fmt.Sprintf("%d-03-31", today.Year()))}
	if today.Month() > 6 { dateList = append(dateList, fmt.Sprintf("%d-06-30", today.Year()))}
	if today.Month() > 9 { dateList = append(dateList, fmt.Sprintf("%d-09-30", today.Year()))}

	return dateList
}

// 获取服务器时间的js地址: "https://s.thsi.cn/js/chameleon/time." + parseInt((new Date).getTime() / 1200000) + ".js"
func reportTime(rtFilePath, dateStr string) {
	curPage := 1
	maxPage := 0
	var dateList [][]string

	formatStr := "http://data.10jqka.com.cn/financial/yypl/date/%s/board/ALL/field/stockcode/order/DESC/page/%d/ajax/1/"

	logger.Info(fmt.Sprintf("\t开始更新 %s 的财报披露时间...", dateStr))

	for {
		// 每次都得重新计算sessionId否则403
		session := akqj01.New10JQKASession()
		session.UpdateServerTime()
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
	if len(dateList) <= 0 { dateList = reportDateList() }

	for _, dateItem := range dateList {
		rtFilePath := fmt.Sprintf("%s%srt%s.csv", configure.App.DataPath, configure.Tdx.Files.StockReport,
			strings.Replace(dateItem, "-", "", -1))
		reportTime(rtFilePath, dateItem)
	}
}
