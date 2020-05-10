package mqsi

import (
	// "fmt"
	// "time"
	"runtime"
	"log"
	"os/exec"
	"strings"
	"encoding/json"
)

//added my structs
type brokerJSON struct {
    Brokers []Broker
}
type Broker struct {
    Name    string
    Status string
	ExecutionGroups []ExecutionGroup
}

type ExecutionGroup struct {
	Name	string
	Status	string
	Components	[]Component
}
type Component struct {
	Name	string
	Status	string
	Type	string
	DeployedTime string
	BarFileName	string
	BarFileLastModifiedTime	string
}
var mqsiscriptloc string

//my fuction
func PullnodeInfoJson(mqsiscriptlocation string) []byte {
	mqsiscriptloc = mqsiscriptlocation
	mqsibrkoutput := ExecMqsi()
	
	lines := Filter(strings.Split(strings.TrimSuffix(mqsibrkoutput, "\n"), "\n"), func(v string) bool {
		return len(strings.Split(v, "'")) == 5 || len(strings.Split(v, "'")) == 3
	})

	brkjsn := brokerJSON {
		Brokers: CreateBrokerArray(lines),
	}

	buf, err := json.MarshalIndent(brkjsn,"","   ")
	if err != nil {
		log.Fatal(err)
		//logp.Error(err)
	}

	return buf
}

func CreateBrokerArray(lines []string) []Broker {
	brkarray := make([]Broker, 0)
	for _, line := range lines {
			
			brokerName := strings.Split(line, "'")[1]
			status := "Not found"
			if strings.Contains(strings.Split(line, "'")[2],"stopped") || strings.Contains(strings.Split(line, "'")[2],"running") {
				status = strings.Split(line, "'")[2][4:11]
			}else if strings.Contains(strings.Split(line, "'")[4],"stopped") || strings.Contains(strings.Split(line, "'")[4],"running"){
				status = strings.Split(line, "'")[4][4:11]
			}else {
				status = strings.Split(line, "'")[2]
			}
			executionGroupArray := make([]ExecutionGroup, 0)

			if (status == "running") {
				executionGroupArray = CreateExecutionGroupArray(brokerName)
			}else if (status == "stopped") {
				executionGroupArray = []ExecutionGroup{
					{
						Name: "NA",
						Status: "NA",
						Components: []Component{
							{
								Name: "NA",
								Status: "NA",
								Type: "NA",
								DeployedTime: "NA", 
								BarFileName: "NA",
								BarFileLastModifiedTime: "NA",
							},						
						},
					},
				}
			}
			brk := Broker{
				Name: brokerName,
				Status: status,
				ExecutionGroups: executionGroupArray,
			}
			brkarray = append(brkarray, brk)
	}
	return brkarray
}

func CreateExecutionGroupArray(brokerName string) []ExecutionGroup {
	mqsioutput := ExecMqsi(brokerName)
	lines := Filter(strings.Split(strings.TrimSuffix(mqsioutput, "\n"), "\n"), func(v string) bool {
        return len(strings.Split(v, "'")) == 5
    })
	
	executionGroupArray := make([]ExecutionGroup, 0)
	for _, line := range lines {
		executionGroupName := strings.Split(line, "'")[1]
		status := strings.Split(line, "'")[4][4:11]
		applicationArray := make([]Component, 0)

		if (status == "running") {
			applicationArray = CreateComponentArray(brokerName, executionGroupName)
		}

		if (status == "stopped") || ((status == "running") && len(applicationArray) == 0) {
			applicationArray = []Component{
				{
					Name: "NA",
					Status: "NA",
					Type: "NA",
					DeployedTime: "NA", 
					BarFileName: "NA",
					BarFileLastModifiedTime: "NA",
				},			
			}
		}
		executionGroup := ExecutionGroup{
			Name: executionGroupName,
			Status: status,
			Components: applicationArray, 
		}
		executionGroupArray = append(executionGroupArray, executionGroup)
	}
	return executionGroupArray
}

func CreateComponentArray(brokerName string, executionGroupName string) []Component {
	mqsioutput := ExecMqsi(brokerName, "-e", executionGroupName, "-d2")	
	applicationArray := CreateApplicationArray(mqsioutput, "Application")	
	RESTArray := CreateApplicationArray(mqsioutput, "REST")
	componentArray := make([]Component, 0)
	componentArray = append(applicationArray, RESTArray...)	
	return componentArray
}

func CreateApplicationArray(mqsioutput string, cmptype string) []Component {
	applicationLines := Filter(strings.Split(strings.TrimSuffix(mqsioutput, "--------"), "--------"), func(v string) bool {
        return strings.Split(v, " ")[1] == cmptype
    })
	
	applicationArray := make([]Component, 0)
	for _, applicationLine := range applicationLines {
		lines := strings.Split(strings.TrimSuffix(applicationLine, "\n"), "\n")
		applicationName := strings.Split(lines[1], "'")[1]
		status := strings.Split(lines[1], "'")[4][4:11]
		deployedTime := strings.Split(lines[3], "'")[1]
		barFileName	:=strings.Split(lines[3], "'")[3]
		barFileLastModifiedTime	:=strings.Split(lines[4], "'")[1]
		Component := Component{
			Name: applicationName,
			Status: status,
			Type: cmptype,
			DeployedTime: deployedTime, 
			BarFileName: barFileName,
			BarFileLastModifiedTime: barFileLastModifiedTime,
		}
		applicationArray = append(applicationArray, Component)
	}
	return applicationArray
}

func ExecMqsi(mqsiargs ...string) string {
	var mqsifileloc string
	if runtime.GOOS == "windows" {
		mqsifileloc = "mqsicmd.bat"
	}else if runtime.GOOS == "linux" {
		mqsifileloc = mqsiscriptloc//"./mqsicmd.sh"
	}else {
		mqsifileloc = "OSOtherThanWinAndLinux"
	}
	stdoutStderr, err := exec.Command(mqsifileloc, mqsiargs...).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	return string(stdoutStderr)
}
func Filter(vs []string, f func(string) bool) []string {
    vsf := make([]string, 0)
    for _, v := range vs {
        if f(v) {
            vsf = append(vsf, v)
        }
    }
    return vsf
}