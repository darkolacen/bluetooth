// +build ignore

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"tinygo.org/x/bluetooth"
)

type Service struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	UUID       string `json:"uuid"`
	Source     string `json:"source"`
}

func (s Service) VarName() string {
	str := strings.ReplaceAll(s.Name, "Service", "")
	str = strings.ReplaceAll(str, ":", "")
	str = strings.ReplaceAll(str, "-", "")
	str = strings.Title(str)
	return strings.ReplaceAll(str, " ", "")
}

func (s Service) UUIDFunc() string {
	if len(s.UUID) == 4 {
		return "New16BitUUID(0x" + s.UUID + ")"
	}
	uuid, err := bluetooth.ParseUUID(strings.ToLower(s.UUID))
	if err != nil {
		panic(err)
	}
	b := uuid.Bytes()
	bs := hex.EncodeToString(b[:])
	bss := ""
	for i := 0; i < len(bs); i += 2 {
		bss = "0x" + bs[i:i+2] + "," + bss
	}
	return "NewUUID([16]byte{" + bss + "})"
}

func main() {
	jsonFile, err := os.Open("data/service_uuids.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	data, _ := ioutil.ReadAll(jsonFile)

	var services []Service
	json.Unmarshal(data, &services)

	f, err := os.Create("service_uuids.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	packageTemplate := template.Must(template.New("").Parse(tmpl))

	packageTemplate.Execute(f, struct {
		Timestamp time.Time
		Services  []Service
	}{
		Timestamp: time.Now(),
		Services:  services,
	})
}

var tmpl = `// Code generated by bin/gen-service-uuids; DO NOT EDIT.
// This file was generated on {{.Timestamp}} using the list of standard service UUIDs from
// https://github.com/NordicSemiconductor/bluetooth-numbers-database/blob/master/v1/service_uuids.json
//
package bluetooth

var (
{{ range .Services }}
	// ServiceUUID{{.VarName}} - {{.Name}}
	ServiceUUID{{.VarName}} = {{.UUIDFunc}}
{{ end }}
)
`
