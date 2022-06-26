package calculator

import (
	log "github.com/sirupsen/logrus"
	"lz/model"
	"time"
)

type SlicePushDataStruct struct {
	Slice   [][]float32 `json:"slice"`
	Start   int         `json:"start"`
	End     int         `json:"end"`
	Current int         `json:"current"`
}

type TemperatureFieldData struct {
	XScale int    `json:"x_scale"`
	YScale int    `json:"y_scale"`
	ZScale int    `json:"z_scale"`
	Start  int    `json:"start"`   // 切片开始位置
	End    int    `json:"end"`     // 切片结束位置
	IsFull bool   `json:"is_full"` // 切片是否充满铸机
	IsTail bool   `json:"is_tail"` // 是否拉尾坯
	Sides  *Sides `json:"sides"`
}

type Sides struct {
	Up    [][]float32 `json:"up"`
	Left  [][]float32 `json:"left"`
	Right [][]float32 `json:"right"`
	Front [][]float32 `json:"front"`
	Back  [][]float32 `json:"back"`
	Down  [][]float32 `json:"down"`
}

type PushData struct {
	Top    Encoding
	Arc    Encoding
	Bottom Encoding
}

type Encoding struct {
	Start []byte
	Data  [][]byte
	Max   []byte
}

type Decoding struct {
	Start []int
	Data  [][]int
	Max   []int
}

var (
	UpLength   float32
	ArcLength  float32
	DownLength float32

	StepX = 2
	StepY = 1
	StepZ = 2

	width  int
	length int

	sides = &Sides{}
)

func initPushData(up, arc, down float32) {
	UpLength, ArcLength, DownLength = up, arc, down
	width = Width / YStep / StepY * 2
	length = Length / XStep / StepX * 2
	log.Debug("pushData:", width, length, ZLength/ZStep/StepZ)
	sides = &Sides{
		Up:    make([][]float32, width),
		Left:  make([][]float32, ZLength/ZStep/StepZ),
		Right: make([][]float32, ZLength/ZStep/StepZ),
		Front: make([][]float32, ZLength/ZStep/StepZ),
		Back:  make([][]float32, ZLength/ZStep/StepZ),
		Down:  make([][]float32, width),
	}

	for i := 0; i < width; i++ {
		sides.Up[i] = make([]float32, length)
		sides.Down[i] = make([]float32, length)
	}
	for i := 0; i < ZLength/ZStep/StepZ; i++ {
		sides.Left[i] = make([]float32, width)
		sides.Right[i] = make([]float32, width)
		sides.Front[i] = make([]float32, length)
		sides.Back[i] = make([]float32, length)
	}
}

func (c *calculatorWithArrDeque) BuildData() *TemperatureFieldData {
	temperatureData := &TemperatureFieldData{
		Sides: sides,
	}

	startTime := time.Now()
	startSlice := c.Field.GetSlice(0)
	EndSlice := c.Field.GetSlice(c.Field.Size() - 1)
	for y := Width/YStep - 1; y >= 0; y -= StepY {
		for x := Length/XStep - 1; x >= 0; x -= StepX {
			temperatureData.Sides.Up[width/2+y/StepY][length/2+x/StepX] = startSlice[y][x]
			temperatureData.Sides.Up[(width/2-1)-y/StepY][(length/2-1)-x/StepX] = startSlice[y][x]
			temperatureData.Sides.Up[width/2+y/StepY][(length/2-1)-x/StepX] = startSlice[y][x]
			temperatureData.Sides.Up[(width/2-1)-y/StepY][length/2+x/StepX] = startSlice[y][x]
		}
	}

	for z := c.Field.Size() - 1; z >= 0; z -= StepZ {
		slice := c.Field.GetSlice(z)
		for x := Length/XStep - 1; x >= 0; x -= StepX {
			temperatureData.Sides.Front[z/StepZ][length/2+x/StepX] = slice[Width/YStep-1][x]
			temperatureData.Sides.Front[z/StepZ][length/2-1-x/StepX] = slice[Width/YStep-1][x]

			temperatureData.Sides.Back[z/StepZ][length/2+x/StepX] = slice[Width/YStep-1][x]
			temperatureData.Sides.Back[z/StepZ][length/2-1-x/StepX] = slice[Width/YStep-1][x]
		}

		for y := Width/YStep - 1; y >= 0; y -= StepY {
			temperatureData.Sides.Left[z/StepZ][width/2+y/StepY] = slice[y][Length/XStep-1]
			temperatureData.Sides.Left[z/StepZ][width/2-1-y/StepY] = slice[y][Length/XStep-1]

			temperatureData.Sides.Right[z/StepZ][width/2+y/StepY] = slice[y][Length/XStep-1]
			temperatureData.Sides.Right[z/StepZ][width/2-1-y/StepY] = slice[y][Length/XStep-1]
		}
	}

	for y := Width/YStep - 1; y >= 0; y -= StepY {
		for x := Length/XStep - 1; x >= 0; x -= StepX {
			temperatureData.Sides.Down[width/2+y/StepY][length/2+x/StepX] = EndSlice[y][x]
			temperatureData.Sides.Down[(width/2-1)-y/StepY][(length/2-1)-x/StepX] = EndSlice[y][x]
			temperatureData.Sides.Down[width/2+y/StepY][(length/2-1)-x/StepX] = EndSlice[y][x]
			temperatureData.Sides.Down[(width/2-1)-y/StepY][length/2+x/StepX] = EndSlice[y][x]
		}
	}

	temperatureData.XScale = StepX
	temperatureData.YScale = StepY
	temperatureData.ZScale = StepZ
	temperatureData.Start = c.start
	temperatureData.End = c.end
	temperatureData.IsFull = c.Field.IsFull()
	temperatureData.IsTail = c.isTail
	log.Debug("build data cost: ", time.Since(startTime))
	return temperatureData
}

// 横切面推送数据
type SliceInfo struct {
	HorizontalSolidThickness  float32     `json:"horizontal_solid_thickness"`
	VerticalSolidThickness    float32     `json:"vertical_solid_thickness"`
	HorizontalLiquidThickness float32     `json:"horizontal_liquid_thickness"`
	VerticalLiquidThickness   float32     `json:"vertical_liquid_thickness"`
	Slice                     [][]float32 `json:"slice"`
	Length                    int         `json:"length"`
}

func (c *calculatorWithArrDeque) GenerateSLiceInfo(index int) *SliceInfo {
	return c.buildSliceGenerateData(index)
}

func (c *calculatorWithArrDeque) buildSliceGenerateData(index int) *SliceInfo {
	solidTemp := c.steel1.SolidPhaseTemperature
	liquidTemp := c.steel1.LiquidPhaseTemperature
	sliceInfo := &SliceInfo{}
	slice := make([][]float32, Width/YStep*2)
	for i := 0; i < len(slice); i++ {
		slice[i] = make([]float32, Length/XStep*2)
	}
	originData := c.Field.GetSlice(index)
	// 从右上角的四分之一还原整个二维数组
	for i := 0; i < Width/YStep; i++ {
		for j := 0; j < Length/XStep; j++ {
			slice[i][j] = originData[Width/YStep-1-i][Length/XStep-1-j]
		}
	}
	for i := 0; i < Width/YStep; i++ {
		for j := Length / XStep; j < Length/XStep*2; j++ {
			slice[i][j] = originData[Width/YStep-1-i][j-Length/XStep]
		}
	}
	for i := Width / YStep; i < Width/YStep*2; i++ {
		for j := Length / XStep; j < Length/XStep*2; j++ {
			slice[i][j] = originData[i-Width/YStep][j-Length/XStep]
		}
	}
	for i := Width / YStep; i < Width/YStep*2; i++ {
		for j := 0; j < Length/XStep; j++ {
			slice[i][j] = originData[i-Width/YStep][Length/XStep-1-j]
		}
	}
	sliceInfo.Slice = slice
	length := Length/XStep - 1
	width := Width/YStep - 1
	// 宽面
	j := width
	for j = width; j >= 0; j-- {
		if originData[j][0] > solidTemp {
			break
		}
	}
	if j == width {
		sliceInfo.VerticalSolidThickness = 0
	} else if j < 0 {
		sliceInfo.VerticalSolidThickness = float32(Width)
	} else {
		sliceInfo.VerticalSolidThickness = float32(YStep*(width-j)) + float32(YStep)*(c.steel1.SolidPhaseTemperature-originData[j+1][0])/(originData[j][0]-originData[j+1][0])
	}
	for j = width; j >= 0; j-- {
		if originData[j][0] > liquidTemp {
			break
		}
	}
	if j == width {
		sliceInfo.VerticalSolidThickness = 0
	} else if j < 0 {
		sliceInfo.VerticalLiquidThickness = float32(Width)
	} else {
		sliceInfo.VerticalLiquidThickness = float32(YStep*(width-j)) + float32(YStep)*(c.steel1.LiquidPhaseTemperature-originData[j+1][0])/(originData[j][0]-originData[j+1][0])
	}
	// 窄面
	i := length
	for i = length; i >= 0; i-- {
		if originData[0][i] > solidTemp {
			break
		}
	}
	if i == length {
		sliceInfo.HorizontalSolidThickness = 0
	} else if i < 0 || sliceInfo.VerticalSolidThickness == float32(Width)  {
		sliceInfo.HorizontalSolidThickness = float32(Length)
	} else {
		sliceInfo.HorizontalSolidThickness = float32(XStep*(length-i)) + float32(XStep)*(c.steel1.SolidPhaseTemperature-originData[0][i+1])/(originData[0][i]-originData[0][i+1])
	}
	i = length
	for i = length; i >= 0; i-- {
		if originData[0][i] > liquidTemp {
			break
		}
	}
	if i == length {
		sliceInfo.HorizontalLiquidThickness = 0
	} else if i < 0 || sliceInfo.VerticalLiquidThickness == float32(Width) {
		sliceInfo.HorizontalLiquidThickness = float32(Length)
	} else {
		sliceInfo.HorizontalLiquidThickness = float32(XStep*(length-i)) + float32(XStep)*(c.steel1.LiquidPhaseTemperature-originData[0][i+1])/(originData[0][i]-originData[0][i+1])
	}


	sliceInfo.Length = c.Field.Size()
	return sliceInfo
}

type VerticalSliceData1 struct {
	CenterOuter [][2]float32 `json:"center_outer"`
	CenterInner [][2]float32 `json:"center_inner"`

	EdgeOuter [][2]float32 `json:"edge_outer"`
	EdgeInner [][2]float32 `json:"edge_inner"`
}

// 纵切面曲线
func (c *calculatorWithArrDeque) GenerateVerticalSlice1Data() *VerticalSliceData1 {
	var index int
	res := &VerticalSliceData1{
		CenterOuter: make([][2]float32, 0),
		CenterInner: make([][2]float32, 0),
		EdgeOuter:   make([][2]float32, 0),
		EdgeInner:   make([][2]float32, 0),
	}
	step := 0
	c.Field.Traverse(func(z int, item *model.ItemType) {
		step++
		if step == 5 {
			index = Length/XStep - 1
			res.CenterOuter = append(res.CenterOuter, [2]float32{float32((z + 1) * model.ZStep), item[Width/YStep-1][Length/XStep-1-index]})
			res.CenterInner = append(res.CenterInner, [2]float32{float32((z + 1) * model.ZStep), item[0][Length/XStep-1-index]})

			index = 0
			res.EdgeOuter = append(res.EdgeOuter, [2]float32{float32((z + 1) * model.ZStep), item[Width/YStep-1][Length/XStep-1-index]})
			res.EdgeInner = append(res.EdgeInner, [2]float32{float32((z + 1) * model.ZStep), item[0][Length/XStep-1-index]})

			step = 0
		}
	}, 0, c.Field.Size())
	return res
}

type VerticalSliceData2 struct {
	Length        int         `json:"length"`
	VerticalSlice [][]float32 `json:"vertical_slice"`
	Solid         []float32   `json:"solid"`
	Liquid        []float32   `json:"liquid"`
	SolidJoin     Join        `json:"solid_join"`
	LiquidJoin    Join        `json:"liquid_join"`
}

type Join struct {
	IsJoin    bool `json:"is_join"`
	JoinIndex int  `json:"join_index"`
}

// 纵切面云图
func (c *calculatorWithArrDeque) GenerateVerticalSlice2Data(reqData model.VerticalReqData) *VerticalSliceData2 {
	solidTemp := c.steel1.SolidPhaseTemperature
	liquidTemp := c.steel1.LiquidPhaseTemperature
	index := reqData.Index
	zScale := reqData.ZScale // 拉坯方向的缩放比例
	res := &VerticalSliceData2{
		Length:        c.Field.Size(),
		VerticalSlice: make([][]float32, c.Field.Size()/zScale),
		Solid:         make([]float32, c.Field.Size()),
		Liquid:        make([]float32, c.Field.Size()),
	}

	for i := 0; i < len(res.VerticalSlice); i++ {
		res.VerticalSlice[i] = make([]float32, Width/YStep*2)
	}

	var temp float32
	var solidJoinSet, liquidJoinSet bool
	step := 0
	zIndex := 0
	c.Field.Traverse(func(z int, item *model.ItemType) {
		step++
		if step == zScale {
			for i := 0; i < Width/YStep; i++ {
				res.VerticalSlice[zIndex][Width/YStep+i] = item[i][Length/XStep-1-index]
			}
			for i := Width/YStep - 1; i >= 0; i-- {
				res.VerticalSlice[zIndex][Width/YStep-1-i] = item[i][Length/XStep-1-index]
			}
			step = 0
			zIndex++
		}
		i := Width/YStep - 1
		for i = Width/YStep - 1; i >= 0; i-- {
			temp = item[i][Length/XStep-1-index]
			if temp > solidTemp {
				break
			}
		}
		if i == Width/YStep-1 {
			res.Solid[z] = 0
		} else if i < 0 {
			res.Solid[z] = float32(Width / YStep)
		} else {
			res.Solid[z] = float32(Width/YStep-1-i) + (solidTemp-item[i+1][Length/XStep-1-index])/(item[i][Length/XStep-1-index]-item[i+1][Length/XStep-1-index])
		}
		if res.Solid[z] >= float32(Width/YStep) && !solidJoinSet {
			res.Solid[z] = float32(Width / YStep)
			res.SolidJoin.IsJoin = true
			res.SolidJoin.JoinIndex = z
			solidJoinSet = true
		}

		i = Width/YStep - 1
		for i = Width/YStep - 1; i >= 0; i-- {
			temp = item[i][Length/XStep-1-index]
			if temp > liquidTemp {
				break
			}
		}
		if i == Width/YStep-1 {
			res.Liquid[z] = 0
		} else if i < 0 {
			res.Liquid[z] = float32(Width / YStep)
		} else {
			res.Liquid[z] = float32(Width/YStep-1-i) + (liquidTemp-item[i+1][Length/XStep-1-index])/(item[i][Length/XStep-1-index]-item[i+1][Length/XStep-1-index])
		}
		if res.Liquid[z] >= float32(Width/YStep) && !liquidJoinSet {
			res.Liquid[z] = float32(Width / YStep)
			res.LiquidJoin.IsJoin = true
			res.LiquidJoin.JoinIndex = z
			liquidJoinSet = true
		}
	}, 0, c.Field.Size())
	return res
}
