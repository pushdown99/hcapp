module hancom

go 1.12

require (
	github.com/fatih/color v1.12.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jacobsa/go-serial v0.0.0-20180131005756-15cf729a72d4 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kardianos/service v1.2.0
	github.com/matishsiao/goInfo v0.0.0-20200404012835-b5f882ee2288
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	hancom.com/icon v0.0.0
	hancom.com/systray v0.0.0
	hancom.com/utils v0.0.0
	hancom.com/websocket v0.0.0
)

replace (
	hancom.com/icon v0.0.0 => ./icon
	hancom.com/systray v0.0.0 => ./systray
	hancom.com/utils v0.0.0 => ./utils
	hancom.com/websocket v0.0.0 => ./websocket
)
