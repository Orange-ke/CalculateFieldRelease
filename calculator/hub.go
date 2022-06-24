package calculator

type CalcHub struct {
	// 温度场推送
	Stop             chan struct{}
	PeriodCalcResult chan struct{}
}

func NewCalcHub() *CalcHub {
	return &CalcHub{
		PeriodCalcResult: make(chan struct{}),
	}
}

// 温度场计算
func (ch *CalcHub) PushSignal() {
	ch.PeriodCalcResult <- struct{}{}
}

func (ch *CalcHub) StopSignal() {
	close(ch.Stop)
}

func (ch *CalcHub) StartSignal() {
	ch.Stop = make(chan struct{})
}
