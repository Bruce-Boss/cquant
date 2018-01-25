package cmd

import (
	"github.com/urfave/cli"
	"github.com/datochan/gcom/logger"
	"github.com/datochan/ctdx"
	"cquant/comm"
)

var UpdateCommandList = []cli.Command{
	{
		Name:    "basics",
		Usage:   "更新沪深股债基列表信息",
		Action: Basics,
	}, {
		Name:    "bonus",
		Usage:   "更新沪深股票权息数据",
		Action: Bonus,
	}, {
		Name:    "days",
		Usage:   "更新股票日线的交易数据",
		Action: Days,
	}, {
		Name:    "mins",
		Usage:   "更新股票5分钟级别的交易数据",
		Action: Mins,
	}, {
		Name:    "report",
		Usage:   "更新股票财报信息",
		Action: Report,
	},
}

func Basics(c *cli.Context) error {
	logger.Info("准备更新市场基础数据...")
	configure := c.App.Metadata["configure"].(*comm.Configure)

	tdxClient := ctdx.NewDefaultTdxClient(configure)
	defer tdxClient.Close()

	tdxClient.Conn()
	tdxClient.UpdateStockBase()

	// 更新结束后会向管道中发送一个通知
	<- tdxClient.Finished
	logger.Info("市场基础数据更新完毕...")

	return nil
}

func Bonus(c *cli.Context) error {
	logger.Info("准备更新高送转数据...")
	configure := c.App.Metadata["configure"].(*comm.Configure)

	tdxClient := ctdx.NewDefaultTdxClient(configure)
	defer tdxClient.Close()

	tdxClient.Conn()
	tdxClient.UpdateStockBonus()

	// 更新结束后会向管道中发送一个通知
	<- tdxClient.Finished
	logger.Info("高送转数据更新完毕...")

	return nil
}

func Days(c *cli.Context) error {
	logger.Info("准备更新日线数据...")
	configure := c.App.Metadata["configure"].(*comm.Configure)

	tdxClient := ctdx.NewDefaultTdxClient(configure)
	defer tdxClient.Close()

	tdxClient.Conn()
	tdxClient.UpdateDays()

	// 更新结束后会向管道中发送一个通知
	<- tdxClient.Finished
	logger.Info("日线数据更新完毕...")

	return nil
}

func Mins(c *cli.Context) error {
	logger.Info("准备更新五分钟线数据...")
	configure := c.App.Metadata["configure"].(*comm.Configure)

	tdxClient := ctdx.NewDefaultTdxClient(configure)
	defer tdxClient.Close()

	tdxClient.Conn()
	tdxClient.UpdateMins()

	// 更新结束后会向管道中发送一个通知
	<- tdxClient.Finished
	logger.Info("五分钟线数据更新完毕...")

	return nil
}

func Report(c *cli.Context) error {
	logger.Info("准备更新财报数据...")
	configure := c.App.Metadata["configure"].(*comm.Configure)

	tdxClient := ctdx.NewDefaultTdxClient(configure)
	defer tdxClient.Close()

	tdxClient.UpdateReport()

	logger.Info("财报数据更新完毕...")

	return nil
}
