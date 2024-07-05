package job

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

type Entry struct {
	Name string
	Time string
	Func func()
}

func stack() []Entry {
	/**
	 *
	 */
	return []Entry{
		{
			Name: "lobby_server_provisioner",
			Time: "@every 3s",
			Func: Job_Lobby_Server_Provisioner,
		},
		{
			Name: "lobby_server_stewardship",
			Time: "@every 8s",
			Func: Job_Lobby_Server_Stewardship,
		},
		{
			Name: "lobby_stewardship",
			Time: "@every 15s",
			Func: Job_Lobby_Stewardship,
		},
	}
}

func Start() {
	run := cron.New(cron.WithSeconds())
	for _, v := range stack() {
		fmt.Println("Scheduler Scheduling: ", v.Name)
		run.AddFunc(v.Time, v.Func)
	}
	run.Start()
}
