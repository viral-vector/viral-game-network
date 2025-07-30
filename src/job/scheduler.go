package job

import (
	"log"

	"github.com/robfig/cron/v3"
)

type Entry struct {
	Name string
	Time string
	Func func()
}

func Stack() []Entry {
	/**
	 *
	 */
	return []Entry{
		{
			Name: "monitoring",
			Time: "@every 1s",
			Func: Job_Monitoring,
		},
		{
			Name: "lobby_server_provisioner",
			Time: "@every 3s",
			Func: Job_Lobby_Server_Provisioner,
		},
		{
			Name: "lobby_server_stewardship",
			Time: "@every 10s",
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
	for _, v := range Stack() {
		log.Println("Scheduling Job: ", v.Name)
		run.AddFunc(v.Time, v.Func)
	}
	run.Start()
}
