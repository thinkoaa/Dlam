package sub

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatedier/frp/pkg/util/xlog"
)

var (
	EffectiveList []string
	mu            sync.Mutex
)

func NewStart() {
	banner()
	fmt.Println("\n\nBy:thinkoaa GitHub:https://github.com/thinkoaa/Dlam\n")
	go StartAccept()
	fmt.Println("loading....")
	time.Sleep(2 * time.Second)
	config := LoadConfig()
	args := os.Args[1:]
	if len(args) > 0 && strings.HasSuffix(args[0], "check") {
		xlog.CheckFlag = 0
		var IP_Port_List []Data
		//fofa
		for _, address := range config.InternetAddress {
			IP_Port_List = append(IP_Port_List, Data{IP: address, Port: 7000})
		}
		fmt.Println("已获取配置文件中的whoseInternetAddress")
		fofaData, err := GetDataFromFofa(config.FOFA)
		if err != nil {
			fmt.Println(err)
		}
		IP_Port_List = append(IP_Port_List, ConvertFofaResultsToData(fofaData.Results)...)
		//quake
		quakeData, err := GetDataFromQuake(config.QUAKE)
		if err != nil {
			fmt.Println(err)
		}
		IP_Port_List = append(IP_Port_List, quakeData.Data...)
		//hunter
		hunterData, err := GetDataFromHunter(config.HUNTER)
		if err != nil {
			fmt.Println(err)
		}
		IP_Port_List = append(IP_Port_List, hunterData.RsData.Arr...)
		fmt.Printf("已获取%d条互联网IP\t", len(IP_Port_List))
		IP_Port_List = RemoveDuplicates(IP_Port_List)
		fmt.Printf("去重后:%d条\n", len(IP_Port_List))
		fmt.Println("***开始检测***")
		listServerIP(IP_Port_List, config.MyPortList)
		if len(EffectiveList) != 0 {
			config.InternetAddress = EffectiveList
			WriteConfig(&config)
		}

		fmt.Printf("\n***检测完成***,共发现%d个可用互联网IP,已写入config.toml的whoseInternetAddress字段,从中选取互联网IP,然后配置config.toml中的dnats内容,命令行启动本程序即可\n", len(EffectiveList))
		os.Exit(1)
	} else {

		//获得配置的dnats内容

		//检查是否能映射
		var wg sync.WaitGroup
		//
		fmt.Println("正检测dnats远程ip与端口是否有效，请稍候......")
		xlog.CheckFlag = 1
		for _, dnat := range config.DNATS {
			mappings := []DNATMapping{}
			for _, mapping := range dnat.Mappings {
				mappings = append(mappings, DNATMapping{LocalPort: TmpPort, RemotePort: mapping.RemotePort})
			}
			configToml := buildConfig(Data{
				Port: 7000,
				IP:   dnat.RemoteIP}, mappings)
			time.Sleep(10 * time.Millisecond)
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = runClient(configToml)
				// if err != nil {
				// 	fmt.Printf("******************************** [%s]\n", configToml)
				// }
			}()

		}
		wg.Wait()
		fmt.Println("正式启动前测试完毕,dnats中配置的互联网IP:PORT可用,开始正式映射......")
		xlog.CheckFlag = 2
		for _, dnat := range config.DNATS {
			configToml := buildConfig(Data{
				Port: 7000,
				IP:   dnat.RemoteIP}, dnat.Mappings)
			time.Sleep(10 * time.Millisecond)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := runClient(configToml)
				if err != nil {
					fmt.Printf("********************************映射报错 [%s]\n", configToml)
				}
			}()

		}
		wg.Wait()
	}

}

func listServerIP(IP_Port_List []Data, myPortList []int) error {
	var checkServerIPWG sync.WaitGroup
	semaphore := make(chan struct{}, 50)
	for _, serverIPPort := range IP_Port_List {
		checkServerIPWG.Add(1)
		semaphore <- struct{}{}
		go startFRP(serverIPPort, myPortList, &checkServerIPWG, semaphore)
	}
	checkServerIPWG.Wait()
	return nil
}

func startFRP(serverIPPort Data, myPortList []int, checkServerIPWG *sync.WaitGroup, semaphore chan struct{}) {

	defer checkServerIPWG.Done()
	defer func() {
		<-semaphore
	}()
	mappings := []DNATMapping{}
	for _, myPort := range myPortList {
		mappings = append(mappings, DNATMapping{LocalPort: TmpPort, RemotePort: myPort})
	}
	configToml := buildConfig(serverIPPort, mappings)
	time.Sleep(10 * time.Millisecond)
	err := runClient(configToml)
	if err != nil {
		// fmt.Printf("********************************frpc service error for config file [%s]\n", configToml)
	}

}

func buildConfig(serverIPPort Data, mapping []DNATMapping) string {
	var confStr strings.Builder
	tomlStr := `serverAddr = "%s"
serverPort = %d
		`
	tomlTemp := fmt.Sprintf(tomlStr, serverIPPort.IP, serverIPPort.Port)
	confStr.WriteString(tomlTemp)
	if TmpPort != 0 {
		confStr.WriteString("\nloginFailExit=true")
	}

	// 使用循环进行字符串拼接
	for _, m := range mapping {
		config := fmt.Sprintf(`
[[proxies]]
name = "frp-self-chek-process-%s:%d-%d"
type = "tcp"
localIP = "127.0.0.1"
localPort = %d
remotePort = %d
`, serverIPPort.IP, m.RemotePort, time.Now().UnixMilli(), m.LocalPort, m.RemotePort)
		confStr.WriteString(config)
	}
	return confStr.String()
}

func addEffectiveList(serverIP string) {
	mu.Lock()
	EffectiveList = append(EffectiveList, serverIP)
	mu.Unlock()
}

func banner() {
	banner := `
 ____    __       ______              
/\  _D\ /\ \     /L  _  \  /A\_/M\    
\ \ \/\ \ \ \    \ \ \L\ \/\      \   
 \ \ \ \ \ \ \  __\ \  __ \ \ \__\ \  
  \ \ \_\ \ \ \L\ \\ \ \/\ \ \ \_/\ \ 
   \ \____/\ \____/ \ \_\ \_\ \_\\ \_\
    \/___/  \/___/   \/_/\/_/\/_/ \/_/
`
	print(banner)
}
