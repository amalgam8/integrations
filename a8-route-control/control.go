package main

import (
    "encoding/json"
    "log"
    "net/http"
)

type controlDetails struct {
	id    string
	human string
	icon  string
	dead  bool
}

func (p *Plugin) allControlDetails() []controlDetails {
	return []controlDetails{
		{
			id:    "decrease",
			human: "Decrease Routing by 10%",
			icon:  "fa-angle-down",
			dead:  !p.routesEnabled,
		},
		{
			id:    "increase",
			human: "Increase Routing by 10%",
			icon:  "fa-angle-up",
			dead:  !p.routesEnabled,
		},
		{
			id:    "clearroutes",
			human: "Delete route rules for this service",
			icon:  "fa-trash",
			dead:  !p.routesEnabled,
		},
		{
			id:    "enableroutes",
			human: "Enable routes for this service",
			icon:  "fa-rocket",
			dead:  p.routesEnabled,
		},
	}
}

func (p *Plugin) controlDetails() (string, string, string) {
	for _, details := range p.allControlDetails() {
		if !details.dead {
			return details.id, details.human, details.icon
		}
	}
	return "", "", ""
}

// Control is called by scope when a control is activated. It is part
// of the "controller" interface.
func (p *Plugin) Control(w http.ResponseWriter, r *http.Request) {
	p.lock.Lock()
	defer p.lock.Unlock()
	log.Println("CONTROL")
	log.Println(r.URL.String())
	xreq := request{}
	err := json.NewDecoder(r.Body).Decode(&xreq)
	if err != nil {
		log.Printf("Bad request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Print out the node & control values
	log.Println("node ID: " + xreq.NodeID)
	log.Println("control ID: " + xreq.Control)

	switch {
		case xreq.Control == "clearroutes": clearRoutes(xreq.NodeID)
		case xreq.Control == "increase": adjustWeight(xreq.NodeID, 0.1)
		case xreq.Control == "decrease": adjustWeight(xreq.NodeID, -0.1)
	}

	rpt, err := p.makeReport()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := response{ShortcutReport: rpt}
	raw, err := json.Marshal(res)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
