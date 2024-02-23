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
            Time: "* * * * * *",
            Func: Job_Lobby_Server_Provisioner,
        },
    }
}

func Start() {
    c := cron.New(cron.WithSeconds())
    for _, v := range stack() {
        c.AddFunc(v.Time, func () { 
            fmt.Println("Scheduler Running: ", v.Name)
            v.Func()
        })
    }
    c.Start()
}