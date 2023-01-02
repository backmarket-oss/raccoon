package k8s

import (
	"context"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestTTLFromPod(t *testing.T) {
	t.Parallel()

	type unitData struct {
		pod         v1.Pod
		defaultTTL  int
		expectedTTL int
		err         error
	}

	data := map[string]unitData{
		"use annotation": {
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"backmarket.com/raccoon-ttl": "45",
					},
				},
			},
			defaultTTL:  60,
			expectedTTL: 45,
			err:         nil,
		},
		"no annotation, default ttl": {
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"app": "test",
					},
				},
			},
			defaultTTL:  60,
			expectedTTL: 60,
			err:         nil,
		},
		"wrong annotation format": {
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"backmarket.com/raccoon-ttl": "wrong",
					},
				},
			},
			defaultTTL:  60,
			expectedTTL: -1,
			err:         fmt.Errorf("strconv.Atoi: parsing \"wrong\": invalid syntax"),
		},
	}

	for name, unit := range data {
		t.Run(name, func(unit unitData) func(t *testing.T) {
			return func(t *testing.T) {

				ttl, err := TTLFromPod(unit.pod, unit.defaultTTL)

				if err != nil {
					if unit.err == nil {
						t.Fatalf(err.Error())
					}

					if err.Error() != unit.err.Error() {
						t.Fatalf("expected err: %v, got: %v", unit.err.Error(), err.Error())
					}
				} else {

					if ttl != unit.expectedTTL {
						t.Fatalf("expected ttl: %v, got: %v", unit.expectedTTL, ttl)
					}
				}
			}
		}(unit))
	}
}

func TestListPods(t *testing.T) {
	t.Parallel()

	type unitData struct {
		clientSet      kubernetes.Interface
		labelSelector  string
		inputNamespace string
		sortedPodsName []string
	}

	initialTime, err := time.Parse(time.RFC3339, "2022-06-01T22:08:41+02:00")
	if err != nil {
		t.Fatalf("Can't parse initialTime: %v", err)
	}

	data := map[string]unitData{
		"test multiple pods order": {
			clientSet: testclient.NewSimpleClientset(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-1",
					Namespace:         "ns1",
					CreationTimestamp: metav1.NewTime(initialTime),
					Labels: map[string]string{
						"app": "test",
					},
				},
			}, &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-2",
					Namespace:         "ns1",
					CreationTimestamp: metav1.NewTime(initialTime.Add(-time.Hour * 1)),
					Labels: map[string]string{
						"app": "test",
					},
				},
			}),
			labelSelector:  "app=test",
			inputNamespace: "ns1",
			sortedPodsName: []string{"pod-2", "pod-1"},
		},
		"test single pod": {
			clientSet: testclient.NewSimpleClientset(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-1",
					Namespace:         "ns1",
					CreationTimestamp: metav1.NewTime(initialTime),
					Labels: map[string]string{
						"app": "test",
					},
				},
			}),
			labelSelector:  "app=test",
			inputNamespace: "ns1",
			sortedPodsName: []string{"pod-1"},
		},
		"no pod": {
			clientSet:      testclient.NewSimpleClientset(),
			labelSelector:  "app=test",
			inputNamespace: "ns1",
			sortedPodsName: []string{},
		},
	}

	for name, unit := range data {
		t.Run(name, func(unit unitData) func(t *testing.T) {
			return func(t *testing.T) {
				k8sClient := InitKubernetesClient(unit.clientSet)

				pods, _ := k8sClient.ListPods(context.Background(), unit.inputNamespace, unit.labelSelector)

				for i, pod := range pods {
					if unit.sortedPodsName[i] != pod.Name {
						t.Fatalf("exepcted pod: %v, got: %v",
							unit.sortedPodsName[i], pod.Name)
					}
				}
			}
		}(unit))
	}
}
