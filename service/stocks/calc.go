package stocks

import (
	"os"
	"fmt"
	"github.com/datochan/gcom/logger"
	"github.com/datochan/gcom/utils"
	tcomm "github.com/datochan/ctdx/comm"
	"github.com/kniren/gota/series"
	"github.com/kniren/gota/dataframe"

	"cquant/comm"
)

func fixedItem(configure *comm.Configure, bonusDF dataframe.DataFrame, market int, code string) {

	prevOpen := 0.0
	prevLow := 0.0
	prevHigh := 0.0
	prevClose := 0.0

	prevFixedOpen := 0.0
	prevFixedLow := 0.0
	prevFixedHigh := 0.0
	prevFixedClose := 0.0

	allFixedList := make([][]string, 0)

	// 得到指定股票的高送转信息
	filterDF := bonusDF.Filter(dataframe.F{Colname: "code", Comparator: series.Eq, Comparando: code},
	).Filter(dataframe.F{Colname: "type", Comparator: series.Eq, Comparando: 1},)
	filterDF.Arrange(dataframe.Sort("date"),)

	fileName := fmt.Sprintf("%d%s.csv.zip", market, code)
	stocksDayPath := fmt.Sprintf("%s%s%s", configure.GetApp().DataPath, configure.GetTdx().Files.StockDay, fileName)

	colTypes := map[string]series.Type{
		"date": series.String, "open": series.Float, "low": series.Float, "high": series.Float,
		"close": series.Float, "volume": series.Int, "amount": series.Float}

	stockDayDF := utils.ReadCSV(stocksDayPath, dataframe.WithTypes(colTypes))

	// 没有交易记录就不处理
	if stockDayDF.Err != nil || stockDayDF.Nrow() <= 0 { return }

	stocksFixedPath := fmt.Sprintf("%s%s%s", configure.GetApp().DataPath,
		configure.Extend.Files.StockDayFixed, fileName)

	fixedColType := map[string]series.Type{
		"code": series.String, "date": series.String, "open": series.Float, "low": series.Float,
		"high": series.Float, "close": series.Float,
	}

	fixedDF := utils.ReadCSV(stocksFixedPath, dataframe.WithTypes(fixedColType))
	if fixedDF.Err == nil && fixedDF.Nrow() > 0{
		// 已经计算过，就继续计算不需要重头算
		prevOpen = utils.Element(stockDayDF, fixedDF.Nrow()-1, "open").Float()
		prevLow = utils.Element(stockDayDF, fixedDF.Nrow()-1, "low").Float()
		prevHigh = utils.Element(stockDayDF, fixedDF.Nrow()-1, "high").Float()
		prevClose = utils.Element(stockDayDF, fixedDF.Nrow()-1, "close").Float()

		prevFixedOpen = utils.Element(fixedDF, fixedDF.Nrow()-1, "open").Float()
		prevFixedLow = utils.Element(fixedDF, fixedDF.Nrow()-1, "low").Float()
		prevFixedHigh = utils.Element(fixedDF, fixedDF.Nrow()-1, "high").Float()
		prevFixedClose = utils.Element(fixedDF, fixedDF.Nrow()-1, "close").Float()

		stockDayDF = stockDayDF.Filter(dataframe.F{Colname: "date", Comparator: series.Greater,
			Comparando: utils.Element(stockDayDF, fixedDF.Nrow()-1, "date").String()},)
	}

	for _, item := range stockDayDF.Maps() {
		money := 0.0  // 分红
		count := 0.0  // 送股数

		itemDF := filterDF.Filter(dataframe.F{Colname: "date", Comparator: series.Eq, Comparando: item["date"]})
		if itemDF.Nrow() > 0 {
			money = utils.Element(itemDF, 0, "money").Float() / 10
			count = utils.Element(itemDF, 0, "count").Float() / 10
		}

		// 除息除权日当天复权后的涨幅 =（当天不复权收盘价 *（1 + 每股送股数量）+每股分红金额） / 上一个交易日的不复权收盘价
		// 复权收盘价 = 上一个交易日的复权收盘价 *（1 + 复权涨幅)
		tmpOpen := item["open"].(float64)
		tmpLow := item["low"].(float64)
		tmpHigh := item["high"].(float64)
		tmpClose := item["close"].(float64)

		if prevOpen > 0 {
			dailyRateOpen := (tmpOpen * (1 + count) + money) / prevOpen
			dailyRateLow := (tmpLow * (1 + count) + money) / prevLow
			dailyRateHigh := (tmpHigh * (1 + count) + money) / prevHigh
			dailyRateClose := (tmpClose * (1 + count) + money) / prevClose

			prevFixedOpen *= dailyRateOpen
			prevFixedLow *= dailyRateLow
			prevFixedHigh *= dailyRateHigh
			prevFixedClose *= dailyRateClose

		} else {
			prevFixedOpen = tmpOpen
			prevFixedLow = tmpLow
			prevFixedHigh = tmpHigh
			prevFixedClose = tmpClose

		}

		prevOpen = tmpOpen
		prevLow = tmpLow
		prevHigh = tmpHigh
		prevClose = tmpClose

		allFixedList = append(allFixedList, []string{code, item["date"].(string),
			fmt.Sprintf("%f", prevFixedOpen),
			fmt.Sprintf("%f", prevFixedLow),
			fmt.Sprintf("%f", prevFixedHigh),
			fmt.Sprintf("%f", prevFixedClose),
		})
	}

	fixedResultDF := dataframe.LoadRecords(allFixedList, dataframe.WithTypes(fixedColType))

	isExist, _ := utils.FileExists(stocksFixedPath)
	if ! isExist {
		utils.WriteCSV(stocksFixedPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, &fixedResultDF)
	} else {
		utils.WriteCSV(stocksFixedPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, &fixedResultDF, dataframe.WriteHeader(false))
	}
}

// 计算后复权价
func Fixed(configure *comm.Configure) {
	baseDF := tcomm.GetFinanceDataFrame(configure, tcomm.STOCKA)
	if nil != baseDF.Err {
		logger.Error(fmt.Sprintf("读取股票基础数据失败! err:%v", baseDF))
		return
	}

	bonusPath := fmt.Sprintf("%s%s", configure.GetApp().DataPath, configure.GetTdx().Files.StockBonus)

	colTypes := map[string]series.Type{
		"code": series.String, "date": series.Int, "market": series.Int, "type": series.Int,
		"money": series.Float, "price": series.Float, "count": series.Float, "rate": series.Float}

	bonusDF := utils.ReadCSV(bonusPath, dataframe.WithTypes(colTypes))

	//fixedItem(configure, bonusDF, 1, "600000")
	for _, row := range baseDF.Maps() {
		logger.Info("开始计算 %d%s 的复权价...", row["market"].(int), row["code"].(string))
		fixedItem(configure, bonusDF, row["market"].(int), row["code"].(string))
	}
}
