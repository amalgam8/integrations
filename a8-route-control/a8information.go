package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "os/exec"
    "strings"
    "time"
)

/**********
*
* Structs for handling the Amalgam8 service list API responses
*
***********/

// Value is typically the IP address of the service instance
type serviceEndpoint struct {
	Type string `json:type`
	Value string `json:value`
}

type serviceInstance struct {
	Id string `json:"id"`
	Name string `json:"service_name"`
	Endpoint serviceEndpoint `json:endpoint`
	Tags []string `json:tags`
	ContainerID string `json:"containerid,omitempty"`
	Weight float64
}

type serviceDetails struct {
	Name string `json:"service_name"`
	Instances []serviceInstance `json:instances`
}

type serviceListResponse struct {
	Services []string `json:services`
}



type idAddressPair struct {
	ID string `json:id`
	IP string `json:ip`
}

var latestHostServerResponse string

// hostDockerQuery queries the server running on the host for a list of 
// running Container IDs (docker ps) paired with IP addresses
func hostDockerQuery() {
	log.Println("hostDockerQuery")
	for {
		time.Sleep(2 * time.Second)
		c, err := net.Dial("unix", "/var/run/dockerConnection/hostconnection.sock")
		if err != nil {
			continue;
		}
		// send to socket
		log.Println("sending request to server")
		fmt.Fprintf(c, "hi" + "\n")
		// listen for reply
		message, _ := bufio.NewReader(c).ReadString('\n')
		//log.Println("Message from server: " + message)
		log.Println("Received update from host server")

		// set  this to be the latest response
		latestHostServerResponse = message
	}
} 

// map of service instances, with the IP addresses as the keys
//var serviceInstances []serviceInstance
// map of service instances, with the container ID as the key
var serviceInstancesByContainerID map[string]serviceInstance

func updateAmalgam8ServiceInstances() map[string]serviceInstance{
	// amalgam8
	cmdArgs := []string{"-H 'Accept: application/json'","http://localhost:31300/api/v1/services"}
	o, errrr := exec.Command("curl", cmdArgs...).Output()
	if errrr != nil {
		log.Fatal(errrr)
	}
	s := string(o)
	m := make(map[string]serviceInstance) // IP addresses are the keys to the map of instances
	var svcResponse serviceListResponse
	json.Unmarshal([]byte(s), &svcResponse)
	
	// Get the IP address of each service
	for _, serviceName := range svcResponse.Services {
		var svcDetails serviceDetails
		//log.Println(serviceName)
		cmdArgs = []string{"-H 'Accept: application/json'","http://localhost:31300/api/v1/services/" + serviceName}
		nu, errrr := exec.Command("curl", cmdArgs...).Output()
		if errrr != nil {
			log.Fatal(errrr)
		}
		a := string(nu)
		json.Unmarshal([]byte(a), &svcDetails)

		// Add each instance of the service to the list
		for _, instance := range svcDetails.Instances {
			ip := strings.Split(instance.Endpoint.Value, ":")
			m[ip[0]] = instance
		}
	}
	log.Println("Updated Amalgam8 Service Instances without Container IDs")
	return m
}

func getAllContainerIdAddressPairs(serverJsonString string) []idAddressPair {
	var pairs []idAddressPair
	if len(serverJsonString) == 0 {
		return pairs
	}
	json.Unmarshal([]byte(serverJsonString), &pairs)
	return pairs
}

// Look at the Amalgam8 IP addresses and use this list to filter the list of ID/IP pairs from hostDockerQuery
func getAmalgam8ContainerIds() {
	for {
		log.Println("getAmalgam8ContainerIds")
		time.Sleep(3 * time.Second)

		// Get the IP addresses of Amalgam8 containers
		//addressMap := findAmalgam8Addresses()
		m := make(map[string]serviceInstance) // containerIDs are the keys to this map of instances

		containerIDAddressPairs := getAllContainerIdAddressPairs(latestHostServerResponse)

		// map of service instances by IP address
		serviceInstances := updateAmalgam8ServiceInstances()

		time.Sleep(1 * time.Second)

		// Add the pairs with Amalgam8 IP addresses to our collection
		for _, pa := range containerIDAddressPairs {
			// Check to see if this container is in the collection of Amalgam8 services
			if _, ok := serviceInstances[pa.IP]; ok { 
				var tmp = serviceInstances[pa.IP]
				tmp.ContainerID = pa.ID
				serviceInstances[pa.IP] = tmp
				m[pa.ID] = tmp
				//log.Println(serviceInstances[pa.IP])
			}
		}
		serviceInstancesByContainerID = m
		log.Println("Added Container IDs to Amalgam8 services")
	}
}