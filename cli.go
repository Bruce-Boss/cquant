package main

import (
	"os"
	"fmt"
	"strings"
	"github.com/urfave/cli"

	"github.com/datochan/gcom/logger"
	tcomm "github.com/datochan/ctdx/comm"

	"cquant/cmd"
	"cquant/comm"
)

func main() {
	var configureFile string
	app := cli.NewApp()
	app.Name = "cquant"
	app.Usage = "个人量化投资工具"
	app.Version = "0.0.1"
	app.UsageText = "cli [global options] command [command options] [arguments...]"
	app.Authors = []cli.Author{cli.Author{Name:  "datochan",Email: "datochan@qq.com"}}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "configure.toml",
			Usage: "载入系统配合文件.",
			Destination: &configureFile,
		},
	}

	// 每个命令执行前，先加载配置信息
	app.Before = func(c *cli.Context) error {
		configure := new(comm.Configure)
		configure.Parse(configureFile)

		c.App.Metadata = map[string]interface{}{
			"configure": configure,
		}

		strLevel := strings.ToUpper(configure.App.Logger.Level)
		switch strLevel {
		case "DEBUG": logger.InitFileLog(c.App.Writer, configure.App.Logger.Name, logger.LvDebug)
		case "INFO": logger.InitFileLog(c.App.Writer, configure.App.Logger.Name, logger.LvInfo)
		case "WARN": logger.InitFileLog(c.App.Writer, configure.App.Logger.Name, logger.LvWarn)
		case "ERROR": logger.InitFileLog(c.App.Writer, configure.App.Logger.Name, logger.LvError)
		case "FATAL": logger.InitFileLog(c.App.Writer, configure.App.Logger.Name, logger.LvFatal)
		default:
			logger.InitFileLog(c.App.Writer, configure.App.Logger.Name, logger.LvWarn)
		}

		// 默认加载股票日历数据
		calendarPath := fmt.Sprintf("%s%s", configure.App.DataPath, configure.Tdx.Files.Calendar)
		_, err := tcomm.DefaultStockCalendar(calendarPath)

		if nil != err {
			logger.Error("%v", err)
			return err
		}

		return nil
	}

	app.After = func(c *cli.Context) error {
		// 清理股票数据日历
		_, err := tcomm.DefaultStockCalendar("")
		if nil != err {
			logger.Error(fmt.Sprintf("%v", err))
			return err
		}

		return nil
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "命令 %q 没有找到.\n", command)
		cli.ShowAppHelp(c)
	}

	app.Commands = []cli.Command{
		{
			Name:    "update",
			Usage:   "更新原始数据",
			Subcommands: cmd.UpdateCommandList,
		},
		{
			Name:    "convert",
			Usage:   "转换原始数据(未实现)",
			//Subcommands: cmd.UpdateCommandList,
		},
		{
			Name:    "calc",
			Usage:   "计算数据(未实现)",
			Subcommands: cmd.CalcCommandList,
		},
		{
			Name:    "monitor",
			Usage:   "行情监控服务(未实现)",
			//Subcommands: cmd.UpdateCommandList,
		},
		{
			Name:    "server",
			Usage:   "Web服务(未实现)",
			//Subcommands: cmd.UpdateCommandList,
		},


		////////// 数据转换及回测类命令  //////////////////////
		//{
		//	Name:    "basics",
		//	ArgsUsage:   "[arrgh]",
		//	Usage:   "更新沪深股票基础数据及权息数据",
		//  UsageText: "basics [command options] [arguments...]",
		//	Flags: []cli.Flag{
		//		cli.BoolFlag{Name: "name, namespace"},
		//	},
		//	Category: "数据更新",
		//	Action: cmd.Basics,
		//},

		////////// 分析类命令  //////////////////////



	}

	app.Run(os.Args)
}