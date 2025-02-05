package nds

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/shippomx/zard/core/logx"
	"github.com/golang/mock/gomock"
	"github.com/nacos-group/nacos-sdk-go/v2/mock"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
)

type MockINamingClient struct {
	*mock.MockINamingClient
}

// GetAllServicesInfo implements naming_client.INamingClient.
func (m *MockINamingClient) GetAllServicesInfo(_ vo.GetAllServiceInfoParam) (model.ServiceList, error) {
	panic("unimplemented")
}

// SelectAllInstances implements naming_client.INamingClient.
func (m *MockINamingClient) SelectAllInstances(_ vo.SelectAllInstancesParam) ([]model.Instance, error) {
	panic("unimplemented")
}

// ServerHealthy implements naming_client.INamingClient.
func (m *MockINamingClient) ServerHealthy() bool {
	panic("unimplemented")
}

func (m *MockINamingClient) UpdateInstance(_ vo.UpdateInstanceParam) (bool, error) {
	panic("unimplemented")
}

func NewMockINamingClient(t *testing.T) *MockINamingClient {
	ic := mock.NewMockINamingClient(gomock.NewController(gomock.TestReporter(t)))
	return &MockINamingClient{ic}
}

func (m *MockINamingClient) BatchRegisterInstance(_ vo.BatchRegisterInstanceParam) (bool, error) {
	panic("unimplemented")
}

func (m *MockINamingClient) CloseClient() {
	logx.Debug("mock nacos client close")
}

func GetMockNacosTransport(t *testing.T) *NacosTransport {
	sv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	url, err := url.Parse(sv.URL)
	assert.NoError(t, err)
	port, err := strconv.Atoi(url.Port())
	assert.NoError(t, err)
	n := &NacosTransport{}
	n.inited = true
	m := mock.NewMockINamingClient(gomock.NewController(t))
	iport, err := I64ToU64(int64(port))
	if err != nil {
		logx.Must(err)
	}
	m.EXPECT().SelectOneHealthyInstance(gomock.Any()).Return(&model.Instance{
		Ip:          url.Hostname(),
		Port:        iport,
		ServiceName: "test",
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		ClusterName: "test",
	}, nil).MinTimes(0)
	n.nacosClient = &MockINamingClient{m}
	n.config = &NacosDiscoveryConfig{
		GroupName: "test",
		Clusters:  []string{"test"},
	}
	return n
}
