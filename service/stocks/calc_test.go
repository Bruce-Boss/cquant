package stocks

import (
	"testing"
	"cquant/comm"
)

func TestFixed(t *testing.T) {
	conf := new(comm.Configure)
	conf.Parse("../../configure.toml")

	Fixed(conf)


}