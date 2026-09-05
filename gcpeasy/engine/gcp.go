// Package engine is gcpeasy's domain, and it has NO TERMINAL CONCEPTS AT ALL.
//
// No colour, no width, no keys, no framework, no tuikit import. guard.Engine
// holds that closed.
//
// # It also does not talk to the person
//
// Nothing here prints and nothing here reads stdin. That is a stronger rule
// than the import ban and it is the one the original gcpeasy broke: its
// `SelectCluster` prompted the user with fmt.Printf and bufio, which meant the
// TUI could not call it — a function that writes to stdout would draw over the
// frame and then block on a read that never comes. Selection had to be
// implemented a second time inside the interface.
//
// So a choice is a value handed back, never a question asked.
//
// # Running things is the caller's job
//
// [Console] and [Shell] return an *exec.Cmd rather than running it. A process
// is a domain idea; how to run it is not. The CLI runs one inline, and the TUI
// hands the terminal over with tea.Exec so `rails console` gets a real TTY with
// real line editing. Neither decision belongs here.
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Project is a GCP project.
type Project struct {
	ID     string
	Name   string
	Number string
	Active bool
}

// Cluster is a GKE cluster.
type Cluster struct {
	Name     string
	Location string
	Version  string
	Nodes    int
	Status   string
	Project  string
}

// Pod is a Kubernetes pod, in whatever cluster kubectl is pointed at.
type Pod struct {
	Name      string
	Namespace string
	Phase     string
	Ready     string
	Restarts  int
	Age       time.Duration
	Node      string
	Images    []string
}

// State is what a thing is doing, in the four terms the interface colours.
//
// A domain type rather than a colour, because "degraded" is a fact about a pod
// and "yellow" is a decision about a screen.
type State int

// The four states.
const (
	Ready State = iota
	Pending
	Degraded
	Failed
)

func (s State) String() string {
	switch s {
	case Ready:
		return "ready"
	case Pending:
		return "pending"
	case Degraded:
		return "degraded"
	case Failed:
		return "failed"
	}
	return "unknown"
}

// State maps a pod's phase and readiness onto the four.
//
// Running but not all containers ready is Degraded rather than Ready, because
// a pod serving from two of three containers is a pod somebody should look at.
func (p Pod) State() State {
	switch p.Phase {
	case "Running":
		if ready, want, ok := p.readiness(); ok && ready < want {
			return Degraded
		}
		if p.Restarts > 0 {
			return Degraded
		}
		return Ready
	case "Succeeded":
		return Ready
	case "Pending", "ContainerCreating":
		return Pending
	}
	return Failed
}

// readiness splits "2/3" into its parts.
func (p Pod) readiness() (ready, want int, ok bool) {
	a, b, found := strings.Cut(p.Ready, "/")
	if !found {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(a, "%d", &ready); err != nil {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(b, "%d", &want); err != nil {
		return 0, 0, false
	}
	return ready, want, true
}

// App reports whether this looks like an application pod rather than part of
// the cluster's own machinery.
//
// The original gcpeasy had the same rule and called it isSystemNamespace. It is
// a judgement about Kubernetes, so it belongs here and not in the interface.
func (p Pod) App() bool {
	switch p.Namespace {
	case "kube-system", "kube-public", "kube-node-lease", "gke-system",
		"gmp-system", "gke-managed-system", "istio-system", "cert-manager":
		return false
	}
	return true
}

// Snapshot is everything gcpeasy knows, at one moment.
type Snapshot struct {
	Projects []Project
	Clusters []Cluster
	Pods     []Pod
	Fetched  time.Time

	// Partial says a read failed and what is here is the rest. A cluster that
	// cannot be reached should not blank the project list beside it.
	Partial error
}

// Healthy reports that no pod needs attention.
func (s Snapshot) Healthy() bool {
	for _, p := range s.Pods {
		if p.State() != Ready {
			return false
		}
	}
	return true
}

// Namespaces is the distinct namespaces of the pods, sorted.
//
// Sorted rather than map order, because an interface that reorders itself
// between runs is one nobody can screenshot.
func (s Snapshot) Namespaces() []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range s.Pods {
		if !seen[p.Namespace] {
			seen[p.Namespace] = true
			out = append(out, p.Namespace)
		}
	}
	sort.Strings(out)
	return out
}

// Projects reads the projects the signed-in account can see.
func Projects(ctx context.Context) ([]Project, error) {
	out, err := run(ctx, "gcloud", "projects", "list", "--format=json")
	if err != nil {
		return nil, err
	}
	var raw []struct {
		ProjectID     string `json:"projectId"`
		Name          string `json:"name"`
		ProjectNumber string `json:"projectNumber"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("reading the project list: %w", err)
	}
	current, _ := CurrentProject(ctx)
	projects := make([]Project, 0, len(raw))
	for _, r := range raw {
		projects = append(projects, Project{
			ID: r.ProjectID, Name: r.Name, Number: r.ProjectNumber,
			Active: r.ProjectID == current,
		})
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })
	return projects, nil
}

// CurrentProject is the project gcloud is configured for.
func CurrentProject(ctx context.Context) (string, error) {
	out, err := run(ctx, "gcloud", "config", "get-value", "project")
	return strings.TrimSpace(string(out)), err
}

// Clusters reads the GKE clusters in a project.
func Clusters(ctx context.Context, project string) ([]Cluster, error) {
	out, err := run(ctx, "gcloud", "container", "clusters", "list",
		"--project", project, "--format=json")
	if err != nil {
		return nil, err
	}
	var raw []struct {
		Name             string `json:"name"`
		Location         string `json:"location"`
		CurrentNodeCount int    `json:"currentNodeCount"`
		CurrentMasterVer string `json:"currentMasterVersion"`
		Status           string `json:"status"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("reading the cluster list: %w", err)
	}
	clusters := make([]Cluster, 0, len(raw))
	for _, r := range raw {
		clusters = append(clusters, Cluster{
			Name: r.Name, Location: r.Location, Version: r.CurrentMasterVer,
			Nodes: r.CurrentNodeCount, Status: r.Status, Project: project,
		})
	}
	return clusters, nil
}

// UseCluster points kubectl at a cluster.
//
// Named for what it does to the world rather than for how — the original called
// this ConfigureKubectl, which says which tool it shells out to and not what
// changes.
func UseCluster(ctx context.Context, c Cluster) error {
	args := []string{"container", "clusters", "get-credentials", c.Name,
		"--project", c.Project}
	if strings.Count(c.Location, "-") >= 2 {
		args = append(args, "--zone", c.Location)
	} else {
		args = append(args, "--region", c.Location)
	}
	_, err := run(ctx, "gcloud", args...)
	return err
}

// Pods reads every pod kubectl can see.
func Pods(ctx context.Context) ([]Pod, error) {
	out, err := run(ctx, "kubectl", "get", "pods", "--all-namespaces", "-o=json")
	if err != nil {
		return nil, err
	}
	var raw struct {
		Items []struct {
			Metadata struct {
				Name              string    `json:"name"`
				Namespace         string    `json:"namespace"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Spec struct {
				NodeName   string `json:"nodeName"`
				Containers []struct {
					Image string `json:"image"`
				} `json:"containers"`
			} `json:"spec"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					Ready        bool `json:"ready"`
					RestartCount int  `json:"restartCount"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("reading the pod list: %w", err)
	}

	now := clock()
	pods := make([]Pod, 0, len(raw.Items))
	for _, i := range raw.Items {
		var ready, restarts int
		for _, cs := range i.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
			restarts += cs.RestartCount
		}
		images := make([]string, 0, len(i.Spec.Containers))
		for _, c := range i.Spec.Containers {
			images = append(images, c.Image)
		}
		pods = append(pods, Pod{
			Name:      i.Metadata.Name,
			Namespace: i.Metadata.Namespace,
			Phase:     i.Status.Phase,
			Ready:     fmt.Sprintf("%d/%d", ready, len(i.Status.ContainerStatuses)),
			Restarts:  restarts,
			Age:       now.Sub(i.Metadata.CreationTimestamp),
			Node:      i.Spec.NodeName,
			Images:    images,
		})
	}
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].Namespace != pods[j].Namespace {
			return pods[i].Namespace < pods[j].Namespace
		}
		return pods[i].Name < pods[j].Name
	})
	return pods, nil
}

// Logs reads a pod's recent output.
func Logs(ctx context.Context, p Pod, lines int) ([]byte, error) {
	return run(ctx, "kubectl", "logs", p.Name, "-n", p.Namespace,
		fmt.Sprintf("--tail=%d", lines))
}

// Describe reads everything Kubernetes will say about a pod.
func Describe(ctx context.Context, p Pod) ([]byte, error) {
	return run(ctx, "kubectl", "describe", "pod", p.Name, "-n", p.Namespace)
}

// Console is the command that opens a Rails console in a pod.
//
// Returned rather than run, which is the whole point. A console needs a real
// terminal — line editing, history, colour, Ctrl-C reaching Ruby rather than
// gcpeasy — and only the caller knows how to give it one. The TUI suspends
// itself with tea.Exec; the CLI is already at a terminal and just runs it.
//
// The fallback chain is a fact about Rails deployments rather than about
// terminals: `bundle exec rails console` is right in most images and wrong in
// enough of them to be worth trying six ways before dropping to a shell.
func Console(ctx context.Context, p Pod) *exec.Cmd {
	attempts := []string{
		"bundle exec rails console",
		"bundle exec rails c",
		"rails console",
		"rails c",
		"bin/rails console",
		"bin/rails c",
	}
	chain := make([]string, 0, len(attempts))
	for _, a := range attempts {
		chain = append(chain, "exec "+a)
	}
	script := "(" + strings.Join(chain, " || ") + ")"
	return Shell(ctx, p, script)
}

// Shell is the command that opens a shell in a pod, optionally running a
// script instead of an interactive login.
func Shell(ctx context.Context, p Pod, script string) *exec.Cmd {
	args := []string{"exec", "-it", p.Name, "-n", p.Namespace, "--"}
	if script == "" {
		args = append(args, "/bin/sh", "-lc", "exec /bin/bash || exec /bin/sh")
	} else {
		args = append(args, "/bin/sh", "-lc", script)
	}
	return exec.CommandContext(ctx, "kubectl", args...)
}

// Ago is a duration as a person would say it. Coarse on purpose: "4m ago" is
// what the reader wants, and "4m13.204s ago" is the same fact made unreadable.
func Ago(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
