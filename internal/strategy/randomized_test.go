package strategy

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sClientMock struct {
	mock.Mock
}

func (m *K8sClientMock) ListPods(ctx context.Context, namespace, labelSelector string) ([]v1.Pod, error) {
	args := m.Called(ctx, namespace, labelSelector)
	return args.Get(0).([]v1.Pod), args.Error(1)
}

func (m *K8sClientMock) EvictPod(ctx context.Context, namespace, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}

func TestFindPodsToCollect(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	type unitData struct {
		pods                []v1.Pod
		namespace, selector string
		defaultTTL          int
		shouldBeMarked      map[string]bool
	}

	data := map[string]unitData{
		"pod-1 too young, pod-2 collectable, pod-3 collectable": {
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod-1",
						Namespace:         "namespace-1",
						CreationTimestamp: metav1.NewTime(time.Now()),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod-2",
						Namespace:         "namespace-1",
						CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second * -500)),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod-3",
						Namespace:         "namespace-1",
						CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second * -400)),
					},
				},
			},
			namespace:      "namespace-2",
			selector:       "app=app-1",
			defaultTTL:     360,
			shouldBeMarked: map[string]bool{"pod-1": false, "pod-2": true, "pod-3": true},
		},
		"all pods collectable": {
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod-1",
						Namespace:         "namespace-2",
						CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second * -6000)),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod-2",
						Namespace:         "namespace-2",
						CreationTimestamp: metav1.NewTime(time.Now().Add(time.Second * -4000)),
					},
				},
			},
			namespace:      "namespace-2",
			selector:       "app=app-1",
			defaultTTL:     3600,
			shouldBeMarked: map[string]bool{"pod-1": true, "pod-2": true},
		},
		"all pods too young": {
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod-1",
						Namespace:         "namespace-2",
						CreationTimestamp: metav1.NewTime(time.Now()),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "pod-2",
						Namespace:         "namespace-2",
						CreationTimestamp: metav1.NewTime(time.Now()),
					},
				},
			},
			namespace:      "namespace-2",
			selector:       "app=app-1",
			defaultTTL:     3600,
			shouldBeMarked: map[string]bool{"pod-1": false, "pod-2": false},
		},
	}

	for name, unit := range data {
		t.Run(name, func(unit unitData) func(t *testing.T) {
			return func(t *testing.T) {
				k8sMock := new(K8sClientMock)
				ctx := context.Background()
				collector := make(chan *namespacedPod)

				k8sMock.On("ListPods", ctx, unit.namespace, unit.selector).Return(unit.pods, nil)

				//synchronization primitive to make sure channel has finished its work
				wgClosed := new(sync.WaitGroup)
				wgClosed.Add(1)
				//emulate collect behavior to test if pods marked as to be collected are the good ones
				var podsToCollect []namespacedPod

				isCollected := func(name string) bool {
					for _, pod := range podsToCollect {
						if name == pod.name {
							return true
						}
					}
					return false
				}

				go func() {
					for ns := range collector {
						if ns != nil {
							podsToCollect = append(podsToCollect, *ns)
						}
					}
					wgClosed.Done()
				}()

				err := findPodsToCollect(ctx, k8sMock, unit.namespace, unit.selector, unit.defaultTTL, collector)
				close(collector)

				wgClosed.Wait()
				//are the pods collected, the good ones?
				for name, isExpected := range unit.shouldBeMarked {
					assert.Equal(isExpected, isCollected(name), name)
				}

				assert.Nil(err)
				k8sMock.AssertExpectations(t)
			}
		}(unit))
	}
}

func TestCollect(t *testing.T) {
	t.Parallel()

	type unitData struct {
		dryRun    bool
		markedPod namespacedPod
	}

	data := map[string]unitData{
		"dry run activated": {
			dryRun: true,
			markedPod: namespacedPod{
				name:      "pod-1",
				namespace: "namespace-1",
			},
		},
		"dry run deactivated": {
			dryRun: false,
			markedPod: namespacedPod{
				name:      "pod-2",
				namespace: "namespace-1",
			},
		},
	}

	for name, unit := range data {
		t.Run(name, func(unit unitData) func(t *testing.T) {
			return func(t *testing.T) {
				k8sMock := new(K8sClientMock)
				ctx := context.Background()

				if !unit.dryRun {
					k8sMock.On("EvictPod", ctx, unit.markedPod.namespace, unit.markedPod.name).Return(nil)
				}
				collectMarkedPod(ctx, unit.dryRun, k8sMock, unit.markedPod)
				k8sMock.AssertExpectations(t)
			}
		}(unit))
	}
}
