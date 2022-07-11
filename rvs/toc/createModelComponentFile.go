package toc

import (
	l "altair/rvs/globlog"
	"altair/rvs/utils"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var model Models

type Models struct {
	Model Model `json:"Model"`
}
type Model struct {
	Statistics []string `json:"Statistics"`
	Pools      Pool     `json:"Pools"`
	Parts      []Part   `json:"Parts"`
}

type Part struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Pool struct {
	AssemblyPool []AssemblyPool `json:"AssemblyPool"`
	ElementPool  []ElementPool  `json:"ElementPool"`
	NodePool     []NodePool     `json:"NodePool"`
	PartPool     []PartPool     `json:"PartPool"`
	SystemPool   []SystemPool   `json:"SystemPool"`
}

type ElementPool struct {
	Name string `json:"name"`
}

type AssemblyPool struct {
	Name     string     `json:"name"`
	Assembly []Assembly `json:"Assembly"`
}

type Assembly struct {
	Name string `json:"name"`
}

type NodePool struct {
	Name string `json:"name"`
}

type PartPool struct {
	Name string `json:"name"`
}
type SystemPool struct {
	Name string `json:"name"`
}

func readModelfile(modelfilepath string, modelComponentsFilePath string) {

	jsonFile, err := os.Open(modelfilepath)
	// if we os.Open returns an error then handle it
	if err != nil {
		l.Log().Error(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &model)
	updateModelComponentsFile(modelComponentsFilePath)
}

func updateModelComponentsFile(modelComponentsFilePath string) {

	type Statistics struct {
		NoOfNodes      int `json:"noOfNodes"`
		NoOfElements   int `json:"noOfElements"`
		NoOfParts      int `json:"noOfParts"`
		NoOfSystems    int `json:"noOfSystems"`
		NoOfAssemblies int `json:"noOfAssemblies"`
	}

	type Pool struct {
		Name string `json:"name"`
	}

	type Pools struct {
		Type string `json:"type"`
		Pool []Pool `json:"Pool"`
	}

	type ModelComponent struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		PoolName string `json:"poolName"`
	}

	type ModelComponents struct {
		IsAllSelected  bool             `json:"isAllSelected"`
		ModelComponent []ModelComponent `json:"ModelComponent"`
	}
	type Model struct {
		Statistics      Statistics      `json:"Statistics"`
		Pools           []Pools         `json:"Pools"`
		ModelComponents ModelComponents `json:"ModelComponents"`
	}

	type Models struct {
		Model Model `json:"Model"`
	}

	var v Models
	var elePoolData Pools
	var nodePoolData Pools
	var assemPoolData Pools
	var partsPoolData Pools
	var sysPoolData Pools

	for i := 0; i < len(model.Model.Statistics); i++ {
		arrStatVal := strings.Fields(model.Model.Statistics[i])
		firstWord, secondWord := arrStatVal[0], arrStatVal[1]
		if strings.Contains(firstWord, utils.NODES) {
			nodes, _ := strconv.Atoi(secondWord)
			v.Model.Statistics.NoOfNodes = nodes
		} else if strings.Contains(firstWord, utils.ELEMENTS) {
			elements, _ := strconv.Atoi(secondWord)
			v.Model.Statistics.NoOfElements = elements
		} else if strings.Contains(firstWord, utils.SYSTEMS) {
			systems, _ := strconv.Atoi(secondWord)
			v.Model.Statistics.NoOfSystems = systems
		} else if strings.Contains(firstWord, utils.PARTS) {
			parts, _ := strconv.Atoi(secondWord)
			v.Model.Statistics.NoOfParts = parts
		}
	}

	for i := 0; i < len(model.Model.Pools.ElementPool); i++ {
		elePoolData.Pool = append(elePoolData.Pool, Pool{Name: model.Model.Pools.ElementPool[i].Name})
	}

	if len(model.Model.Pools.ElementPool) != 0 {
		elePoolData.Type = "ELEMENTS"
		v.Model.Pools = append(v.Model.Pools, elePoolData)
	}

	for i := 0; i < len(model.Model.Pools.NodePool); i++ {
		nodePoolData.Pool = append(nodePoolData.Pool, Pool{Name: model.Model.Pools.NodePool[i].Name})
	}

	if len(model.Model.Pools.NodePool) != 0 {
		nodePoolData.Type = "NODES"
		v.Model.Pools = append(v.Model.Pools, nodePoolData)
	}

	for i := 0; i < len(model.Model.Pools.AssemblyPool); i++ {
		assemPoolData.Pool = append(assemPoolData.Pool, Pool{Name: model.Model.Pools.AssemblyPool[i].Name})
	}

	if len(model.Model.Pools.AssemblyPool) != 0 {
		assemPoolData.Type = "ASSEMBLIES"
		v.Model.Pools = append(v.Model.Pools, assemPoolData)
	}

	for i := 0; i < len(model.Model.Pools.PartPool); i++ {
		partsPoolData.Pool = append(partsPoolData.Pool, Pool{Name: model.Model.Pools.PartPool[i].Name})
	}

	if len(model.Model.Pools.PartPool) != 0 {
		partsPoolData.Type = "PARTS"
		v.Model.Pools = append(v.Model.Pools, partsPoolData)
	}

	for i := 0; i < len(model.Model.Pools.SystemPool); i++ {
		sysPoolData.Pool = append(sysPoolData.Pool, Pool{Name: model.Model.Pools.SystemPool[i].Name})
	}

	if len(model.Model.Pools.SystemPool) != 0 {
		sysPoolData.Type = "SYSTEMS"
		v.Model.Pools = append(v.Model.Pools, sysPoolData)
	}

	v.Model.ModelComponents.IsAllSelected = false

	for i := 0; i < len(model.Model.Parts); i++ {
		var name string
		arrPartVal := strings.Fields(model.Model.Parts[i].Value)
		for k := 1; k < len(arrPartVal); k++ {
			name = name + " " + arrPartVal[k]
		}
		v.Model.ModelComponents.ModelComponent = append(v.Model.ModelComponents.ModelComponent,
			ModelComponent{Id: arrPartVal[0], Name: name, Type: "PART", PoolName: model.Model.Parts[i].Type})
	}

	if xmlstring, err := json.MarshalIndent(v, "", "    "); err == nil {
		_ = ioutil.WriteFile(modelComponentsFilePath, xmlstring, 0644)
	}

}
