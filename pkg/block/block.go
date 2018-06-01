package block

import (
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

var version = "0.2.0"

type block struct {
	driver *csicommon.CSIDriver

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

type blockVolume struct {
	VolName string
	VolID   string
	// VolSizeRequest: unit GB
	VolSizeRequest int
	// VolSizeCapacity: unit GB
	VolSizeCapacity int
}

func GetBlockDriver() *block {
	return &block{}
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func NewControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}

func NewNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
	}
}

func (blk *block) Run(driverName, nodeID, endpoint string) {
	glog.Infof("Driver: %v version: %v", driverName, version)

	// Initialize default library driver
	blk.driver = csicommon.NewCSIDriver(driverName, version, nodeID)
	if blk.driver == nil {
		glog.Fatalln("Failed to initialize CSI Driver.")
	}
	blk.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	blk.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	// Create GRPC servers
	blk.ids = NewIdentityServer(blk.driver)
	blk.ns = NewNodeServer(blk.driver)
	blk.cs = NewControllerServer(blk.driver)

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(endpoint, blk.ids, blk.cs, blk.ns)
	s.Wait()
}
