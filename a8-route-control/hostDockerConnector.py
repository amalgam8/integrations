# Socket server that exists on the host machine to get information from the Docker API for the 
# app in the container

import socket,os
import subprocess
import json
from socket import error as SocketError
import errno
script_dir = os.path.dirname(__file__) 
abs_file_path = os.path.join(script_dir, "dockerConnection/hostconnection.sock")
abs_file_path = "/tmp/dockerConnection/hostconnection.sock"
s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

def getDockerIDsWithIPs():
    cmd = ["docker inspect -f '{{.Id}}-{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -q)"]
    p = subprocess.Popen(cmd, stdout=subprocess.PIPE, shell=True)
    (output, err) = p.communicate()
    #print output
    # create a json array of the containerID, IP pairs
    arr = []
    lines = output.splitlines()
    for i,line in enumerate(lines):
        #filter out the container IDs that don't have an IP address
        pair = line.split('-')
        if len(pair) == 2:
            if pair[1] not in [""]:
                #print pair
                # add pair to the array
                obj = {"id":pair[0], "ip":pair[1]}
                arr.append(obj)
    #print json.dumps(arr)
    return json.dumps(arr)

def getDockerID_IP_PodName():
    podNameLabel = '"io.kubernetes.pod.name"'
    cmd = ["docker inspect -f '{{.Id}}     {{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}     {{index .Config.Labels " + podNameLabel + "}}' $(docker ps -q)"]
    p = subprocess.Popen(cmd, stdout=subprocess.PIPE, shell=True)
    (output, err) = p.communicate()
    #print output
    kubernetes = False
    # create a json array of the containerID, IP, pod label triples
    arr = []
    karray = []
    podDictionary = {"":[]}
    lines = output.splitlines()
    for i,line in enumerate(lines):
        #filter out the container IDs that don't have an IP address
        triplet = line.split("     ")
        if len(triplet) == 3:
            if triplet[2] in ["<no value>"]: # docker without kubernetes
                if triplet[1] not in [""]:
                    obj = {"id":triplet[0], "ip":triplet[1]}
                    arr.append(obj)
            else: # kubernetes
                kubernetes = True
                podLabel = triplet[2]
                if podLabel not in podDictionary:
                    #print 'new key'
                    podDictionary[podLabel] = {"ip":triplet[1], "values":[{"id":triplet[0], "ip":triplet[1]}]}
                else:
                    #print 'existing key'
                    if triplet[1] not in [""]: # set the pod's IP address if this is the container with the IP address
                        podDictionary[podLabel]["ip"] = triplet[1]
                    podDictionary[podLabel]["values"].append({"id":triplet[0], "ip":triplet[1]})
                #print json.dumps(podDictionary[podLabel]) + "\n"  

    # Add the missing IP address to members of the same pods  
    if kubernetes == True:
        for key in podDictionary:
            for i, v in enumerate(podDictionary[key]):
                if v in ["ip"]:
                    #print 'continuing'
                    continue
                #print podDictionary[key]["values"]
                #print "\n\n"
                for i, pair in enumerate(podDictionary[key]["values"]):
                    #print json.dumps(pair)
                    pair["ip"] = podDictionary[key]["ip"]
                    karray.append(pair)
        #print "\n\n\n\n"
        print "kubernetes"
        #print json.dumps(karray)
        #print "\n\n\n"
        return json.dumps(karray)
    else:       
        #print "\n\n\n\n"
        print "docker"
        #print json.dumps(arr)
        #print "\n\n\n"
        return json.dumps(arr)

#getDockerID_IP_PodName()

try:
    os.remove(abs_file_path)
except OSError:
    pass
s.bind(abs_file_path)
s.listen(1)

while 1:
    try:
        conn, addr = s.accept()
        while 1:
            data = conn.recv(1024)
            data = getDockerID_IP_PodName() #getDockerIDsWithIPs()
            data += "\n"
            
            if not data: 
                break
            #print data
            print 'hostDockerConnector: sending reponse'
            conn.send(data)
        conn.close()
    except SocketError as e:
        pass # Handle error here.
