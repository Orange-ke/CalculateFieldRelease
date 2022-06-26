package calculator

import "lz/model"

// calculator 的接口
type Calculator interface {
	// 构建温度场云图data
	BuildData() *TemperatureFieldData
	// 获取温度场计算器
	GetCalcHub() *CalcHub
	// 初始化钢种
	InitSteel(steelValue int, castingMachine *CastingMachine)
	// 初始化铸机
	InitCastingMachine()
	// 初始化推送数据容器
	InitPushData(coordinate model.Coordinate)
	// 获取钢种
	GetCastingMachine() *CastingMachine
	// 运行
	Run()
	// 设置拉尾坯
	SetStateTail() // todo
	// 获取温度场数组的大小
	GetFieldSize() int
	// 横切面数据
	GenerateSLiceInfo(index int) *SliceInfo
	// 纵切面曲线
	GenerateVerticalSlice1Data() *VerticalSliceData1
	// 纵切面云图
	GenerateVerticalSlice2Data(reqData model.VerticalReqData) *VerticalSliceData2
}
