package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/dseries"
	"github.com/fpawel/gohelp"
	"github.com/fpawel/mil82/internal/api/notify"
	"github.com/fpawel/mil82/internal/api/types"
	"github.com/fpawel/mil82/internal/cfg"
)

func readProductVar(x worker, addr modbus.Addr, VarCode modbus.Var) (float64, error) {

	x.log = gohelp.LogPrependSuffixKeys(x.log, "адрес", addr, "var", VarCode)
	value, err := modbus.Read3BCD(x.log, x.ctx, x.portProducts, addr, VarCode)
	if err == nil {
		notify.ReadVar(nil, types.AddrVarValue{Addr: addr, VarCode: VarCode, Value: value})
		dseries.AddPoint(addr, VarCode, value)
		return value, nil
	}
	if !merry.Is(err, context.Canceled) {
		notify.AddrError(nil, types.AddrError{Addr: addr, Message: err.Error()})
	}
	return value, err
}

func blowGas(x worker, n int) error {
	if err := x.performf("включение клапана %d", n)(func(x worker) error {
		return performWithWarn(x, func() error {
			return switchGas(x, n)
		})
	}); err != nil {
		return err
	}
	return delayf(x, minutes(cfg.Get().BlowGasMinutes), "продувка ПГС%d", n)
}

func switchGas(x worker, n int) error {
	s := "отключить газ"
	if n != 0 {
		s = fmt.Sprintf("подать ПГС%d", n)
	}
	return x.perform(s, func(x worker) error {
		_, err := modbus.Request{
			Addr:     5,
			ProtoCmd: 0x10,
			Data: []byte{
				0, 0x10, 0, 1, 2, 0, byte(n),
			},
		}.GetResponse(log, x.ctx, x.portGas, nil)
		return err
	})
}
