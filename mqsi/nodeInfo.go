package mqsi

import (		
	"runtime"
	"log"
	"os/exec"
	"strings"
	"github.com/mygitlab/iibbeat/config"
)

var mqsiscriptloc string

//This function pull all node informations
func PullnodeCollectionInfo(mqsiscriptlocation string) config.NodeCollection {
	mqsiscriptloc = mqsiscriptlocation
	mqsibrkoutput := ExecMqsi()
	
	lines := Filter(strings.Split(strings.TrimSuffix(mqsibrkoutput, "\n"), "\n"), func(v string) bool {
		return len(strings.Split(v, "'")) == 5 || len(strings.Split(v, "'")) == 3
	})

	brkjsn := config.NodeCollection {
		Nodes: CreateNodeArray(lines),
	}

	return brkjsn
}

func CreateNodeArray(lines []string) []config.Node {
	brkarray := make([]config.Node, 0)
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
			integrationServerArray := make([]config.IntegrationServer, 0)

			if (status == "running") {
				integrationServerArray = CreateIntegrationServerArray(brokerName)
			}else if (status == "stopped") {
				integrationServerArray = []config.IntegrationServer{
					{
						Name: "NA",
						Status: "NA",
						Components: []config.Component{
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
			brk := config.Node{
				Name: brokerName,
				Status: status,
				IntegrationServers: integrationServerArray,
			}
			brkarray = append(brkarray, brk)
	}
	return brkarray
}

func CreateIntegrationServerArray(brokerName string) []config.IntegrationServer {
	mqsioutput := ExecMqsi(brokerName)
	lines := Filter(strings.Split(strings.TrimSuffix(mqsioutput, "\n"), "\n"), func(v string) bool {
        return len(strings.Split(v, "'")) == 5
    })
	
	integrationServerArray := make([]config.IntegrationServer, 0)
	for _, line := range lines {
		executionGroupName := strings.Split(line, "'")[1]
		status := strings.Split(line, "'")[4][4:11]
		applicationArray := make([]config.Component, 0)

		if (status == "running") {
			applicationArray = CreateComponentArray(brokerName, executionGroupName)
		}

		if (status == "stopped") || ((status == "running") && len(applicationArray) == 0) {
			applicationArray = []config.Component{
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
		ISArray := config.IntegrationServer{
			Name: executionGroupName,
			Status: status,
			Components: applicationArray, 
		}
		integrationServerArray = append(integrationServerArray, ISArray)
	}
	return integrationServerArray
}

func CreateComponentArray(brokerName string, executionGroupName string) []config.Component {
	mqsioutput := ExecMqsi(brokerName, "-e", executionGroupName, "-d2")	
	applicationArray := CreateApplicationArray(mqsioutput, "Application")	
	RESTArray := CreateApplicationArray(mqsioutput, "REST")
	componentArray := make([]config.Component, 0)
	componentArray = append(applicationArray, RESTArray...)	
	return componentArray
}

func CreateApplicationArray(mqsioutput string, cmptype string) []config.Component {
	applicationLines := Filter(strings.Split(strings.TrimSuffix(mqsioutput, "--------"), "--------"), func(v string) bool {
        return strings.Split(v, " ")[1] == cmptype
    })
	
	applicationArray := make([]config.Component, 0)
	for _, applicationLine := range applicationLines {
		lines := strings.Split(strings.TrimSuffix(applicationLine, "\n"), "\n")
		applicationName := strings.Split(lines[1], "'")[1]
		status := strings.Split(lines[1], "'")[4][4:11]
		deployedTime := strings.Split(lines[3], "'")[1]
		barFileName	:=strings.Split(lines[3], "'")[3]
		barFileLastModifiedTime	:=strings.Split(lines[4], "'")[1]
		Component := config.Component{
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
		mqsifileloc = mqsiscriptloc
	}else if runtime.GOOS == "linux" {
		mqsifileloc = mqsiscriptloc
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