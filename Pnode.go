package main

import (
	"io/ioutil"
	"os"
	"path"

	//"strings"
	//"sync"
	//"time"
	//"sort"
	//"encoding/hex"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"strings"

	"google.golang.org/grpc"

	"github.com/plan-systems/plan-core/tools"
	"github.com/plan-systems/plan-core/tools/ctx"
	"github.com/plan-systems/plan-core/plan"
	"github.com/plan-systems/plan-core/repo"
)

const (
	configFilename = "PnodeConfig.json"
)

// Config specifies all operating parameters if a Snode (PLAN's p2p/server node)
type Config struct {
	Name            string      `json:"node_name"`
	NodeID          tools.Bytes `json:"node_id"`
	DefaultFileMode os.FileMode `json:"default_file_mode"`
	GrpcNetworkName string      `json:"grpc_network_name"`
	GrpcNetworkAddr string      `json:"grpc_network_addr"`
	Version         int32       `json:"version"`
}

// ApplyDefaults sets std fields and values
func (config *Config) ApplyDefaults() {
	config.DefaultFileMode = plan.DefaultFileMode
	config.GrpcNetworkName = "tcp"
	config.GrpcNetworkAddr = ""
	config.Version = 1
}

// Pnode wraps one or more communities replicated to a local dir.
type Pnode struct {
	ctx.Context

	activeSessions ctx.SessionGroup

	BasePath  string
	ReposPath string
	Config    Config

	servicePort string
	grpcServer  *grpc.Server
}

// NewPnode creates a new Pnode
func NewPnode(
	inBasePath string,
	inDoInit bool,
	inServicePort string,
) (*Pnode, error) {

	pn := &Pnode{
		activeSessions: ctx.NewSessionGroup(),
		servicePort:    inServicePort,
	}
	pn.SetLogLabel("pnode")

	var err error
	if pn.BasePath, err = plan.SetupBaseDir(inBasePath, inDoInit); err != nil {
		return nil, err
	}

	pn.ReposPath = path.Join(pn.BasePath, "seeded")
	if err = os.MkdirAll(pn.ReposPath, plan.DefaultFileMode); err != nil {
		return nil, err
	}

	if err = pn.readConfig(inDoInit); err != nil {
		return nil, err
	}

	return pn, nil
}

// readConfig uses BasePath to read in the node's config file
func (pn *Pnode) readConfig(inFirstTime bool) error {

	pathname := path.Join(pn.BasePath, configFilename)

	buf, err := ioutil.ReadFile(pathname)
	if err == nil {
		err = json.Unmarshal(buf, &pn.Config)
	}
	if err != nil {
		if os.IsNotExist(err) && inFirstTime {
			pn.Config.ApplyDefaults()
			pn.Config.NodeID = make([]byte, plan.CommunityIDSz)
			crand.Read(pn.Config.NodeID)

			err = pn.writeConfig()
		} else {
			err = plan.Errorf(err, plan.ConfigFailure, "Failed to load pnode config")
		}
	}

	return err
}

// writeConfig writes out the node config file based on BasePath
func (pn *Pnode) writeConfig() error {

	buf, err := json.MarshalIndent(&pn.Config, "", "\t")
	if err == nil {
		pathname := path.Join(pn.BasePath, configFilename)

		err = ioutil.WriteFile(pathname, buf, pn.Config.DefaultFileMode)
	}

	if err != nil {
		return plan.Errorf(err, plan.FailedToAccessPath, "Failed to write node config")
	}

	return nil
}

// Startup -- see pcore.Flow.Startup
func (pn *Pnode) Startup() error {

	err := pn.CtxStart(
		pn.ctxStartup,
		nil,
		nil,
		pn.ctxStopping,
	)

	return err
}

func (pn *Pnode) ctxStartup() error {

	// TODO: test w/ sym links
	repoDirs, err := ioutil.ReadDir(pn.ReposPath)
	if err != nil {
		return err
	}

	for _, repoDir := range repoDirs {
		repoPath := repoDir.Name()
		if !strings.HasPrefix(repoPath, ".") {
			_, err = pn.createAndStartRepo(repoPath, nil)
			if err != nil {
				break
			}
		}
	}

	//
	//
	//
	// grpc service
	//
	if err == nil {
		pn.grpcServer = grpc.NewServer()
		repo.RegisterRepoServer(pn.grpcServer, pn)

		addr := pn.Config.GrpcNetworkAddr + ":" + pn.servicePort
		err = pn.AttachGrpcServer(
			pn.Config.GrpcNetworkName,
			addr,
			pn.grpcServer,
		)
	}

	return err
}

func (pn *Pnode) ctxStopping() {

}

func (pn *Pnode) createAndStartRepo(
	inRepoSubPath string,
	inSeed *repo.RepoSeed,
) (*repo.CommunityRepo, error) {

	var repoPath string
	var err error

	if inSeed != nil {
		// Only proceed if the dir doesn't exist
		// TODO: change dir name in the event of a name collision.
		repoPath, err = plan.CreateNewDir(pn.ReposPath, inRepoSubPath)

	} else {
		repoPath = path.Join(pn.ReposPath, inRepoSubPath)
	}

	if err != nil {
		return nil, err
	}

	CR, err := repo.NewCommunityRepo(repoPath, inSeed)
	if err != nil {
		return nil, err
	}

	err = CR.Startup()
	if err != nil {
		return nil, err
	}

	pn.Info(0, "mounted repo at ", repoPath)

	pn.CtxAddChild(CR, CR.GenesisSeed.StorageEpoch.CommunityID)

	return CR, nil
}

// seedRepo adds a new repo (if it doesn't already exist)
func (pn *Pnode) seedRepo(
	inSeed *repo.RepoSeed,
) error {

	//var CR *repo.CommunityRepo

	{
		genesis, err := inSeed.ExtractAndVerifyGenesisSeed()
		if err != nil {
			return err
		}

		// If the repo is already seed, nothing further required
		if pn.fetchRepo(genesis.StorageEpoch.CommunityID) != nil {
			return nil
		}
	}

	if !pn.CtxRunning() {
		return plan.Error(nil, plan.AssertFailed, "pnode must be running to seed a new repo")
	}

	// In the unlikely event that pn.Shutdown() is called while this is all happening,
	//    prevent the rug from being snatched out from under us.
	hold := make(chan struct{})
	defer func() {
		hold <- struct{}{}
	}()
	pn.CtxGo(func() {
		<-hold
	})

	// When we pass the seed, it means create from scratch
	CR, err := pn.createAndStartRepo(inSeed.SuggestedDirName, inSeed)

	if err == nil {
		err = pn.writeConfig()
	}

	if err != nil {
		CR.CtxStop("seed failed", nil)

		// TODO: clean up
	}

	return err
}

func (pn *Pnode) fetchMemberSession(ctx context.Context) (*repo.MemberSession, error) {
	session, err := pn.activeSessions.FetchSession(ctx)
	if err != nil {
		return nil, err
	}

	ms, _ := session.Cookie.(*repo.MemberSession)
	if ms == nil {
		return nil, plan.Errorf(nil, plan.AssertFailed, "internal type assertion err")
	}

	err = ms.CtxStatus()
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func (pn *Pnode) fetchRepo(inCommunityID []byte) *repo.CommunityRepo {

	child := pn.CtxGetChildByID(inCommunityID)
	if child != nil {
		return child.(*repo.CommunityRepo)
	}

	return nil
}

/*****************************************************
**
**
**
** rpc service Repo
**
**
**
**/

// SeedRepo -- see service Repo in repo.proto.
func (pn *Pnode) SeedRepo(
	ctx context.Context,
	inRepoSeed *repo.RepoSeed,
) (*plan.Status, error) {

	err := pn.seedRepo(inRepoSeed)
	if err != nil {
		return nil, err
	}

	// Set up the member sub dir and write the intital KeyTome
	// For now we can skip this b/c the KeyTime is already known to be local
	{
		// TODO
	}

	return &plan.Status{}, nil
}

// OpenMemberSession -- see service Repo in repo.proto.
func (pn *Pnode) OpenMemberSession(
	inSessReq *repo.MemberSessionReq,
	inMsgOutlet repo.Repo_OpenMemberSessionServer,
) error {

	CR := pn.fetchRepo(inSessReq.CommunityID)
	if CR == nil {
		return plan.Error(nil, plan.CommunityNotFound, "community not found")
	}

	ms, err := CR.OpenMemberSession(inSessReq, inMsgOutlet)
	if err != nil {
		return err
	}

	//
	// TODO: remove active session when the ms goes away

	// Because this a streaming call, headers and trailers won't ever arrive.
	// Instead, the ms passes it manually and so we have to add it here.
	sess := pn.activeSessions.NewSession(inMsgOutlet.Context(), ms.SessionToken)
	sess.Cookie = ms

	<-ms.CtxStopping()

	return nil
}

// OpenMsgPipe -- see service Repo in repo.proto.
func (pn *Pnode) OpenMsgPipe(inMsgInlet repo.Repo_OpenMsgPipeServer) error {
	ms, err := pn.fetchMemberSession(inMsgInlet.Context())
	if err != nil {
		return err
	}

	return ms.OpenMsgPipe(inMsgInlet)
}
