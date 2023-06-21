package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jpicht/go-netns/netns"
	"github.com/jpicht/go-netns/netnsdocker"
)

func main() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"PID", "Interface", "MAC", "Up", "running"})

	for _, pid_ := range os.Args[1:] {
		var ns *netns.NetNS
		if len(pid_) == 12 {
			var err error
			ns, err = netnsdocker.Open(netnsdocker.OpenOpts{ID: pid_})
			if err != nil {
				log.Printf("invalid container id %v: %v", pid_, err)
				continue
			}
		} else {
			pid, err := strconv.Atoi(pid_)
			if err != nil {
				log.Printf("invalid pid %v: %v", pid_, err)
				continue
			}
			ns, err = netns.Open(pid)
			if err != nil {
				log.Printf("cannot open netns for %d: %v", pid, err)
				continue
			}
		}

		ifs, err := ns.Interfaces()
		if err != nil {
			log.Printf("cannot get interfaces for %s: %v", pid_, err)
			continue
		}

		for _, IF := range ifs {
			t.AppendRow(table.Row{
				pid_, IF.Name, IF.HardwareAddr,
				IF.Flags&net.FlagUp > 0,
				IF.Flags&net.FlagRunning > 0,
			})
		}

		ns.Close()
	}

	t.Render()
}
