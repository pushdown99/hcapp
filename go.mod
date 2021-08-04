module hancom

go 1.12

require (
	github.com/kardianos/service v1.2.0
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	hancom.com/icon v0.0.0
	hancom.com/log v0.0.0
	hancom.com/systray v0.0.0
)

replace (
	hancom.com/icon v0.0.0 => ./icon
	hancom.com/log v0.0.0 => ./log
	hancom.com/systray v0.0.0 => ./systray
)
