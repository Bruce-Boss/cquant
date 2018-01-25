package comm

import (
	"log"
	"github.com/BurntSushi/toml"
	"github.com/datochan/ctdx/comm"
)

type Configure struct {
	comm.Conf

	Extend struct {
		Files struct {
			StockDayFixed string `toml:"stock_day_fixed"`
			StockDayMaPrice string `toml:"stock_day_ma_price"`
			StockDayMaVolume string `toml:"stock_day_ma_volume"`
			StockDayPe string `toml:"stock_day_pe"`
			StockDayPb string `toml:"stock_day_pb"`
		} `toml:"files"`
	} `toml:"extend"`

	Db struct {
		Driver string `toml:"driver"`
		Source string `toml:"source"`
		Debug bool `toml:"debug"`
	} `toml:"db"`
}

func (c *Configure) loadDefaults() {
	// app
	c.App.Logger.Level = "INFO"
	c.App.Logger.Name = "cquant"
	c.App.Mode = "debug"
}

// Will try to parse TOML configuration file.
func (c *Configure) Parse(path string) {
	c.loadDefaults()
	if path == "" {
		log.Printf("Loaded configuration defaults")
		return
	}

	if _, err := toml.DecodeFile(path, c); err != nil {
		panic(err)
	}
}

func (c *Configure) GetApp() comm.CApp {
	return c.Conf.GetApp()
}

func (c *Configure) GetTdx() comm.CTdx {
	return c.Tdx
}
