package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	dis "github.com/Layr-Labs/datalayr/dl-disperser/flags"
	dln "github.com/Layr-Labs/datalayr/dl-node/flags"
	ret "github.com/Layr-Labs/datalayr/dl-retriever/flags"
	rollupCha "github.com/Layr-Labs/datalayr/middleware/rollup-example/challenger/flags"
	rollupSeq "github.com/Layr-Labs/datalayr/middleware/rollup-example/sequencer/flags"

	"github.com/urfave/cli"
)

var myTemplate = `
type {{.Name}} struct{
	{{range $var := .Fields}}
		{{$var.EnvVar}} string 
	{{end}}
}
func (vars {{.Name}}) getEnvMap() map[string]string {
	v := reflect.ValueOf(vars)
	envMap := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		envMap[v.Type().Field(i).Name] = v.Field(i).String()
	}
	return envMap
}
 `

type ServiceConfig struct {
	Name   string
	Fields []Flag
}

type Flag struct {
	Name   string
	EnvVar string
}

func getFlag(flag cli.Flag) Flag {
	strFlag, ok := flag.(cli.StringFlag)
	if ok {
		return Flag{strFlag.Name, strFlag.EnvVar}
	}
	boolFlag, ok := flag.(cli.BoolFlag)
	if ok {
		return Flag{boolFlag.Name, boolFlag.EnvVar}
	}
	intFlag, ok := flag.(cli.IntFlag)
	if ok {
		return Flag{intFlag.Name, intFlag.EnvVar}
	}
	uint64Flag, ok := flag.(cli.Uint64Flag)
	if ok {
		return Flag{uint64Flag.Name, uint64Flag.EnvVar}
	}
	log.Fatalln("Type not found", flag)
	return Flag{}
}

func genVars(name string, flags []cli.Flag) string {

	t, err := template.New("vars").Parse(myTemplate)
	if err != nil {
		panic(err)
	}

	vars := make([]Flag, 0)
	for _, flag := range flags {
		vars = append(vars, getFlag(flag))
	}

	var doc bytes.Buffer
	t.Execute(&doc, ServiceConfig{name, vars})

	return doc.String()

}

func main() {

	configs := `package deploy 

	import "reflect"
	`

	configs += genVars("DisperserVars", dis.Flags)
	configs += genVars("OperatorVars", dln.Flags)
	configs += genVars("RetrieverVars", ret.Flags)
	configs += genVars("RollupSequencerVars", rollupSeq.Flags)
	configs += genVars("RollupChallengerVars", rollupCha.Flags)

	fmt.Println(configs)

	err := os.WriteFile("../service_types.go", []byte(configs), 0644)
	if err != nil {
		log.Panicf("Failed to write file. Err: %s", err)
	}
}
