package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/leominov/self-monitoring/config"
	"github.com/leominov/self-monitoring/gomonversion"
	"github.com/leominov/self-monitoring/msignal"
)

var (
	msgText, msgType string
	// Gomon instance
	Gomon = Monitor{}
)

const (
	msgMask = "%s%s switch status to %s"
)

// Service structure
type Service struct {
	Name                   string
	CurrentState, NewState bool
	DateWatch, DateUpdate  int32
}

// Monitor structure
type Monitor struct {
	Config             *config.File
	CurrentServiceList []string
	ServiceList        []Service
	ListOn, ListOff    []string
	Counter            int
	// Telegram           *tgbotapi.BotAPI
}

// PrepareServiceList for list
func (monitor *Monitor) PrepareServiceList() {
	serviceList := []Service{}

	for _, name := range monitor.Config.ProcessList {
		serviceList = append(serviceList, Service{
			name,
			true,
			false,
			int32(time.Now().Unix()),
			0,
		})
	}

	monitor.ServiceList = serviceList
}

// GetPrefix for messages
func (monitor *Monitor) GetPrefix() string {
	prefixName := ""
	if monitor.Config.NodeName != "" {
		prefixName = monitor.Config.NodeName + ": "
	}

	return prefixName
}

// UpdateServiceList getting current process list
func (monitor *Monitor) UpdateServiceList() error {
	cmd := exec.Command("ps", "-eo", "comm=|sort|uniq")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	monitor.CurrentServiceList = strings.Split(string(output), "\n")

	return nil
}

// CheckStatusList for monitor
func (monitor *Monitor) CheckStatusList() []Service {
	for ID, service := range monitor.ServiceList {
		monitor.ServiceList[ID].NewState = false
		for _, sname := range monitor.CurrentServiceList {
			if monitor.ServiceList[ID].NewState == true {
				continue
			} else if sname == service.Name {
				monitor.ServiceList[ID].NewState = true
				continue
			}
		}
	}

	return monitor.ServiceList
}

// RunLogger service status
func (monitor *Monitor) RunLogger() error {
	if len(monitor.ListOn) > 0 {
		msgText = strings.Join(append(monitor.ListOn), ", ")
		msgType = "ON"
	}

	if len(monitor.ListOff) > 0 {
		msgText = strings.Join(append(monitor.ListOff), ", ")
		msgType = "OFF"
	}

	if msgText != "" {
		logrus.Infof(msgMask, monitor.GetPrefix(), msgText, msgType)
	}

	return nil
}

// RunTelegram service status
func (monitor *Monitor) RunTelegram() error {
	telegram := &monitor.Config.Telegram

	if telegram.Token == "" || telegram.ContactID == 0 {
		logrus.Debug("Check Telegram config parameters: token, contactID")
		return fmt.Errorf("Error Telegram configuration")
	}

	bot, err := tgbotapi.NewBotAPI(telegram.Token)

	if err != nil {
		return err
	}

	if logrus.GetLevel() == logrus.DebugLevel {
		bot.Debug = true
	}

	if len(monitor.ListOn) > 0 {
		msgText = strings.Join(append(monitor.ListOn), ", ")
		msgType = "ON"
	}

	if len(monitor.ListOff) > 0 {
		msgText = strings.Join(append(monitor.ListOff), ", ")
		msgType = "OFF"
	}

	if msgText != "" {
		msg := tgbotapi.NewMessage(telegram.ContactID, fmt.Sprintf(msgMask, monitor.GetPrefix(), msgText, msgType))

		if _, err := bot.SendMessage(msg); err != nil {
			return fmt.Errorf("Error sending message: %v", err)
		}
	}

	return nil
}

// Switch service status
func (monitor *Monitor) Switch() {
	for ID, service := range monitor.ServiceList {
		if service.CurrentState != service.NewState {
			if service.NewState {
				monitor.ListOn = append(monitor.ListOn, service.Name)
			} else {
				monitor.ListOff = append(monitor.ListOff, service.Name)
			}

			if monitor.Counter > 1 {
				monitor.ServiceList[ID].DateUpdate = int32(time.Now().Unix())
			}

			monitor.ServiceList[ID].CurrentState = service.NewState
		}
	}
}

// Notify service status
func (monitor *Monitor) Notify() {
	if monitor.Config.NotifyAtStart == false && monitor.Counter == 1 {
		return
	}

	monitor.RunLogger()

	if monitor.Config.Telegram.Enable {
		if err := monitor.RunTelegram(); err != nil {
			logrus.Error(err)
		}
	}
}

// EmptyTemp data
func (monitor *Monitor) EmptyTemp() {
	msgText = ""
	monitor.ListOn = []string{}
	monitor.ListOff = []string{}
}

// Configure monitor
func (monitor *Monitor) Configure() {
	config, err := config.Load(config.FileFlag)
	config.ParseLoggerFlags()

	if err != nil {
		logrus.Errorf("Error configuring application: %v", err)
		return
	}

	monitor.Config = config
	monitor.PrepareServiceList()
}

// MonitorRoutine loop
func (monitor *Monitor) MonitorRoutine() {
	for {
		monitor.Counter++
		err := monitor.UpdateServiceList()

		if err != nil {
			logrus.Info(err)
			continue
		}

		monitor.CheckStatusList()

		monitor.Switch()
		monitor.Notify()

		monitor.EmptyTemp()

		time.Sleep(monitor.Config.Interval * time.Millisecond)
	}
}

// SignalRoutine loop
func (monitor *Monitor) SignalRoutine() {
	for {
		s := <-msignal.SignalChan
		switch s {
		case msignal.ReloadSignal:
			logrus.Infoln("Reloading configuration...")
			monitor.Configure()
			logrus.Infoln("Done.")

		case msignal.QuitSignal:
			logrus.Infoln("Received shutdown signal")
			msignal.ExitChan <- 0

		case msignal.InfoSignal:
			logrus.Infoln("Counter:", monitor.Counter)
			logrus.Infoln("Service list:", monitor.ServiceList)

		default:
			logrus.Infoln("Catched unknown signal")
		}
	}
}

// Run monitor
func (monitor *Monitor) Run() {
	catched, err := msignal.CatchSender()

	if err != nil {
		logrus.Error(err)
		os.Exit(0)
	}

	if catched {
		logrus.Info("Sended")
		os.Exit(0)
	}

	logrus.Debug("Debug mode enabled")

	logrus.Debugf("Starting Gomon %s...", gomonversion.Version)
	logrus.Debugf("Rinning with PID: %d", os.Getpid())

	go monitor.MonitorRoutine()
	go monitor.SignalRoutine()

	code := <-msignal.ExitChan
	os.Exit(code)
}
