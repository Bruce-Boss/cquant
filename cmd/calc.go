package cmd

import (
	"github.com/urfave/cli"
	"github.com/datochan/gcom/logger"

	"cquant/comm"
	"cquant/service/stocks"
)

var CalcCommandList = []cli.Command{
	{
		Name:    "fixed",
		Usage:   "计算股票后复权价",
		Action: Fixed,
	},
}

func Fixed(c *cli.Context) error {
	logger.Info("开始计算股票后复权价...")
	configure := c.App.Metadata["configure"].(*comm.Configure)

	stocks.Fixed(configure)

	logger.Info("股票日线后复权价计算完毕...")
	return nil
}
