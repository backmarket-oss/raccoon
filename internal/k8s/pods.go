package k8s

import (
	"context"
	"sort"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesClient struct {
	clientSet kubernetes.Interface
}

// InitKubernetesClient inits a KubernetesClient.
func InitKubernetesClient(clientSet kubernetes.Interface) *KubernetesClient {
	return &KubernetesClient{clientSet: clientSet}
}

// ListPods returns a list of pods corresponding to the parameters you set.
// Pods are sorted by age in descending order.
func (k KubernetesClient) ListPods(ctx context.Context, namespace, labelSelector string) ([]v1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	pods, err := k.clientSet.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	sortPodByAgeDesc(pods)

	return pods.Items, nil
}

// DeletePod deletes pods based on namespace & pod's name. Uses foreground deletion policy.
func (k KubernetesClient) DeletePod(ctx context.Context, namespace, name string) error {
	deleteFg := metav1.DeletePropagationForeground

	return k.clientSet.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deleteFg,
	})
}

// EvictPod evicts pods based on namespace & pod's name. Uses foreground deletion policy.
func (k KubernetesClient) EvictPod(ctx context.Context, namespace, name string) error {
	deleteFg := metav1.DeletePropagationForeground

	return k.clientSet.PolicyV1beta1().Evictions(namespace).Evict(ctx, &policy.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		DeleteOptions: &metav1.DeleteOptions{
			PropagationPolicy: &deleteFg,
		},
	})
}

// Return date in seconds from v1.Pod object.
func DateFromPodInSecond(pod v1.Pod) float64 {
	seconds := time.Since(pod.ObjectMeta.CreationTimestamp.Time).Truncate(time.Second).Seconds()
	return seconds
}

// Sort a list of v1.Pod by age in descending order.
func sortPodByAgeDesc(pods *v1.PodList) *v1.PodList {
	sort.Slice(pods.Items, func(i, j int) bool {
		return DateFromPodInSecond(pods.Items[i]) > DateFromPodInSecond(pods.Items[j])
	})
	return pods
}

func TTLFromPod(pod v1.Pod, defaultTTL time.Duration) (time.Duration, error) {
	if annotations := pod.ObjectMeta.GetAnnotations(); annotations != nil {
		if ttlString := annotations["backmarket.com/raccoon-ttl"]; ttlString != "" {
			ttl, err := time.ParseDuration(ttlString)
			if err != nil {
				return time.Duration(0), err
			}
			return ttl, nil
		}
	}
	return defaultTTL, nil
}
