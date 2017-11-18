package tests

import (
	"github.com/hazelcast/go-client"
	"github.com/hazelcast/go-client/core"
	"github.com/hazelcast/go-client/internal"
	"github.com/hazelcast/go-client/internal/common"
	. "github.com/hazelcast/go-client/rc"
	"log"
	"sync"
	"testing"
	"time"
)

type membershipListener struct {
	wg *sync.WaitGroup
}

func (membershipListener *membershipListener) MemberAdded(member core.IMember) {
	membershipListener.wg.Done()
}
func (membershipListener *membershipListener) MemberRemoved(member core.IMember) {
	membershipListener.wg.Done()
}

var remoteController *RemoteControllerClient
var cluster *Cluster

func TestMain(m *testing.M) {
	rc, err := NewRemoteControllerClient("localhost:9701")
	remoteController = rc
	if remoteController == nil || err != nil {
		log.Fatal("create remote controller failed:", err)
	}
	m.Run()
}
func TestInitialMembershipListener(t *testing.T) {
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	remoteController.StartMember(cluster.ID)
	config := hazelcast.NewHazelcastConfig()
	config.AddMembershipListener(&membershipListener{wg: wg})
	wg.Add(1)
	client, _ := hazelcast.NewHazelcastClientWithConfig(config)
	timeout := WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "Cluster initialMembershipListener failed")
	client.Shutdown()
	remoteController.ShutdownCluster(cluster.ID)
}
func TestMemberAddedandRemoved(t *testing.T) {
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	remoteController.StartMember(cluster.ID)
	config := hazelcast.NewHazelcastConfig()
	config.AddMembershipListener(&membershipListener{wg: wg})
	wg.Add(1)
	client, _ := hazelcast.NewHazelcastClientWithConfig(config)
	timeout := WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "Cluster initialMembershipListener failed")
	wg.Add(1)
	member, _ := remoteController.StartMember(cluster.ID)
	timeout = WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "Cluster memberAdded failed")
	wg.Add(1)
	remoteController.ShutdownMember(cluster.ID, member.UUID)
	timeout = WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "Cluster memberRemoved failed")
	client.Shutdown()
	remoteController.ShutdownCluster(cluster.ID)
}
func TestAddListener(t *testing.T) {
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	remoteController.StartMember(cluster.ID)
	client, _ := hazelcast.NewHazelcastClient()
	wg.Add(1)
	registrationId := client.GetCluster().AddListener(&membershipListener{wg: wg})
	member, _ := remoteController.StartMember(cluster.ID)
	timeout := WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "Cluster initialMembershipListener failed")
	client.GetCluster().RemoveListener(registrationId)
	wg.Add(1)
	member2, _ := remoteController.StartMember(cluster.ID)
	timeout = WaitTimeout(wg, Timeout/20)
	AssertEqualf(t, nil, true, timeout, "Cluster RemoveListener failed")
	remoteController.ShutdownMember(cluster.ID, member.UUID)
	registrationId = client.GetCluster().AddListener(&membershipListener{wg: wg})
	remoteController.ShutdownMember(cluster.ID, member2.UUID)
	timeout = WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "Cluster memberRemoved failed")
	client.GetCluster().RemoveListener(registrationId)
	client.Shutdown()
	remoteController.ShutdownCluster(cluster.ID)
}
func TestGetMembers(t *testing.T) {
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	member1, _ := remoteController.StartMember(cluster.ID)
	member2, _ := remoteController.StartMember(cluster.ID)
	member3, _ := remoteController.StartMember(cluster.ID)
	client, _ := hazelcast.NewHazelcastClient()
	members := client.GetCluster().GetMemberList()
	AssertEqualf(t, nil, len(members), 3, "GetMemberList returned wrong number of members")
	client.Shutdown()
	remoteController.ShutdownMember(cluster.ID, member1.UUID)
	remoteController.ShutdownMember(cluster.ID, member2.UUID)
	remoteController.ShutdownMember(cluster.ID, member3.UUID)
	remoteController.ShutdownCluster(cluster.ID)
}
func TestAuthenticationWithWrongCredentials(t *testing.T) {
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	remoteController.StartMember(cluster.ID)
	config := hazelcast.NewHazelcastConfig()
	config.GroupConfig().SetName("wrongName")
	config.GroupConfig().SetPassword("wrongPassword")
	client, err := hazelcast.NewHazelcastClientWithConfig(config)
	if _, ok := err.(*common.HazelcastAuthenticationError); !ok {
		t.Fatal("client should have returned an authentication error")
	}
	client.Shutdown()
	remoteController.ShutdownCluster(cluster.ID)
}
func TestClientWithoutMember(t *testing.T) {
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	client, err := hazelcast.NewHazelcastClient()
	if _, ok := err.(*common.HazelcastIllegalStateError); !ok {
		t.Fatal("client should have returned a hazelcastError")
	}
	client.Shutdown()
	remoteController.ShutdownCluster(cluster.ID)
}
func TestRestartMember(t *testing.T) {
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	member1, _ := remoteController.StartMember(cluster.ID)
	config := hazelcast.NewHazelcastConfig()
	config.ClientNetworkConfig().SetConnectionAttemptLimit(10)
	client, _ := hazelcast.NewHazelcastClientWithConfig(config)
	lifecycleListener := lifecycyleListener{wg: wg, collector: make([]string, 0)}
	wg.Add(1)
	registratonId := client.(*internal.HazelcastClient).LifecycleService.AddListener(&lifecycleListener)
	remoteController.ShutdownMember(cluster.ID, member1.UUID)
	timeout := WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "clusterService reconnect has failed")
	AssertEqualf(t, nil, lifecycleListener.collector[0], internal.LIFECYCLE_STATE_DISCONNECTED, "clusterService reconnect has failed")
	wg.Add(1)
	remoteController.StartMember(cluster.ID)
	timeout = WaitTimeout(wg, Timeout)
	AssertEqualf(t, nil, false, timeout, "clusterService reconnect has failed")
	AssertEqualf(t, nil, lifecycleListener.collector[1], internal.LIFECYCLE_STATE_CONNECTED, "clusterService reconnect has failed")
	client.GetLifecycle().RemoveListener(&registratonId)
	client.Shutdown()
	remoteController.ShutdownCluster(cluster.ID)
}
func TestReconnectToNewNodeViaLastMemberList(t *testing.T) {
	cluster, _ = remoteController.CreateCluster("3.9", DEFAULT_XML_CONFIG)
	oldMember, _ := remoteController.StartMember(cluster.ID)
	config := hazelcast.NewHazelcastConfig()
	config.ClientNetworkConfig().SetConnectionAttemptLimit(100)
	config.ClientNetworkConfig().SetSmartRouting(false)
	client, _ := hazelcast.NewHazelcastClientWithConfig(config)
	newMember, _ := remoteController.StartMember(cluster.ID)
	remoteController.ShutdownMember(cluster.ID, oldMember.UUID)
	time.Sleep(10 * time.Second)
	memberList := client.GetCluster().GetMemberList()
	AssertEqualf(t, nil, len(memberList), 1, "client did not use the last member list to reconnect")
	AssertEqualf(t, nil, memberList[0].Uuid(), newMember.UUID, "client did not use the last member list to reconnect uuid")
	remoteController.ShutdownCluster(cluster.ID)
	client.Shutdown()
}

type mapListener struct {
	wg *sync.WaitGroup
}

func (ml *mapListener) EntryAdded(event core.IEntryEvent) {
	ml.wg.Done()
}
