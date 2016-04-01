package main

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	body, err := ioutil.ReadFile("testdata/bind-9.10-sample.xml")
	if err != nil {
		t.Fatal(err)
	}

	stats, err := unmarshal(body)
	if err != nil {
		log.Fatal("Failed to unmarshal XML response: ", err)
	}
	if stats.Server.BootTime.String() != "2016-02-24 13:11:40 +0000 UTC" {
		t.Fatalf("failed Server.BootTime, got %s, expected %s", stats.Server.BootTime, "2016-02-24 13:11:40 +0000 UTC")
	}
	/* TODO: Work this into an actual test */
	// for _, cnt := range stats.Server.Counters {
	// if cnt.Type == "opcode" {
	// log.Printf("%+v\n", cnt)
	// }
	// if cnt.Type == "qtype" {
	// log.Printf("%+v\n", cnt)
	// for _, c := range cnt.Counter {
	// log.Printf("%+v\n", c.Name)
	// log.Printf("%+v\n", c.Counter)
	// }
	// }
	// }
	// Extract QrySuccess and QrySERVFAIL
	// export the zone with viewName / zoneName
	for _, v := range stats.Views {
		for _, z := range v.Zones {
			for _, c := range z.Counters.Counter {
				switch c.Name {
				case "QrySuccess":
					log.Printf("%s/%s %d\n", v.Name, z.Name, c.Counter)
				case "QrySERVFAIL":
					log.Printf("%s/%s %d\n", v.Name, z.Name, c.Counter)
				}
			}
		}
	}
}
