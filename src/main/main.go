package main

import (
	"machine"
	"model"
	"network"
	"time"
)
func main() {
	go func() {
		network.StartNetworkModule()
	}()
	for ; ;  {
		time.Sleep(time.Millisecond)
		mission := network.FetchMission()
		if mission == nil {
			continue
		}
		var m machine.Machine
		baseMachine := machine.BaseMachine{
			Rid:         mission.Rid,
			Pid:         mission.Pid,
			Code:        mission.Code,
			Status:      model.JudgeStatusWaiting,
			TimeLimit:   mission.TimeLimit,
			MemoryLimit: mission.MemoryLimit,
		}
		switch mission.Language {
		case model.LanguageC:
			m = &machine.CMachine{
				BaseMachine: baseMachine,
			}
		case model.LanguageCpp:
			m = &machine.CppMachine{
				BaseMachine: baseMachine,
			}
		case model.LanguageJava:
			m = &machine.JavaMachine{
				BaseMachine: baseMachine,
			}
		case model.LanguageGo:
			m = &machine.GoMachine{
				BaseMachine: baseMachine,
			}
		default:
			// TODO: log language error
		}
		go m.Run(m)
	}
}

//cat /boot/config-`uname -r` | grep '^CONFIG_HZ='
