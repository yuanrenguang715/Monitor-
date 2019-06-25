package main

import (
	"errors"
	"flag"
	"fmt"
	"monitor/utils"
	"monitor/utils/helpers"
	"monitor/utils/logger"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// ProtocolIPv4 ipv4
	ProtocolIPv4 Protocol = iota
	// ProtocolIPv6 增
	ProtocolIPv6
)

// PortLog 端口数据日志
type PortLog struct {
	Port       string   `json:"port"`
	InputFlow  int64    `json:"input_flow"`
	OutputFlow int64    `json:"output_flow"`
	AccessIp   []string `json:"access_ip"`
	IpLen      int      `json:"ip_len"`
}

// Protocol so byte
type Protocol byte

var (
	//LocalIp 本机外网ip
	LocalIp string
	//LocalPort 监听本地端口
	LocalPort []string
	//IntervalTime 搜集间隔时间
	IntervalTime int64
	//DbUrl DB写入路径
	DbUrl string
	//Iptablescoms 获取流量命令
	Iptablescoms = []string{"", "-L", "-v", "-n", "-x"}
	//DBName 数据库名称
	DBName string
	//Diptablescoms 获取规则命令
	Diptablescoms = []string{"", "-n", "-L", "INPUT"}
	//IptablesAddRule 添加监控流量规则 append（port）
	IptablesAddRule = []string{"", "-I", "INPUT", "-p", "tcp", "--dport"}
	//UfwEnable 打开防火墙
	UfwEnable = []string{"sudo", "ufw", "enable"}
	//SsPortConn 获取端口连接数命令 append（port）
	SsPortConn = []string{"ss", "state", "all", "sport", "eq"}
)

func main() {
	logger.Init()
	var portstr string
	var err error
	//获取命令参数
	flag.StringVar(&portstr, "p", "", "Port Column: 8080,8081 Not NULL")
	flag.Int64Var(&IntervalTime, "s", 5, "Time default 5s")
	flag.StringVar(&DbUrl, "db-url", "", "db-url Column: http://*.*.*.*:8086 Not NULL")
	flag.StringVar(&DBName, "db", "", "DB Column: test Not NULL")
	flag.Parse()
	LocalPort = strings.Split(portstr, ",")
	if DbUrl == "" || DBName == "" || len(LocalPort) < 1 {
		CheckError(errors.New("parameter Missing"))
		return
	}
	if err := httpGet(); err != nil {
		CheckError(err)
		return
	}
	//获取本机外网ip
	LocalIp, err = helpers.GetExternal()
	if err != nil {
		CheckError(err)
	}
	//打开防火墙
	utils.ExecCommand(UfwEnable)
	//获取防火墙命令 默认ProtocolIPv4
	path, err := exec.LookPath(getIptablesCommand(ProtocolIPv4))
	if err != nil {
		CheckError(err)
		return
	}
	Diptablescoms[0], IptablesAddRule[0], Iptablescoms[0] = path, path, path
	//获取iptable规则列表
	str, err := utils.ExecCommand(Diptablescoms)
	CheckError(err)
	for _, v := range LocalPort {
		rules := regexp.MustCompile(`\w+:[\d]+`).FindAllString(str, -1)
		var is bool
		for _, rv := range rules {
			if regexp.MustCompile(`[\d]+`).FindString(rv) == v {
				is = true
			}
		}
		if !is {
			//添加规则
			_, err := utils.ExecCommand(append(IptablesAddRule, v))
			CheckError(err)
		}
	}
	plchan := make(chan []*PortLog, 1)
	go Consumer(plchan)
	//监控函数
	Producer(plchan)
}

// Producer 生产日志
func Producer(plchan chan []*PortLog) {
	var curpl []PortLog
	//获取初始值
	str, err := utils.ExecCommand(Iptablescoms)
	CheckError(err)
	for _, v := range LocalPort {
		var crep PortLog
		input, out := ChainFlow(str, v)
		estr, err := utils.ExecCommand(append(SsPortConn, v))
		CheckError(err)
		crep.Port = v
		crep.InputFlow = input
		crep.OutputFlow = out
		crep.AccessIp = helpers.ComparisonSlieString(PerrAddress(estr), nil)
		curpl = append(curpl, crep)
	}
	//监控日常
	for {
		diffpl := make([]*PortLog, 0)
		str, err := utils.ExecCommand(Iptablescoms)
		CheckError(err)
		for i, v := range curpl {
			input, out := ChainFlow(str, v.Port)
			estr, err := utils.ExecCommand(append(SsPortConn, v.Port))
			CheckError(err)
			strips := helpers.ComparisonSlieString(PerrAddress(estr), v.AccessIp)
			diffpl = append(diffpl, &PortLog{
				Port:       v.Port,
				InputFlow:  input - v.InputFlow,
				OutputFlow: out - v.OutputFlow,
				AccessIp:   strips,
				IpLen:      len(strips),
			})
			curpl[i].InputFlow = input
			curpl[i].OutputFlow = out
			curpl[i].AccessIp = strips
		}
		plchan <- diffpl
		time.Sleep(time.Duration(IntervalTime) * time.Second)
	}
}

// Consumer 写库  LocalIp
func Consumer(plchan chan []*PortLog) {
	for {
		select {
		case pls := <-plchan:
			for _, v := range pls {
				var musqls []string
				for i := 0; i < len(v.AccessIp); i++ {
					connsqls := "d_connection_log,monitor_addres=" + LocalIp + ":" + v.Port + ",access_addres=" + v.AccessIp[i] + " count=" + strconv.Itoa(v.IpLen)
					musqls = append(musqls, connsqls)
				}
				flowsql := "d_flow_log,monitor_addres=" + LocalIp + ":" + v.Port + " input=" + strconv.FormatInt(v.InputFlow, 10) + ",output=" + strconv.FormatInt(v.OutputFlow, 10)
				musqls = append(musqls, flowsql)
				go httpPost(strings.Join(musqls, "\n"))
			}
		}
	}
}

// CheckError 检查err 写入日志
func CheckError(err error) {
	if err != nil {
		logger.Error(err)
	}
}

// getIptablesCommand 判断协议
func getIptablesCommand(proto Protocol) string {
	if proto == ProtocolIPv6 {
		return "ip6tables"
	} else {
		return "iptables"
	}
}

// ChainFlow 解析进入 出入流量
func ChainFlow(str string, port string) (input int64, output int64) {
	text := strings.Split(str, "\n\n")
	var inputs, outputs, inppags, outpags string
	for i := range text {
		if strings.Contains(text[i], "Chain INPUT") {
			inputs = text[i]
		} else if strings.Contains(text[i], "Chain OUTPUT") {
			outputs = text[i]
		}
	}
	inputby := strings.Split(inputs, "\n")
	for j := range inputby {
		ports := regexp.MustCompile(`[\d]+`).FindString(regexp.MustCompile(`\w+:[\d]+`).FindString(inputby[j]))
		if ports == port {
			inppags = inputby[j]
			break
		}
	}
	outbys := strings.Split(outputs, "\n")
	for j := range outbys {
		ports := regexp.MustCompile(`[\d]+`).FindString(regexp.MustCompile(`\w+:[\d]+`).FindString(outbys[j]))
		if ports == port {
			outpags = outbys[j]
			break
		}
	}
	if inppags != "" {
		input, _ = strconv.ParseInt(strings.Fields(inppags)[1], 10, 64)
	}
	if outpags != "" {
		output, _ = strconv.ParseInt(strings.Fields(outpags)[1], 10, 64)
	}
	return
}

// PerrAddress 解析访问ip
func PerrAddress(str string) []string {
	var addres []string
	text := strings.Split(str, "\n")
	for i := range text {
		if i != 0 && text[i] != "" {
			txs := strings.Fields(text[i])
			addres = append(addres, txs[len(txs)-1])
		}
	}
	return addres
}

//httpPost 发送请求插入influxDB数据库
func httpPost(params string) {
	payload := strings.NewReader(params)
	req, err := http.NewRequest("POST", DbUrl+"/write?db="+DBName, payload)
	if err != nil {
		CheckError(errors.New(fmt.Sprintf("influxDB Post error:", err, ",current time:", time.Now().Unix())))
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		CheckError(errors.New(fmt.Sprintf("influxDB Post error:", err, ",current time:", time.Now().Unix())))
	}
	defer res.Body.Close()
	if res.StatusCode != 204 {
		CheckError(errors.New(fmt.Sprintf("StatusCode Not 204, Is :", res.StatusCode, ",current time :", time.Now().Unix(), ",body:", params)))
	}
}

// httpGet 验证influxdb
func httpGet() error {
	req, err := http.NewRequest("HEAD", DbUrl+"/ping", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 204 {
		return errors.New(fmt.Sprintf("StatusCode error,Or does not exist! StatusCode:", res.StatusCode))
	}
	return nil
}
