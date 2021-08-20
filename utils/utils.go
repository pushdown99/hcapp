package utils

import (
	"log"
	"net"
	"strings"

	"github.com/fatih/color"
	//"github.com/jacobsa/go-serial/serial"
)

/*
type Config struct {
  httpHost string
  wsHost string
  servName string
  deptName string
  uidNum string
  rcnNum string
  macAddr string
  ipAddr string
  baudRate string
  Printer string
  Port1 string
  Port2 string
  Token string
  heartBeat string
  intChTimeout string
  minReadSize string
  posType string
  bnkRepo string
  bnkPrefix string
  osArch string
  osInfo string
  envInfo string
  wsConnected int
  license string
  version string
  build string
}

func GetAttr(obj interface{}, fieldName string) reflect.Value {
    pointToStruct := reflect.ValueOf(obj) // addressable
    curStruct := pointToStruct.Elem()
    if curStruct.Kind() != reflect.Struct {
        panic("not struct")
    }
    curField := curStruct.FieldByName(fieldName) // type: reflect.Value
    if !curField.IsValid() {
        panic("not found: " + fieldName)
    }
    return curField
}
*/

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IPv4(0,0,0,0)
	}
  
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP
}

func GetOutboundMac(currentIP string) string {
	var currentNetworkHardwareName string
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				if strings.Contains(addr.String(), currentIP) {
					currentNetworkHardwareName = interf.Name
				}
			}
		}
	}
	netInterface, _ := net.InterfaceByName(currentNetworkHardwareName)
	return netInterface.HardwareAddr.String()
}

/*
func GetConfig(f string, c *Config) bool {

  fmt.Println("godotenv")
	err := godotenv.Load(f)
	if err != nil {
		fmt.Printf("Error loading C:\\hc\\.env file")
		return false
	}
	c.httpHost = os.Getenv("SERVER")
  fmt.Println("host:", c.httpHost)
	if c.httpHost == "" {
		fmt.Printf("[x] Environment(.env) SERVER read error")
		c.httpHost = "https://smart.hancomlifecare.com"
	}
	c.wsHost = os.Getenv("WS")
	c.servName = os.Getenv("SERVICE")
	c.deptName = os.Getenv("NAME")
	c.uidNum = os.Getenv("UID")
	c.rcnNum = os.Getenv("RCN")
	c.baudRate = os.Getenv("BAUDRATE")
	if c.baudRate == "" {
		c.baudRate = "19200"
	}
	c.Printer = os.Getenv("PRINTER")
	c.Port1 = os.Getenv("PORT1")
	c.Port2 = os.Getenv("PORT2")
	c.heartBeat = os.Getenv("HEARTBEAT")
	c.intChTimeout = os.Getenv("INTERCHTMO")
	if c.intChTimeout == "" {
		c.intChTimeout = "50"
	}
	c.minReadSize = os.Getenv("MINREADSIZE")
	if c.minReadSize == "" {
		c.minReadSize = "0"
	}
	c.posType = os.Getenv("POSTYPE")
	if c.posType == "" {
		c.posType = "COM"
	}
	c.bnkRepo = os.Getenv("BNK_REPO")
	c.bnkPrefix = os.Getenv("BNK_PREFIX")
	c.ipAddr = GetOutboundIP().String()
	c.macAddr = GetOutboundMac(c.ipAddr)

  return true
}
*/

func warnLog(s string) {
	color.Set(color.FgRed)
	log.Printf("%s", s)
	color.Set(color.FgWhite)
}

func normLog(s string) {
  color.Set(color.FgRed)
  log.Printf("%s", s)
  color.Set(color.FgWhite)
}

	// For windows usage, these options (termios) do not conform well to the
	//     windows serial port / comms abstractions.  Please see the code in
	//		 open_windows setCommTimeouts function for full documentation.
	//   	 Summary:
	//			Setting MinimumReadSize > 0 will cause the serialPort to block until
	//			until data is available on the port.
	//			Setting IntercharacterTimeout > 0 and MinimumReadSize == 0 will cause
	//			the port to either wait until IntercharacterTimeout wait time is
	//			exceeded OR there is character data to return from the port.

  /*
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
*/

