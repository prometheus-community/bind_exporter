package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

func main() {
	v := Isc{}

	buf, err := ioutil.ReadFile("example_prod.xml")
	if err != nil {
		fmt.Printf("Problem reading file")
		return
	}

	data := string(buf)

	err = xml.Unmarshal([]byte(data), &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Printf("XMLName: %#v\n", v.XMLName)
	fmt.Printf("Bind: %v\n", v.Bind)
	fmt.Printf("Statistics: %v\n", v.Bind.Statistics)
	fmt.Printf("Views: %v\n", v.Bind.Statistics.Views)
	fmt.Printf("View: %v\n", v.Bind.Statistics.Views)
	fmt.Printf("View[0] name: %v\n", v.Bind.Statistics.Views[0])
	fmt.Printf("View[0] Zones: %v\n", v.Bind.Statistics.Views[0].Zones)
	fmt.Printf("View[0] Zones[0]: %v\n", v.Bind.Statistics.Views[0].Zones[0])
	fmt.Printf("View[0] Zones[0] name: %v\n", v.Bind.Statistics.Views[0].Zones[0].Name)
	fmt.Printf("View[0] Resstat: %v\n", v.Bind.Statistics.Views[0].Resstat)
	fmt.Printf("View[0] Resstat[0]: %v\n", v.Bind.Statistics.Views[0].Resstat[0])

	fmt.Printf("View[0] Requests: %v\n", v.Bind.Statistics.Server.Requests)
	fmt.Printf("View[0] QueriesIn: %v\n", v.Bind.Statistics.Server.QueriesIn)

	for _, stat := range v.Bind.Statistics.Server.Requests.Opcode {
		fmt.Printf("serverNode.Requests.Opcode : %s - %d\n", stat.Name, stat.Counter)
	}
	for _, stat := range v.Bind.Statistics.Server.QueriesIn.Rdtype {
		fmt.Printf("serverNode.Requests.Rdtype : %s - %d\n", stat.Name, stat.Counter)
	}

}
