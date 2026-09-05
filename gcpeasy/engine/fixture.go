package engine

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FixtureAt is the moment a fixture is anchored to, so ages hold still and a
// golden does not fail a minute after it was written.
var FixtureAt = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)

// Fixture makes the engine answer from canned data instead of the cloud, and
// returns a function that puts the real one back.
//
// It replaces the RUNNER rather than the reader, so `Projects` and `Pods` parse
// the same JSON with the same code they parse gcloud's with. A fixture that
// returned []Pod directly would be a second implementation, and the parsing —
// which is where the bugs are — would never be exercised by a test.
func Fixture() func() {
	restore := SetRunner(fixtureRun)
	restoreClock := SetClock(func() time.Time { return FixtureAt })
	return func() { restore(); restoreClock() }
}

func fixtureRun(_ context.Context, name string, args ...string) ([]byte, error) {
	line := name + " " + strings.Join(args, " ")
	switch {
	case strings.HasPrefix(line, "gcloud config get-value project"):
		return []byte("acme-staging\n"), nil
	case strings.HasPrefix(line, "gcloud projects list"):
		return []byte(fixtureProjects), nil
	case strings.HasPrefix(line, "gcloud container clusters list"):
		if strings.Contains(line, "acme-sandbox") {
			return []byte(`[]`), nil
		}
		return []byte(fixtureClusters), nil
	case strings.HasPrefix(line, "gcloud container clusters get-credentials"):
		return nil, nil
	case strings.HasPrefix(line, "kubectl get pods"):
		return []byte(fixturePods), nil
	case strings.HasPrefix(line, "kubectl logs"):
		return []byte(fixtureLogs), nil
	case strings.HasPrefix(line, "kubectl describe"):
		return []byte(fixtureDescribe), nil
	}
	return nil, fmt.Errorf("the fixture has no answer for %q", line)
}

const fixtureProjects = `[
 {"projectId":"acme-prod","name":"Acme Production","projectNumber":"411000000001"},
 {"projectId":"acme-staging","name":"Acme Staging","projectNumber":"411000000002"},
 {"projectId":"acme-sandbox","name":"Acme Sandbox","projectNumber":"411000000003"}
]`

const fixtureClusters = `[
 {"name":"apps-euw1","location":"europe-west1","currentNodeCount":6,
  "currentMasterVersion":"1.30.4-gke.1000","status":"RUNNING"},
 {"name":"batch-euw1-b","location":"europe-west1-b","currentNodeCount":2,
  "currentMasterVersion":"1.29.7-gke.1200","status":"RECONCILING"}
]`

// A pod set chosen to exercise every State, and to include the system
// namespaces Pod.App filters out — because a filter nothing tests is a filter
// that stops working quietly.
const fixturePods = `{"items":[
 {"metadata":{"name":"web-7d9f4c8b5-2xk9p","namespace":"acme","creationTimestamp":"2026-09-04T21:00:00Z"},
  "spec":{"nodeName":"gke-apps-euw1-pool-1-a","containers":[{"image":"eu.gcr.io/acme/web:2026.9.1"}]},
  "status":{"phase":"Running","containerStatuses":[{"ready":true,"restartCount":0}]}},
 {"metadata":{"name":"web-7d9f4c8b5-p4m2q","namespace":"acme","creationTimestamp":"2026-09-04T21:00:00Z"},
  "spec":{"nodeName":"gke-apps-euw1-pool-1-b","containers":[{"image":"eu.gcr.io/acme/web:2026.9.1"}]},
  "status":{"phase":"Running","containerStatuses":[{"ready":true,"restartCount":3}]}},
 {"metadata":{"name":"worker-6b8c9d7f4-lm3nq","namespace":"acme","creationTimestamp":"2026-09-05T05:30:00Z"},
  "spec":{"nodeName":"gke-apps-euw1-pool-2-a","containers":[{"image":"eu.gcr.io/acme/worker:2026.9.1"},{"image":"eu.gcr.io/acme/sidecar:1.4.0"}]},
  "status":{"phase":"Running","containerStatuses":[{"ready":true,"restartCount":0},{"ready":false,"restartCount":11}]}},
 {"metadata":{"name":"migrate-2026090501-nn8ck","namespace":"acme","creationTimestamp":"2026-09-05T08:58:00Z"},
  "spec":{"nodeName":"","containers":[{"image":"eu.gcr.io/acme/web:2026.9.1"}]},
  "status":{"phase":"Pending","containerStatuses":[]}},
 {"metadata":{"name":"reporting-5f7a2c1d9-t8wz4","namespace":"acme-jobs","creationTimestamp":"2026-09-03T11:15:00Z"},
  "spec":{"nodeName":"gke-apps-euw1-pool-2-b","containers":[{"image":"eu.gcr.io/acme/reporting:2026.8.7"}]},
  "status":{"phase":"CrashLoopBackOff","containerStatuses":[{"ready":false,"restartCount":94}]}},
 {"metadata":{"name":"fluentbit-gke-x7q2w","namespace":"kube-system","creationTimestamp":"2026-08-20T10:00:00Z"},
  "spec":{"nodeName":"gke-apps-euw1-pool-1-a","containers":[{"image":"gke.gcr.io/fluent-bit:1.9.10"}]},
  "status":{"phase":"Running","containerStatuses":[{"ready":true,"restartCount":0}]}}
]}`

const fixtureLogs = `I, [2026-09-05T08:59:12.004] INFO -- : Started GET "/healthz" for 10.4.0.1
I, [2026-09-05T08:59:12.006] INFO -- : Completed 200 OK in 2ms (Views: 0.1ms)
I, [2026-09-05T08:59:18.881] INFO -- : Started POST "/api/v2/orders" for 10.4.2.7
I, [2026-09-05T08:59:19.402] INFO -- : Completed 201 Created in 521ms (ActiveRecord: 61ms)
W, [2026-09-05T08:59:31.117] WARN -- : Slow query (1.8s): SELECT "orders".* FROM "orders"
I, [2026-09-05T08:59:44.230] INFO -- : Started GET "/healthz" for 10.4.0.1
E, [2026-09-05T08:59:52.918] ERROR -- : PG::ConnectionBad: could not connect to server
I, [2026-09-05T09:00:00.010] INFO -- : Completed 200 OK in 3ms (Views: 0.2ms)`

const fixtureDescribe = `Name:             web-7d9f4c8b5-2xk9p
Namespace:        acme
Priority:         0
Service Account:  web
Node:             gke-apps-euw1-pool-1-a/10.132.0.14
Start Time:       Thu, 04 Sep 2026 21:00:04 +0000
Labels:           app=web
                  pod-template-hash=7d9f4c8b5
Status:           Running
IP:               10.4.1.23
Containers:
  web:
    Image:          eu.gcr.io/acme/web:2026.9.1
    Port:           3000/TCP
    State:          Running
    Ready:          True
    Restart Count:  0
    Limits:
      cpu:     1
      memory:  1Gi
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
Events:
  Type    Reason     Age    From               Message
  ----    ------     ----   ----               -------
  Normal  Scheduled  12h    default-scheduler  Successfully assigned acme/web-7d9f4c8b5-2xk9p
  Normal  Pulled     12h    kubelet            Container image already present on machine
  Normal  Started    12h    kubelet            Started container web`
