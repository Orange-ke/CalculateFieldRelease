package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"lz/calculator"
	"lz/conf"
	"lz/model"
	"strconv"
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	c    calculator.Calculator
	conn *websocket.Conn
	// request
	msg chan model.Msg
	// response
	selectCaster         chan string
	envSet               chan model.Env
	changeInitialTemp    chan float32
	changeNarrowSurface  chan model.NarrowSurface
	changeWideSurface    chan model.WideSurface
	changeV              chan float32
	started              chan struct{}
	stopped              chan struct{}
	tailStart            chan struct{} // 拉尾坯
	startPushSliceDetail chan int
	stopPushSliceDetail  chan struct{}

	generate      chan struct{}
	generateSlice chan int

	generateVerticalSlice1 chan struct{}
	generateVerticalSlice2 chan model.VerticalReqData

	mu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		msg:                 make(chan model.Msg, 10),
		selectCaster:        make(chan string, 10),
		envSet:              make(chan model.Env, 10),
		changeInitialTemp:   make(chan float32, 10),
		changeNarrowSurface: make(chan model.NarrowSurface, 10),
		changeWideSurface:   make(chan model.WideSurface, 10),
		changeV:             make(chan float32, 10),
		started:             make(chan struct{}, 10),
		stopped:             make(chan struct{}, 10),
		tailStart:           make(chan struct{}, 10),

		generateSlice: make(chan int, 10),

		generateVerticalSlice1: make(chan struct{}, 10),
		generateVerticalSlice2: make(chan model.VerticalReqData, 10),
	}
}

func (h *Hub) handleResponse() {
	defer func() {
		log.Fatal("停止handleResponse")
	}()
	for {
		select {
		case fileName := <-h.selectCaster: // 选择铸机
			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Println("err", err)
				return
			}
			reply := model.Msg{
				Type:    "caster_info",
				Content: string(data),
			}
			h.mu.Lock()
			err = h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case env := <-h.envSet: // 设置计算环境
			if h.c == nil {
				// 初始化铸坯尺寸
				calculator.ZLength = env.Coordinate.ZLength
				calculator.Length = env.Coordinate.Length / 2
				calculator.Width = env.Coordinate.Width / 2
				log.Info("ZLength:", calculator.ZLength, " ,Length:", calculator.Length, " ,Width:", calculator.Width)
				h.c = calculator.NewCalculatorWithArrDeque(nil)
			}
			h.c.GetCastingMachine().SetFromJson(env.Coordinate) // 初始化铸机尺寸
			data, err := ioutil.ReadFile(conf.AppConfig.NozzleConfigFile)
			if err != nil {
				log.Println("err", err)
				return
			}
			h.c.GetCastingMachine().SetCoolerConfig(env, data)     // 设置冷却参数
			h.c.GetCastingMachine().SetV(env.DragSpeed)            // 设置拉速
			h.c.InitSteel(env.SteelValue, h.c.GetCastingMachine()) // 设置钢种物性参数
			h.c.InitPushData(env.Coordinate)                       // 设置推送数据相关参数
			reply := model.Msg{
				Type:    "env_set",
				Content: "env is set",
			}
			h.mu.Lock()
			err = h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case temp := <-h.changeInitialTemp: // 改变初始浇铸温度
			h.c.GetCastingMachine().SetStartTemperature(temp)
			reply := model.Msg{
				Type:    "initial_temp_set",
				Content: "initial_temp_set",
			}
			h.mu.Lock()
			err := h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case narrowSurface := <-h.changeNarrowSurface: // 改变结晶器窄面水量
			h.c.GetCastingMachine().SetNarrowSurfaceIn(narrowSurface.In)
			h.c.GetCastingMachine().SetNarrowSurfaceOut(narrowSurface.Out)
			reply := model.Msg{
				Type:    "narrow_surface_temp_set",
				Content: "narrow_surface_temp_set",
			}
			h.mu.Lock()
			err := h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case wideSurface := <-h.changeWideSurface: // 改变结晶器宽面水量
			h.c.GetCastingMachine().SetWideSurfaceIn(wideSurface.In)
			h.c.GetCastingMachine().SetWideSurfaceOut(wideSurface.Out)
			reply := model.Msg{
				Type:    "wide_surface_temp_set",
				Content: "wide_surface_temp_set",
			}
			h.mu.Lock()
			err := h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case v := <-h.changeV: // 改变拉速
			h.c.GetCastingMachine().SetV(v)
			reply := model.Msg{
				Type:    "v_set",
				Content: "v_set",
			}
			h.mu.Lock()
			err := h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case <-h.started: // 开始计算
			// 从calculator里面的hub中获取是否有
			h.c.GetCalcHub().StartSignal()
			go h.c.Run()    // 不断计算
			go h.pushData() // 获取推送的计算结果到前端
			reply := model.Msg{
				Type:    "started",
				Content: "Started",
			}
			h.mu.Lock()
			err := h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case <-h.stopped: // 停止计算
			h.c.GetCalcHub().StopSignal()
			reply := model.Msg{
				Type:    "stopped",
				Content: "stopped",
			}
			h.mu.Lock()
			err := h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case <-h.tailStart: // 拉尾坯 todo
			h.c.SetStateTail()
			reply := model.Msg{
				Type:    "tail_start",
				Content: "started to tail",
			}
			h.mu.Lock()
			err := h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("回复消息失败")
			}
		case index := <-h.generateSlice: // 横切面信息
			reply := model.Msg{
				Type: "slice_generated",
			}
			sliceData := h.c.GenerateSLiceInfo(index)
			data, err := json.Marshal(sliceData)
			if err != nil {
				log.WithField("err", err).Error("温度场切片推送数据json解析失败")
				return
			}
			reply.Content = string(data)
			h.mu.Lock()
			err = h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("发送温度场切片推送消息失败")
			}
		case <-h.generateVerticalSlice1: // 纵切面曲线
			reply := model.Msg{
				Type: "vertical_slice1_generated",
			}
			verticalSliceData := h.c.GenerateVerticalSlice1Data()
			data, err := json.Marshal(verticalSliceData)
			if err != nil {
				log.WithField("err", err).Error("纵向切片1推送数据json解析失败")
				return
			}
			reply.Content = string(data)
			h.mu.Lock()
			err = h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("发送纵向切片1推送消息失败")
			}
		case reqData := <-h.generateVerticalSlice2: // 纵切面云图
			reply := model.Msg{
				Type: "vertical_slice2_generated",
			}
			verticalSliceData := h.c.GenerateVerticalSlice2Data(reqData)
			data, err := json.Marshal(verticalSliceData)
			if err != nil {
				log.WithField("err", err).Error("纵向切片2推送数据json解析失败")
				return
			}
			reply.Content = string(data)
			h.mu.Lock()
			err = h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("发送纵向切片2推送消息失败")
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (h *Hub) handleRequest() {
	// 在此对请求进行预处理
	defer func() {
		log.Fatal("停止handleRequest")
	}()
	for {
		select {
		case msg := <-h.msg:
			switch msg.Type {
			case "select_caster":
				caster := msg.Content
				fileName := conf.AppConfig.CasterHomePath + caster + ".json"
				h.selectCaster <- fileName
			case "env":
				var env model.Env
				err := json.Unmarshal([]byte(msg.Content), &env)
				if err != nil {
					log.Println("err", err)
					return
				}
				log.WithField("env", env).Info("获取到计算环境参数")
				h.envSet <- env
			case "change_initial_temp":
				temp, err := strconv.ParseFloat(msg.Content, 10)
				if err != nil {
					log.Println("err", err)
					return
				}
				log.WithField("temp", temp).Info("获取到初始温度参数")
				h.changeInitialTemp <- float32(temp)
			case "change_narrow_surface":
				var narrowSurface model.NarrowSurface
				err := json.Unmarshal([]byte(msg.Content), &narrowSurface)
				if err != nil {
					log.Println("err", err)
					return
				}
				log.WithField("narrowSurface", narrowSurface).Info("获取到窄面温度参数")
				h.changeNarrowSurface <- narrowSurface
			case "change_wide_surface":
				var wideSurface model.WideSurface
				err := json.Unmarshal([]byte(msg.Content), &wideSurface)
				if err != nil {
					log.Println("err", err)
					return
				}
				log.WithField("wideSurface", wideSurface).Info("获取到宽面温度参数")
				h.changeWideSurface <- wideSurface
			case "change_v":
				v, err := strconv.ParseFloat(msg.Content, 10)
				if err != nil {
					log.Println("err", err)
					return
				}
				log.WithField("v", v).Info("获取到拉速参数")
				h.changeV <- float32(v)
			case "start":
				log.Info("开始计算三维温度场")
				h.started <- struct{}{}
			case "stop":
				log.Info("停止计算三维温度场")
				h.stopped <- struct{}{}
			case "tail":
				h.tailStart <- struct{}{}
			case "generate_slice":
				log.Info("获取到生成切片数据的信号")
				index, err := strconv.ParseInt(msg.Content, 10, 64)
				log.Info("获取到切片下标：", index)
				if err != nil {
					log.WithField("err", err).Error("切片下标不是整数")
					return
				}
				if index < 0 || int(index) >= h.c.GetFieldSize() {
					log.Warn("切片下标越界")
					break
				}
				h.generateSlice <- int(index)
			case "generate_vertical_slice1":
				log.Info("获取到生成纵向切片温度曲线数据的信号")
				h.generateVerticalSlice1 <- struct{}{}
			case "generate_vertical_slice2":
				log.Info("获取到生成纵向切片温度云图数据的信号")
				reqData := model.VerticalReqData{}
				err := json.Unmarshal([]byte(msg.Content), &reqData)
				log.Info("获取到垂直切片云图请求数据：", reqData)
				if err != nil {
					log.Error("json 解析失败")
					return
				}
				if reqData.Index < 0 || reqData.Index >= calculator.Length/calculator.XStep {
					log.Warn("切片下标越界")
					break
				}
				h.generateVerticalSlice2 <- reqData
			default:
				log.Warn("no such type")
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// 周期性的推送温度场云图数据
func (h *Hub) pushData() {
	reply := model.Msg{
		Type: "data_push",
	}
LOOP:
	for {
		select {
		case <-h.c.GetCalcHub().Stop:
			break LOOP
		case <-h.c.GetCalcHub().PeriodCalcResult:
			temperatureData := h.c.BuildData()
			data, err := json.Marshal(temperatureData)
			if err != nil {
				log.WithField("err", err).Error("温度场推送数据json解析失败")
				return
			}
			reply.Content = string(data)
			h.mu.Lock()
			err = h.conn.WriteJSON(&reply)
			h.mu.Unlock()
			if err != nil {
				log.WithField("err", err).Error("发送温度场推送消息失败")
			}
		}
	}
}
