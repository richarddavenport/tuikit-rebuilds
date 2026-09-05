// Package fake is a Kubernetes cluster that does not exist.
package fake

import "time"

// At is the moment the cluster is frozen at.
var At = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)

// Pod is one row of the table k9s is mostly made of.
type Pod struct {
	Namespace string
	Name      string
	Ready     string
	Status    string
	Restarts  int
	CPU       int // millicores
	Mem       int // MiB
	Age       time.Duration
	Node      string
}

// Key is the pod's identity, and what a mark is kept by.
//
// A namespace and a name, because a pod name alone is not unique across a
// cluster — and comp.Marks keys by whatever the tool calls identity.
func (p Pod) Key() string { return p.Namespace + "/" + p.Name }

// Pods is the cluster.
func Pods() []Pod {
	return []Pod{
		{Namespace: "acme", Name: "web-7d9f4c8b5-2xk9p", Ready: "1/1", Status: "Running", CPU: 142, Mem: 310, Age: 12 * time.Hour, Node: "pool-1-a"},
		{Namespace: "acme", Name: "web-7d9f4c8b5-p4m2q", Ready: "1/1", Status: "Running", Restarts: 3, CPU: 388, Mem: 512, Age: 12 * time.Hour, Node: "pool-1-b"},
		{Namespace: "acme", Name: "worker-6b8c9d7f4-lm3nq", Ready: "1/2", Status: "Running", Restarts: 11, CPU: 12, Mem: 96, Age: 3 * time.Hour, Node: "pool-2-a"},
		{Namespace: "acme", Name: "migrate-2026090501-nn8ck", Ready: "0/1", Status: "Pending", CPU: 0, Mem: 0, Age: 2 * time.Minute},
		{Namespace: "acme-jobs", Name: "reporting-5f7a2c1d9-t8wz4", Ready: "0/1", Status: "CrashLoopBackOff", Restarts: 94, CPU: 4, Mem: 41, Age: 45 * time.Hour, Node: "pool-2-b"},
		{Namespace: "acme-jobs", Name: "nightly-2026090400-qq21x", Ready: "0/1", Status: "Completed", CPU: 0, Mem: 0, Age: 22 * time.Hour, Node: "pool-2-b"},
		{Namespace: "ingress", Name: "traefik-9c4b7f1a2-hh8kd", Ready: "1/1", Status: "Running", CPU: 61, Mem: 128, Age: 9 * 24 * time.Hour, Node: "pool-1-a"},
		{Namespace: "ingress", Name: "cert-manager-7b1d2e3f4-mm5rt", Ready: "1/1", Status: "Running", CPU: 8, Mem: 64, Age: 9 * 24 * time.Hour, Node: "pool-1-b"},
	}
}

// Columns is the table's header, in order. The index into this is what
// comp.Sort sorts by.
func Columns() []string {
	return []string{"NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "CPU", "MEM", "AGE"}
}

// Describe is `kubectl describe` for a pod, which k9s shows in a viewer.
func Describe(p Pod) string {
	return "Name:             " + p.Name + "\n" +
		"Namespace:        " + p.Namespace + "\n" +
		"Priority:         0\n" +
		"Node:             " + p.Node + "/10.132.0.14\n" +
		"Status:           " + p.Status + "\n" +
		"IP:               10.4.1.23\n" +
		"Controlled By:    ReplicaSet/" + p.Name + "\n" +
		"Containers:\n" +
		"  app:\n" +
		"    Image:          eu.gcr.io/acme/app:2026.9.1\n" +
		"    Port:           3000/TCP\n" +
		"    State:          " + p.Status + "\n" +
		"    Ready:          " + p.Ready + "\n" +
		"    Limits:\n" +
		"      cpu:     1\n" +
		"      memory:  1Gi\n" +
		"Conditions:\n" +
		"  Type              Status\n" +
		"  Initialized       True\n" +
		"  Ready             True\n" +
		"Events:\n" +
		"  Type     Reason     Age   From      Message\n" +
		"  Normal   Scheduled  12h   default   Successfully assigned\n" +
		"  Warning  BackOff    3m    kubelet   Back-off restarting failed container"
}

// Commands is what the `:` bar completes against.
func Commands() []string {
	return []string{"pods", "deployments", "services", "nodes", "namespaces", "configmaps", "secrets", "events"}
}
