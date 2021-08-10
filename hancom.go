package main

import (
	"bytes"
  "log"
  "os"
  "io"
	"io/ioutil"
  "fmt"
	"sync"
	"time"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"encoding/hex"
	"encoding/json"
	"net/http"

  "hancom.com/systray"
  "hancom.com/icon"
  "hancom.com/utils"
  "hancom.com/serial"
  "hancom.com/websocket"

	"github.com/fatih/color"
  "github.com/kardianos/service"
	"github.com/joho/godotenv"
	"github.com/matishsiao/goInfo"
	"github.com/skratchdot/open-golang/open"
)

/////////////////////////////////////////////////////////////////////////////

func onReady() {
  log.Printf("[-] onReady (systray)")
  systray.SetIcon(icon.Data)
  systray.SetTitle("Hancom Receipt App")
  systray.SetTooltip("Hancom Receipt App")

	go func() {
	  mQRcode := systray.AddMenuItem("QR코드 리더기 연결하기", "")
	  mSetting := systray.AddMenuItem("설정보기", "")
	  systray.AddSeparator()
	  mUpdate := systray.AddMenuItem("최신 업데이트 확인", "")
    mQuit := systray.AddMenuItem("프로그램 종료", "")

		for {
			select {
			case <-mSetting.ClickedCh:
				u := httpHost + "/pos/registered/" + license
				log.Printf("[-] browser: %s", u)
				open.Run(u)

			case <-mQRcode.ClickedCh:
				u := httpHost + "/pos/pairing/" + license
				log.Printf("[-] browser: %s", u)
				open.Run(u)

			case <-mUpdate.ClickedCh:
				u := httpHost + "/pos/update-check/" + license
				log.Printf("[-] browser: %s", u)
				open.Run(u)

      case <-mQuit.ClickedCh:
        log.Printf("[-] Requesting quit")
        systray.Quit()
        log.Printf("[-] Finished quitting")
			}
		}
  }()
}

func onExit() {
  log.Printf("[-] onExit (systray)")
  // clean up here
  os.Exit(0)
}


/////////////////////////////////////////////////////////////////////////////

type GoWindowsService struct{}

func (goWindowsService *GoWindowsService) Start(windowsService service.Service) error {
	color.Set(color.FgHiWhite)
	osArch = runtime.GOOS + "/" + runtime.GOARCH
	log.Println("#")
	log.Printf("# POS Agent Started! (%s)", osArch)
	log.Println("#")
	color.Set(color.FgWhite)

	log.Println("")
	gi := goInfo.GetInfo()
	log.Println("[v]", gi)
	osInfo = string(gi.Kernel + "/" + gi.Core)
	log.Println("")

	if _, err := os.Stat("C:\\Program Files (x86)"); !os.IsNotExist(err) {
		osInfo += " (64bit)"
	} else {
		osInfo += " (32bit)"
	}

	// configuratiion
	log.Println("----------------------------------------------------------------")
	if getConfig() == false {
		color.Set(color.FgRed)
		log.Printf("[x] Please, check your POS configuration.")
		color.Set(color.FgWhite)
		return nil
	}
	log.Println("----------------------------------------------------------------")

  go goWindowsService.run()
  return nil
}

func (goWindowsService *GoWindowsService) run() {
	baudrate, _ := strconv.Atoi(baudRate)
	intchtmo, _ := strconv.Atoi(intChTimeout)
	minreadsize, _ := strconv.Atoi(minReadSize)

  wg := sync.WaitGroup{}

  if strings.ToUpper(posType) == "BNK" {
    wg.Add(5)
    go Extract()

    if wsHost == "" {
      CheckIn()
      HeartBeat()
    } else {
      SignIn()
      WS()
      PosHeartBeat()
      if(osVersion >= 6) {
        SysTray()
      }
    }
    wg.Wait()

  } else if strings.ToUpper(posType) == "COM" {
    in := OpenCOM(Port1, baudrate, intchtmo, minreadsize)
    if in == nil {
      return
    }
    out := OpenCOM(Printer, baudrate, intchtmo, minreadsize)
    if out == nil {
      return
    }

    wg.Add(5)
    go RunCOM(in, out, Port1, Printer)
    go RunCOM(out, in, Printer, Port1)

    if wsHost == "" {
      CheckIn()
      HeartBeat()
    } else {
      SignIn()
      WS()
      PosHeartBeat()
      if(osVersion >= 6) {
        SysTray()
      }
    }

    wg.Wait()
  } else {
    com := OpenCOM(Port1, baudrate, intchtmo, minreadsize)
    if com == nil {
      return
    }
    wg.Add(5)

    go RunLPT(com, Port1, Printer)

    if wsHost == "" {
      HeartBeat()
      CheckIn()
    } else {
      SignIn()
      WS()
      PosHeartBeat()
      if(osVersion >= 6) {
        SysTray()
      }
    }
  }
}

func (goWindowsService *GoWindowsService) Stop(windowsService service.Service) error {
	color.Set(color.FgHiWhite)
	log.Println("#")
	log.Println("# POS Agent Stopped!")
	log.Println("#")
	color.Set(color.FgWhite)

  return nil
}

func initService () {
  serviceConfig := &service.Config{
    Name:        "GoWindowsService",
    DisplayName: "Go Windows service",
    Description: "Go Windows service",
  }

  goWindowsService := &GoWindowsService{}
  windowsService, err := service.New(goWindowsService, serviceConfig)
  if err != nil {
    log.Println(err)
  }

  err = windowsService.Run()
  if err != nil {
    log.Println(err)
  }
}

/////////////////////////////////////////////////////////////////////////////

func initLog (f string) *os.File {
  fp, err := os.OpenFile(f, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
  if err != nil {
    log.Panic(err)
    return nil
  }
  multiWriter := io.MultiWriter(fp, os.Stdout)
  log.SetOutput(multiWriter)
  log.SetFlags(log.LstdFlags | log.Lmicroseconds)

  return fp
}

func termLog (fp *os.File) {
  fp.Close()
}

func initFunc () {
	// service registeration
  initService ()
}

func termFunc () {
}

/////////////////////////////////////////////////////////////////////////////

type JsonPost struct {
	Data      string
	Timestamp int64
}

func PostReceiptUid(buf []byte) {
	b, _ := json.Marshal(JsonPost{Data: hex.EncodeToString(buf), Timestamp: time.Now().Unix()})
	resp, err := http.Post(httpHost+"/receipt/probe/"+uidNum, "application/json", bytes.NewBuffer(b))
	log.Printf("[p] http post %s/receipt/probe/ %s", httpHost, bytes.NewBuffer(b))
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[x] Http post receipt uid error: ", err)
		color.Set(color.FgWhite)
	} else {
		log.Printf("[-] Http response: %d", resp.StatusCode)
	}
}

func PostReceiptLicense(buf []byte) {
	b, _ := json.Marshal(JsonPost{Data: hex.EncodeToString(buf), Timestamp: time.Now().Unix()})
	resp, err := http.Post(httpHost+"/receipt/probe/"+license, "application/json", bytes.NewBuffer(b))
	log.Printf("[p] http post %s/receipt/probe/%s %s", httpHost, license, bytes.NewBuffer(b))
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[x] Http post receipt license error: ", err)
		color.Set(color.FgWhite)
	} else {
		log.Printf("[-] Http response: %d", resp.StatusCode)
	}
}

func PostReceipt(buf []byte) {
	if wsHost == "" {
		PostReceiptUid(buf)
	} else {
		PostReceiptLicense(buf)
	}
}

/////////////////////////////////////////////////////////////////////////////

func Extract() {
	color.Set(color.FgGreen)
	log.Printf("[v] thread Extract files (%s%s_*)", bnkRepo, bnkPrefix)
	color.Set(color.FgWhite)
	if _, err := os.Stat(bnkRepo); os.IsNotExist(err) {
		color.Set(color.FgRed)
		log.Printf("[x] Repositories is not exist (%s)", bnkRepo)
		color.Set(color.FgWhite)
		return
	}

	latest := time.Now()
	for {
		files, err := ioutil.ReadDir(bnkRepo)
		if err != nil {
			fmt.Println(err)
		}
		for _, f := range files {
			idx := strings.Index(f.Name(), bnkPrefix)
			if idx == 0 {
				//log.Printf("[-] File (%s %s)", f.Name(), f.ModTime())
				if f.ModTime().Sub(latest) >= 0 {
					fmt.Println(f.Name(), f.ModTime(), f.ModTime().Sub(latest))
					buf, _ := ioutil.ReadFile(bnkRepo + f.Name())
					PostReceipt(buf)
				}
			}
		}
		latest = time.Now()         // update
		time.Sleep(5 * time.Second) // 1 sec => 5 sec
	}
}

func readWithTimeout(r io.Reader, n int, tmo int) ([]byte, int, error) {
  var nb int = -1
	buf := make([]byte, n)
	done := make(chan error)
	readAndCallBack := func() {
		nbyte, err := io.ReadAtLeast(r, buf, n)
    nb = nbyte
		done <- err
	}

	go readAndCallBack()

	timeout := make(chan bool)
	sleepAndCallBack := func() { time.Sleep(2e9); timeout <- true }
	go sleepAndCallBack()

	select {
	case err := <-done:
		return buf, nb, err
	case <-timeout:
		return nil, 0, errors.New("Timed out.")
	}

	return nil, -1, errors.New("Can't get here.")
}

func OpenCOM(device string, baudrate int, intchtmo int, minreadsize int) io.ReadWriteCloser {
  options := serial.OpenOptions{
    PortName:              device,
    BaudRate:              uint(baudrate),
    DataBits:              8,
    StopBits:              1,
    InterCharacterTimeout: uint(intchtmo),    //msec
    MinimumReadSize:       uint(minreadsize), //4
  }
  port, err := serial.Open(options)
  if err != nil {
    color.Set(color.FgRed)
    log.Printf("[x] %s serial.Open: %v", device, err)
    color.Set(color.FgWhite)
    return nil
  }
  return port
}

func RunCOM(in io.ReadWriteCloser, out io.ReadWriteCloser, port1 string, port2 string) {
  color.Set(color.FgGreen)
  log.Printf("[v] thread RunCOM (%s => %s)", port1, port2)
  color.Set(color.FgWhite)
	//minreadsize, _ := strconv.Atoi(minReadSize)
	//intchtmo, _ := strconv.Atoi(intChTimeout)

  //buf := make([]byte, 32768)

  for {
    buf := make([]byte, 32768)
    n, err := in.Read(buf)

    if err != nil {
      if err != io.EOF {
        color.Set(color.FgRed)
        log.Printf("[x] Reading from serial port: ", err)
        color.Set(color.FgWhite)
      }
    } else {
      if n > 0 {
        buf = buf[:n]
        log.Printf("[-] read %s  %d bytes, %s", port1, n, hex.EncodeToString(buf))
        nb, _ := out.Write(buf)
        log.Printf("[-] write %s %d bytes %s", port2, nb, hex.EncodeToString(buf))
        PostReceipt(buf)
      }
    }
    buf = nil
  }
}

func RunLPT(in io.ReadWriteCloser, port1 string, port2 string) {
  color.Set(color.FgGreen)
  log.Printf("[v] thread RunLPT (%s => %s)", port1, port2)
  color.Set(color.FgWhite)

  //buf := make([]byte, 32768)

  for {
    buf := make([]byte, 32768)
    n, err := in.Read(buf)

    if err != nil {
      if err != io.EOF {
        color.Set(color.FgRed)
        log.Printf("[x] Reading from serial port: %v", err)
        color.Set(color.FgWhite)
      }
    } else {
      if n > 0 {
        buf = buf[:n]
        log.Printf("[-] read %s  %d bytes, %s", port1, n, hex.EncodeToString(buf))
        out, _ := os.OpenFile(port2, os.O_RDWR, 0666)
        nb, _ := out.Write(buf)
        out.Sync()
        out.Close()
        log.Printf("[-] write %s %d bytes %s", port2, nb, hex.EncodeToString(buf))
        PostReceipt(buf)
      }
    }
    buf = nil
  }
}

/////////////////////////////////////////////////////////////////////////////

type License struct {
	Mac string
	Rcn string
	Ver string
}

func SignIn() {
	b, _ := json.Marshal(License{Mac: macAddr, Rcn: rcnNum, Ver: version})
	resp, err := http.Post(httpHost+"/pos/sign-in/", "application/json", bytes.NewBuffer(b))

	log.Printf("[p] http post %s/pos/sign-in/ %s", httpHost, bytes.NewBuffer(b))
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[x] Http sign-in post error: ", err)
		color.Set(color.FgWhite)
	} else {
		s, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[-] Http response: %d", resp.StatusCode)
		var result map[string]interface{}
		json.Unmarshal([]byte(s), &result)
		if result["code"].(float64) != 200 {
			/*
				go func() {
					debug := true
					w := webview.New(debug)
					defer w.Destroy()
					w.SetTitle("Minimal webview example")
					w.SetSize(800, 600, webview.HintNone)
					w.Navigate(httpHost + "/pos/sign-up/" + macAddr)
					w.Run()
				}()
			*/
		}
		license = result["license"].(string)
		log.Printf("[-] license: %s, version: %s", license, version)
	}
}

type JsonCheckIn struct {
	Name string
	Uid  string
	Rcn  string
	Mac  string
	Arch string
	Info string
	Env  string
}

func CheckIn() {
	b, _ := json.Marshal(JsonCheckIn{Name: deptName, Uid: uidNum, Rcn: rcnNum, Mac: macAddr, Arch: osArch, Info: osInfo, Env: envInfo})
	resp, err := http.Post(httpHost+"/check-in", "application/json", bytes.NewBuffer(b))

	log.Printf("[p] http post %s/check-in %s", httpHost, bytes.NewBuffer(b))
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[x] Http check-in post error: ", err)
		color.Set(color.FgWhite)
	} else {
		log.Printf("[-] Http response: %d", resp.StatusCode)
	}
}

func doHeartBeat(t time.Time) {
	b, _ := json.Marshal(JsonCheckIn{Name: deptName, Uid: uidNum, Rcn: rcnNum, Mac: macAddr, Arch: osArch, Info: osInfo, Env: envInfo})
	resp, err := http.Post(httpHost+"/heartbeat/", "application/json", bytes.NewBuffer(b))
	//defer resp.Body.Close()

	log.Printf("[p] http post %s/heartbeat/ %s", httpHost, bytes.NewBuffer(b))
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[x] Http heartbeat post error: ", err)
		color.Set(color.FgWhite)
	} else {

		log.Printf("[-] Http response: %d", resp.StatusCode)
	}

}

func HeartBeat() {
	hbtimer, _ := strconv.ParseInt(heartBeat, 10, 32)
	color.Set(color.FgGreen)
	log.Printf("[t] thread HeartBeat (per %d seconds)", hbtimer)
	color.Set(color.FgWhite)
	go func() {
		ticker := time.NewTicker(time.Duration(hbtimer) * time.Second)
		defer ticker.Stop()

		//doHeartBeat(time.Now())
		for {
			select {
			case t := <-ticker.C:
				doHeartBeat(t)
			}
		}
	}()
}

func doPosHeartBeat(t time.Time) {
	b, _ := json.Marshal(License{Mac: macAddr, Rcn: rcnNum, Ver: version})
	resp, err := http.Post(httpHost+"/pos/heartbeat/", "application/json", bytes.NewBuffer(b))

	log.Printf("[p] http post %s/pos/heartbeat/ %s", httpHost, bytes.NewBuffer(b))
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("[x] Http pos heartbeat post error: ", err)
		color.Set(color.FgWhite)
	} else {

		log.Printf("[-] Http response: %d", resp.StatusCode)
	}

}

func PosHeartBeat() {
	hbtimer, _ := strconv.ParseInt(heartBeat, 10, 32)
	color.Set(color.FgGreen)
	log.Printf("[t] thread PosHeartBeat (per %d seconds)", hbtimer)
	color.Set(color.FgWhite)
	go func() {
		ticker := time.NewTicker(time.Duration(hbtimer) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				doPosHeartBeat(t)
			}
		}
	}()
}

func WS() {
	color.Set(color.FgGreen)
	log.Printf("[t] thread hcWS %s", wsHost)
	color.Set(color.FgWhite)

	c := websocket.Connect(wsHost, license)
	defer c.Close()

	go func() {
		for {
			if websocket.Connected == 0 {
				c = websocket.Connect(wsHost, license)
			  time.Sleep(1)
			}
			_, message, err := c.ReadMessage()
			if err != nil {
				color.Set(color.FgRed)
				log.Printf("[x] websocket connection error: ", err)
				color.Set(color.FgWhite)
				websocket.Connected = 0
				continue
			}
			log.Printf("[w] received message: %s", message)

			var result map[string]interface{}
			json.Unmarshal([]byte(message), &result)
			/*
				if result["Command"].(string) == "Callback" {
					go func() {
						debug := true
						w := webview.New(debug)
						defer w.Destroy()
						w.SetTitle("Hancom: WS")
						w.SetSize(800, 600, webview.HintNone)
						w.Navigate(result["Message"].(string))
						w.Run()
					}()
					//go myBrowser(result["Message"].(string))
				}
			*/
		}
	}()
}

func SysTray() {
	color.Set(color.FgGreen)
	log.Printf("[t] thread Systray")
	color.Set(color.FgWhite)

	go func() {
		systray.Run(onReady, onExit)
	}()
}

/////////////////////////////////////////////////////////////////////////////

var httpHost string
var wsHost string
var servName string
var deptName string
var uidNum string
var rcnNum string
var macAddr string
var ipAddr string
var baudRate string
var Printer string
var Port1 string
var Port2 string
var Token string
var heartBeat string
var intChTimeout string
var minReadSize string
var posType string
var bnkRepo string
var bnkPrefix string
var osArch string
var osInfo string
var osVersion int
var envInfo string
var wsConnected int = 0
var license string = "1234"
var version string = ""
var build string = ""
var platform string = ""

func getConfig() bool {
	err := godotenv.Load("c:\\hc\\.env")
	if err != nil {
		log.Printf("Error loading C:\\hc\\.env file")
		return false
	}
	//httpHost = "https://smart.hancomlifecare.com"
	httpHost = os.Getenv("SERVER")
	if httpHost == "" {
		color.Set(color.FgRed)
		log.Printf("[x] Environment(.env) SERVER read error")
		color.Set(color.FgWhite)
		httpHost = "https://smart.hancomlifecare.com"
	}
	wsHost = os.Getenv("WS")
	servName = os.Getenv("SERVICE")
	deptName = os.Getenv("NAME")
	uidNum = os.Getenv("UID")
	rcnNum = os.Getenv("RCN")
	baudRate = os.Getenv("BAUDRATE")
	if baudRate == "" {
		baudRate = "19200"
	}
	Printer = os.Getenv("PRINTER")
	Port1 = os.Getenv("PORT1")
	Port2 = os.Getenv("PORT2")
	heartBeat = os.Getenv("HEARTBEAT")
	intChTimeout = os.Getenv("INTERCHTMO")
	if intChTimeout == "" {
		intChTimeout = "50"
	}
	minReadSize = os.Getenv("MINREADSIZE")
	if minReadSize == "" {
		minReadSize = "0"
	}
	posType = os.Getenv("POSTYPE")
	if posType == "" {
		posType = "COM"
	}
	bnkRepo = os.Getenv("BNK_REPO")
	bnkPrefix = os.Getenv("BNK_PREFIX")
	ipAddr = utils.GetOutboundIP().String()
	macAddr = utils.GetOutboundMac(ipAddr)

	log.Printf("[*] Host                     : %s", httpHost)
	if wsHost != "" {
		log.Printf("[*] WS                       : %s", wsHost)
	}
	if servName != "" {
		log.Printf("[*] Service                  : %s", servName)
	}
	if deptName != "" {
		log.Printf("[*] Name                     : %s", deptName)
	}
	if uidNum != "" {
		log.Printf("[*] Uid                      : %s", uidNum)
	}
	log.Printf("[*] Rcn                      : %s", rcnNum)
	log.Printf("[*] BaudRate                 : %s", baudRate)
	log.Printf("[*] Printer                  : %s", Printer)
	log.Printf("[*] Port1                    : %s", Port1)
	log.Printf("[*] Port2 (POS)              : %s", Port2)
	log.Printf("[*] HeartBeat                : %s", heartBeat)
	log.Printf("[*] Inter Character Timeout  : %s", intChTimeout)
	log.Printf("[*] Minimum Read Size        : %s", minReadSize)
	log.Printf("[*] POS Type                 : %s", posType)
	if strings.ToUpper(posType) == "BNK" {
		log.Printf("[*] BNK Repo                 : %s", bnkRepo)
		log.Printf("[*] BNK Prefix               : %s", bnkPrefix)
	}
	log.Printf("[*] IP address               : %s", ipAddr)
	log.Printf("[*] Mac address              : %s", macAddr)
	if strings.ToUpper(posType) != "BNK" {
		log.Println("----------------------------------------------------------------")
		log.Printf("[v] POS: vPort %s (thuru %s) <=> Printer (%s)", Port2, Port1, Printer)
	}
	envInfo = string(httpHost + "/" + deptName + "/" + uidNum + "/" + rcnNum + "/" + baudRate + "/" + Printer + "/" + Port1 + "/" + Port2 + "/" + heartBeat + "/" + intChTimeout + "/" + minReadSize + "/" + posType + "/" + bnkRepo + "/" + bnkPrefix)

	return true
}

/////////////////////////////////////////////////////////////////////////////

func main() {
  initLog(string("c:\\hc\\hancom.log"))
  //fp := initLog(string("c:\\hc\\hancom.log"))
  //defer fp.Close()

  initFunc()
}

