package strategy

import (
	"context"
	"math/rand"
	"time"

	"github.com/backmarket-oss/raccoon/internal"
	"github.com/backmarket-oss/raccoon/internal/k8s"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type k8sClient interface {
	ListPods(ctx context.Context, namespace, labelSelector string) ([]v1.Pod, error)
	EvictPod(ctx context.Context, namespace, name string) error
}

type namespacedPod struct {
	name      string
	namespace string
}

type RandomizedDelay struct {
	defaultSettings *internal.DefaultSettings
	maxDelay        int
	collector       chan *namespacedPod
	randomizer      *rand.Rand
	k8sClient       k8sClient
}

var (
	podsDeleted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "raccoon_pods_deleted_total",
			Help: "The total number of deleted pods",
		},
		[]string{"namespace"})
)

// InitRandomizedDelay initializes RandomizedDelay struct.
func InitRandomizedDelay(ctx context.Context, maxDelay int,
	dSettings *internal.DefaultSettings, k8sClient k8sClient) *RandomizedDelay {

	rndSource := rand.NewSource(time.Now().UnixNano())
	delay := &RandomizedDelay{
		defaultSettings: dSettings,
		maxDelay:        maxDelay,
		collector:       make(chan *namespacedPod),
		randomizer:      rand.New(rndSource),
		k8sClient:       k8sClient,
	}

	go delay.collectEventLoop(ctx)

	return delay
}

// Run requests k8s api to retrieve pods with an age older than the ttl.
// It sends pods' name to an internal channel.
// The sending action isn't blocking.
func (d RandomizedDelay) Run(ctx context.Context) error {
	return findPodsToCollect(ctx, d.k8sClient, d.defaultSettings.Namespace,
		d.defaultSettings.Selector, d.defaultSettings.TTL, d.collector)
}

func findPodsToCollect(ctx context.Context, k8sClient k8sClient,
	namespace, selector string, defaultTTL int, collector chan *namespacedPod) error {
	pods, err := k8sClient.ListPods(ctx, namespace, selector)
	if err != nil {
		return err
	}
	for _, pod := range pods {
		configuredTTL, err := k8s.TTLFromPod(pod, defaultTTL)
		if err != nil {
			return err
		}
		nsPod := &namespacedPod{
			name:      pod.ObjectMeta.Name,
			namespace: pod.ObjectMeta.Namespace,
		}

		tDiff := k8s.DateFromPodInSecond(pod)
		lFields := logrus.Fields{
			"namespace": nsPod.namespace,
			"selector":  selector,
			"pod":       pod.ObjectMeta.Name,
			"age":       tDiff,
		}
		log.WithFields(lFields).Debug("checking pod's age")

		if tDiff > configuredTTL {
			select {
			case collector <- nsPod:
				log.WithFields(lFields).Info("pod's age greater than ttl, marking pod")
			case <-ctx.Done():
			}
		}
	}

	return nil
}

// collect listen to the internal channel for pods to delete.
// This is where it applies the randomized delay between consecutive deletion.
func (d *RandomizedDelay) collectEventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case markedPod := <-d.collector:
			collectMarkedPod(ctx, d.defaultSettings.DryRun, d.k8sClient, *markedPod)
			/*
			   here we apply a randomized delay before going to the next iteration.
			   we don't want to use 0 as minimum value to avoid too short period between deletion.
			   maxDelay / 2 should be enough.
			*/
			minRandomized := d.maxDelay / 2
			randomizedDelay := d.randomizer.Intn(d.maxDelay-minRandomized+1) + minRandomized
			log.WithFields(logrus.Fields{
				"delay": randomizedDelay,
			}).Debug("waiting randomized delay")
			waitRandomizedDelay(ctx, randomizedDelay)
		}
	}
}

func collectMarkedPod(ctx context.Context, isDryRun bool,
	k8sClient k8sClient, markedPod namespacedPod) {
	lFields := logrus.Fields{
		"pod":       markedPod.name,
		"namespace": markedPod.namespace,
	}
	log.WithFields(lFields).Debug("new pod to collect")

	if !isDryRun {
		err := k8sClient.EvictPod(ctx, markedPod.namespace, markedPod.name)
		if err != nil {
			log.WithFields(lFields).Errorf("error while deleting pod: %v", err)
		} else {
			log.WithFields(lFields).Info("pod deleted")
			podsDeleted.With(prometheus.Labels{"namespace": markedPod.namespace}).Inc()
		}
	} else {
		log.WithFields(lFields).Debug("dry-run, pod should have been deleted")
	}
}

func waitRandomizedDelay(ctx context.Context, delay int) {
	select {
	case <-time.After(time.Duration(delay) * time.Second):
	case <-ctx.Done():
		//to avoid waiting the end of randomized delay when we trap a SIGTERM signal
		return
	}
}
