package main

import (
  "log"
  "os"
  "io"
  "fmt"
  "hancom.com/systray"
  "hancom.com/icon"
  "github.com/kardianos/service"
)

/////////////////////////////////////////////////////////////////////////////

func onReady() {
  systray.SetIcon(icon.Data)
  systray.SetTitle("Awesome App")
  systray.SetTooltip("Pretty awesome超级棒")
  mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

  // Sets the icon of a menu item. Only available on Mac and Windows.
  mQuit.SetIcon(icon.Data)

  go func() {
    <-mQuit.ClickedCh
    fmt.Println("Requesting quit")
    systray.Quit()
    fmt.Println("Finished quitting")
  }()
}

func onExit() {
  // clean up here
  os.Exit(0)
}

/////////////////////////////////////////////////////////////////////////////

type GoWindowsService struct{}

func (goWindowsService *GoWindowsService) Start(windowsService service.Service) error {
  go goWindowsService.run()
  return nil
}

func (goWindowsService *GoWindowsService) run() {
}

func (goWindowsService *GoWindowsService) Stop(windowsService service.Service) error {
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

func initLog (f string) {
  fp, err := os.OpenFile(f, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
  if err != nil {
    log.Panic(err)
  }
  defer fp.Close()
  multiWriter := io.MultiWriter(fp, os.Stdout)
  log.SetOutput(multiWriter)
  log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func initFunc () {
  initLog ("c:\\hc\\hancom.log")
  initService ()
}

func termFunc () {
}

/////////////////////////////////////////////////////////////////////////////

func main() {
  initFunc("c:\\hc\\hancom.log")
}

