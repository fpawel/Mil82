package app

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/mil82/internal/cfg"
	"github.com/fpawel/mil82/internal/last_party"
	"time"
)

type runner struct{}

func (_ runner) Cancel() {
	cancelFunc()
	log.Info("выполнение прервано")
}

func (_ runner) SkipDelay() {
	skipDelayFunc()
	log.Info("задержка прервана")
}

func (_ runner) RunMainWork() {
	runWork(ctxApp, true, "настройка МИЛ-82", func() error {

		if err := switchGas(1); err != nil {
			return err
		}

		if err := delay("продувка ПГС1", 2*time.Minute); err != nil {
			return err
		}
		return delay("продувка ПГС2", time.Minute)
	})
}

func (_ runner) RunReadVars() {

	runWork(ctxApp, true, "опрос", func() error {
		vars := cfg.Get().Vars
		for {
			products := last_party.CheckedProducts()
			if len(products) == 0 {
				return merry.New("для опроса необходимо установить галочку для как минимум одного прибора")
			}
		loopProducts:
			for _, p := range products {
				for _, v := range vars {
					_, err := readProductVar(p.Addr, v.Code)
					if err != nil {
						if merry.Is(err, comm.ErrProtocol) {
							continue loopProducts
						}
						return err
					}
				}
			}
		}
	})
}
