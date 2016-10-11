package main

import (
	"bytes"
    "encoding/json"
	"io/ioutil"
    "log"
	"net/http"
    "os/exec"
	"strings"
    "time"
)

// Get the routing weight for service+tags that match this service instance's tag
// Using the Amalgam8 controller API
func (p *Plugin) routingPercentage(service serviceInstance) (map[string]metric, float64, error) {
	id, _ := p.metricIDAndName()
	value := 0.0

	cmdArgs := []string{"-H 'Accept: application/json'","http://localhost:31200/v1/rules/routes/" + service.Name}
	nu, errrr := exec.Command("curl", cmdArgs...).Output()
	if errrr != nil {
		log.Fatal(errrr)
	}
	a := string(nu)
	var rules RulesList
	json.Unmarshal([]byte(a), &rules)

	//log.Println(service)

	for _, rule := range rules.Rules {
		if rule.Destination != service.Name {
			continue
		}
		if len(rule.Route.Backends) == 1 {
			//log.Println("only 1 backend")
			value = 1.0 //100.0
		} else {
			for _, backend := range rule.Route.Backends {
				weight := backend.Weight
				for _, s := range backend.Tags {
					for _, tag := range service.Tags {
						if s == tag {
							value = weight
							if weight == 0 {
								//log.Println("found tag, but weight is 0 or not present")
								return nil, 0.0, nil
							}
							metrics := map[string]metric{
								id: {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: value,
										},
									},
									Min: 0,
									Max: 1,
								},
							}
							return metrics, value, nil
						}
					}
				}
			}
		}
	}
	return nil, 0.0, nil	
}



/********************************
*
* Amalgam8 Controller /v1/rules/routes/{service_name}
*
********************************/

type Backend struct {
	Weight float64 `json:"weight,omitempty"`
	Tags []string `json:"tags"`
}

type Route struct {
	Backends []Backend `json:"backends"`
}

type Rule struct {
	Id string `json:"id"`
	Priority int `json:"priority"`
	Destination string `json:"destination"`
	Route Route `json:"route"`
}

type RulesList struct {
	Rules []Rule `json:"rules"`
}

func getRouteList(serviceName string) RulesList {
	cmdArgs := []string{"-H 'Accept: application/json'","http://localhost:31200/v1/rules/routes/" + serviceName}
	nu, errrr := exec.Command("curl", cmdArgs...).Output()
	if errrr != nil {
		log.Fatal(errrr)
	}
	a := string(nu)
	var rules RulesList
	json.Unmarshal([]byte(a), &rules)
	return rules
}

func clearRoutes(NodeId string) {
	idParts := strings.Split(NodeId, ";")
	name := serviceInstancesByContainerID[idParts[0]].Name
	log.Println("Clearing routes for " + name)
	req, _ := http.NewRequest("DELETE", "http://localhost:31200/v1/rules/routes/" + name, nil)
	_, _ = http.DefaultClient.Do(req)
}

func adjustWeight(NodeId string, changeAmount float64) {
	idParts := strings.Split(NodeId, ";")
	serviceName := serviceInstancesByContainerID[idParts[0]].Name
	serviceName = strings.Replace(serviceName, " ", "", -1)
	routes := getRouteList(serviceName)

	var newBackends []Backend 

	for _, b := range routes.Rules[0].Route.Backends {
		if b.Tags[0] == serviceInstancesByContainerID[idParts[0]].Tags[0] {
			log.Println("Incrementing .",serviceName,".", b.Tags[0])
			log.Println("Old weight: ", b.Weight)
			b.Weight += changeAmount
			log.Println("New weight: ", b.Weight)
		}
		newBackends = append(newBackends, b)
	}
	
	routes.Rules[0].Route.Backends = newBackends
	body, _ := json.Marshal(routes)
	log.Println(string(body))
	req, _ := http.NewRequest("PUT", "http://localhost:31200/v1/rules/routes/" + serviceName, bytes.NewBuffer([]byte(string(body))))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	rbody, _ := ioutil.ReadAll(resp.Body)
    log.Println("response Body:", string(rbody))

}