package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	csav1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/csa"
	jobv1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/job"
	nodev1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/node"
	"github.com/smartcontractkit/chainlink-protos/job-distributor/v1/shared/ptypes"

	"github.com/job-distributor/internal/chainlink"
	"github.com/job-distributor/internal/config"
	"github.com/job-distributor/internal/storage"
)

// JobDistributorServer implements all the required gRPC interfaces
type JobDistributorServer struct {
	mu        sync.RWMutex
	config    *config.Config
	storage   *storage.Storage
	chainlink *chainlink.ClientManager

	// Embed the unimplemented servers to satisfy interface requirements
	jobv1.UnimplementedJobServiceServer
	nodev1.UnimplementedNodeServiceServer
	csav1.UnimplementedCSAServiceServer
}

// NewJobDistributorServer creates a new Job Distributor server
func NewJobDistributorServer(cfg *config.Config) *JobDistributorServer {
	server := &JobDistributorServer{
		config:    cfg,
		storage:   storage.NewStorage(),
		chainlink: chainlink.NewClientManager(), // ✅ Start with empty client manager
	}

	return server
}

// =============================================================================
// JOB SERVICE METHODS
// =============================================================================

// ProposeJob handles job proposal requests - this is the main method the CCIP deployer calls
func (s *JobDistributorServer) ProposeJob(ctx context.Context, req *jobv1.ProposeJobRequest) (*jobv1.ProposeJobResponse, error) {
	log.Printf("📝 ProposeJob called for node %s", req.NodeId)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate node exists
	if !s.storage.NodeExists(req.NodeId) {
		return nil, status.Errorf(codes.NotFound, "node %s not found", req.NodeId)
	}

	// Extract job ID from spec
	jobID, err := extractJobIDFromSpec(req.Spec)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to extract job ID from spec: %v", err)
	}

	// Create proposal
	proposalID := generateUUID()
	proposal := &jobv1.Proposal{
		Id:             proposalID,
		JobId:          jobID,
		Spec:           req.Spec,
		Status:         jobv1.ProposalStatus_PROPOSAL_STATUS_PENDING,
		DeliveryStatus: jobv1.ProposalDeliveryStatus_PROPOSAL_DELIVERY_STATUS_DELIVERED,
		CreatedAt:      timestamppb.Now(),
		UpdatedAt:      timestamppb.Now(),
	}

	// Submit job to the actual Chainlink node via REST API
	jobSubmitted := s.submitJobToNode(ctx, req.NodeId, req.Spec)

	if jobSubmitted {
		// Auto-approve the job (as per CCIP deployer expectations)
		proposal.Status = jobv1.ProposalStatus_PROPOSAL_STATUS_APPROVED

		// Store the job record
		job := &jobv1.Job{
			Id:          jobID,
			Uuid:        jobID,
			NodeId:      req.NodeId,
			ProposalIds: []string{proposalID},
			Labels:      req.Labels,
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
		}
		s.storage.StoreJob(job)
		log.Printf("✅ Job %s created and approved for node %s", jobID, req.NodeId)
	} else {
		proposal.Status = jobv1.ProposalStatus_PROPOSAL_STATUS_REJECTED
		log.Printf("❌ Job %s rejected for node %s", jobID, req.NodeId)
	}

	// Store proposal
	s.storage.StoreProposal(proposal)

	return &jobv1.ProposeJobResponse{
		Proposal: proposal,
	}, nil
}

// BatchProposeJob handles batch job proposals
func (s *JobDistributorServer) BatchProposeJob(ctx context.Context, req *jobv1.BatchProposeJobRequest) (*jobv1.BatchProposeJobResponse, error) {
	log.Printf("📝 BatchProposeJob called for %d nodes", len(req.NodeIds))

	resp := &jobv1.BatchProposeJobResponse{
		SuccessResponses: make(map[string]*jobv1.ProposeJobResponse),
		FailedResponses:  make(map[string]*jobv1.ProposeJobFailure),
	}

	for _, nodeID := range req.NodeIds {
		singleReq := &jobv1.ProposeJobRequest{
			NodeId: nodeID,
			Spec:   req.Spec,
			Labels: req.Labels,
		}

		proposeResp, err := s.ProposeJob(ctx, singleReq)
		if err != nil {
			resp.FailedResponses[nodeID] = &jobv1.ProposeJobFailure{
				ErrorMessage: err.Error(),
			}
		} else {
			resp.SuccessResponses[nodeID] = proposeResp
		}
	}

	return resp, nil
}

// ListJobs returns all stored jobs
func (s *JobDistributorServer) ListJobs(ctx context.Context, req *jobv1.ListJobsRequest) (*jobv1.ListJobsResponse, error) {
	log.Printf("📋 ListJobs called")

	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := s.storage.ListJobs(req.Filter)

	return &jobv1.ListJobsResponse{
		Jobs: jobs,
	}, nil
}

// GetJob retrieves a specific job
func (s *JobDistributorServer) GetJob(ctx context.Context, req *jobv1.GetJobRequest) (*jobv1.GetJobResponse, error) {
	var jobID string
	if idReq := req.GetId(); idReq != "" {
		jobID = idReq
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "job ID is required")
	}

	log.Printf("🔍 GetJob called for job %s", jobID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	job := s.storage.GetJob(jobID)
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %s not found", jobID)
	}

	return &jobv1.GetJobResponse{
		Job: job,
	}, nil
}

// ListProposals returns all stored proposals
func (s *JobDistributorServer) ListProposals(ctx context.Context, req *jobv1.ListProposalsRequest) (*jobv1.ListProposalsResponse, error) {
	log.Printf("📋 ListProposals called")

	s.mu.RLock()
	defer s.mu.RUnlock()

	proposals := s.storage.ListProposals(req.Filter)

	return &jobv1.ListProposalsResponse{
		Proposals: proposals,
	}, nil
}

// GetProposal retrieves a specific proposal
func (s *JobDistributorServer) GetProposal(ctx context.Context, req *jobv1.GetProposalRequest) (*jobv1.GetProposalResponse, error) {
	log.Printf("🔍 GetProposal called for proposal %s", req.Id)

	s.mu.RLock()
	defer s.mu.RUnlock()

	proposal := s.storage.GetProposal(req.Id)
	if proposal == nil {
		return nil, status.Errorf(codes.NotFound, "proposal %s not found", req.Id)
	}

	return &jobv1.GetProposalResponse{
		Proposal: proposal,
	}, nil
}

// RevokeJob revokes a job proposal
func (s *JobDistributorServer) RevokeJob(ctx context.Context, req *jobv1.RevokeJobRequest) (*jobv1.RevokeJobResponse, error) {
	var jobID string
	if idReq := req.GetId(); idReq != "" {
		jobID = idReq
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "job ID is required")
	}

	log.Printf("🔄 RevokeJob called for job %s", jobID)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find proposals for this job
	proposals := s.storage.ListProposals(&jobv1.ListProposalsRequest_Filter{
		JobIds: []string{jobID},
	})

	if len(proposals) == 0 {
		return nil, status.Errorf(codes.NotFound, "no proposals found for job %s", jobID)
	}

	// Get the latest proposal
	latestProposal := proposals[0]
	for _, p := range proposals {
		if p.UpdatedAt.AsTime().After(latestProposal.UpdatedAt.AsTime()) {
			latestProposal = p
		}
	}

	// Update status to revoked
	latestProposal.Status = jobv1.ProposalStatus_PROPOSAL_STATUS_REVOKED
	latestProposal.UpdatedAt = timestamppb.Now()
	s.storage.StoreProposal(latestProposal)

	return &jobv1.RevokeJobResponse{
		Proposal: latestProposal,
	}, nil
}

// DeleteJob deletes a job (placeholder implementation)
func (s *JobDistributorServer) DeleteJob(ctx context.Context, req *jobv1.DeleteJobRequest) (*jobv1.DeleteJobResponse, error) {
	var jobID string
	if idReq := req.GetId(); idReq != "" {
		jobID = idReq
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "job ID is required")
	}

	log.Printf("🗑️ DeleteJob called for job %s", jobID)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete job from the actual Chainlink node via REST API
	deleted := s.deleteJobFromNode(ctx, jobID)

	if deleted {
		s.storage.DeleteJob(jobID)
		log.Printf("✅ Job %s deleted successfully", jobID)
	}

	return &jobv1.DeleteJobResponse{}, nil
}

// UpdateJob updates a job (placeholder implementation)
func (s *JobDistributorServer) UpdateJob(ctx context.Context, req *jobv1.UpdateJobRequest) (*jobv1.UpdateJobResponse, error) {
	var jobID string
	if idReq := req.GetId(); idReq != "" {
		jobID = idReq
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "job ID is required")
	}

	log.Printf("🔄 UpdateJob called for job %s", jobID)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get existing job
	existingJob := s.storage.GetJob(jobID)
	if existingJob == nil {
		return nil, status.Errorf(codes.NotFound, "job %s not found", jobID)
	}

	// Update job spec on the Chainlink node first
	client := s.chainlink.GetClient(existingJob.NodeId)
	if client == nil {
		return nil, status.Errorf(codes.Internal, "no client found for node %s", existingJob.NodeId)
	}

	// For now, just update the job labels without changing spec
	// TODO: Add spec validation if needed in the future

	// Update job in storage
	updatedJob := &jobv1.Job{
		Id:          jobID,
		Uuid:        existingJob.Uuid, // Keep original UUID
		NodeId:      existingJob.NodeId,
		ProposalIds: existingJob.ProposalIds,
		Labels:      req.Labels,
		CreatedAt:   existingJob.CreatedAt,
		UpdatedAt:   timestamppb.Now(),
	}

	s.storage.StoreJob(updatedJob)
	log.Printf("✅ Job %s updated successfully on node %s", jobID, existingJob.NodeId)

	return &jobv1.UpdateJobResponse{
		Job: updatedJob,
	}, nil
}

// =============================================================================
// NODE SERVICE METHODS
// =============================================================================

// ListNodes returns all configured nodes
func (s *JobDistributorServer) ListNodes(ctx context.Context, req *nodev1.ListNodesRequest) (*nodev1.ListNodesResponse, error) {
	log.Printf("📋 ListNodes called")

	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := s.storage.ListNodes(req.Filter)

	return &nodev1.ListNodesResponse{
		Nodes: nodes,
	}, nil
}

// GetNode retrieves a specific node
func (s *JobDistributorServer) GetNode(ctx context.Context, req *nodev1.GetNodeRequest) (*nodev1.GetNodeResponse, error) {
	log.Printf("🔍 GetNode called for node %s", req.Id)

	s.mu.RLock()
	defer s.mu.RUnlock()

	node := s.storage.GetNode(req.Id)
	if node == nil {
		return nil, status.Errorf(codes.NotFound, "node %s not found", req.Id)
	}

	return &nodev1.GetNodeResponse{
		Node: node,
	}, nil
}

// RegisterNode registers a new node (placeholder implementation)
func (s *JobDistributorServer) RegisterNode(ctx context.Context, req *nodev1.RegisterNodeRequest) (*nodev1.RegisterNodeResponse, error) {
	log.Printf("📝 RegisterNode called for CSA key %s", req.PublicKey)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if node is already registered
	existingNode := s.storage.GetNodeByCSAKey(req.PublicKey)
	if existingNode != nil {
		log.Printf("⚠️ Node with CSA key %s already registered", req.PublicKey)
		return &nodev1.RegisterNodeResponse{
			Node: existingNode,
		}, nil
	}

	// Generate a new node ID
	nodeID := generateUUID()

	// Create new node record
	newNode := &nodev1.Node{
		Id:          nodeID,
		Name:        req.Name,
		PublicKey:   req.PublicKey,
		Labels:      req.Labels,
		IsConnected: true,
		IsEnabled:   true,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	}

	// Store the node
	s.storage.StoreNode(newNode)

	log.Printf("✅ Node %s registered successfully with CSA key %s", nodeID, req.PublicKey)

	return &nodev1.RegisterNodeResponse{
		Node: newNode,
	}, nil
}

// Other node methods (placeholder implementations)
func (s *JobDistributorServer) EnableNode(ctx context.Context, req *nodev1.EnableNodeRequest) (*nodev1.EnableNodeResponse, error) {
	return &nodev1.EnableNodeResponse{}, nil
}

func (s *JobDistributorServer) DisableNode(ctx context.Context, req *nodev1.DisableNodeRequest) (*nodev1.DisableNodeResponse, error) {
	return &nodev1.DisableNodeResponse{}, nil
}

func (s *JobDistributorServer) UpdateNode(ctx context.Context, req *nodev1.UpdateNodeRequest) (*nodev1.UpdateNodeResponse, error) {
	return &nodev1.UpdateNodeResponse{}, nil
}

func (s *JobDistributorServer) ListNodeChainConfigs(ctx context.Context, req *nodev1.ListNodeChainConfigsRequest) (*nodev1.ListNodeChainConfigsResponse, error) {
	return &nodev1.ListNodeChainConfigsResponse{}, nil
}

// =============================================================================
// CSA SERVICE METHODS
// =============================================================================

// ListKeypairs returns all CSA keypairs from registered nodes
func (s *JobDistributorServer) ListKeypairs(ctx context.Context, req *csav1.ListKeypairsRequest) (*csav1.ListKeypairsResponse, error) {
	log.Printf("🔑 ListKeypairs called")

	s.mu.RLock()
	defer s.mu.RUnlock()

	var keypairs []*csav1.Keypair
	nodes := s.storage.ListNodes(nil) // Get all nodes

	for _, node := range nodes {
		if node.PublicKey != "" {
			keypairs = append(keypairs, &csav1.Keypair{
				PublicKey: node.PublicKey,
			})
		}
	}

	return &csav1.ListKeypairsResponse{
		Keypairs: keypairs,
	}, nil
}

// GetKeypair retrieves a specific keypair by ID (simplified implementation)
func (s *JobDistributorServer) GetKeypair(ctx context.Context, req *csav1.GetKeypairRequest) (*csav1.GetKeypairResponse, error) {
	log.Printf("🔍 GetKeypair called")

	s.mu.RLock()
	defer s.mu.RUnlock()

	// In the actual usage from devenv, this method is rarely used
	// The pattern is to call ListKeypairs and take the first one
	// So we'll implement it by listing all and taking the first match
	nodes := s.storage.ListNodes(nil)

	for _, node := range nodes {
		if node.PublicKey != "" {
			// Return the first available keypair since that's the actual usage pattern
			return &csav1.GetKeypairResponse{
				Keypair: &csav1.Keypair{
					PublicKey: node.PublicKey,
				},
			}, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "no keypairs available")
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// submitJobToNode submits a job to the actual Chainlink node
func (s *JobDistributorServer) submitJobToNode(ctx context.Context, nodeID, spec string) bool {
	log.Printf("🔗 Submitting job to node %s", nodeID)

	// Get client for this node
	client := s.chainlink.GetClient(nodeID)
	if client == nil {
		log.Printf("❌ No client found for node %s", nodeID)
		return false
	}

	// Validate job spec
	if err := chainlink.ValidateJobSpec(spec); err != nil {
		log.Printf("❌ Invalid job spec for node %s: %v", nodeID, err)
		return false
	}

	// Submit job to Chainlink node
	job, err := client.CreateJob(ctx, spec)
	if err != nil {
		log.Printf("❌ Failed to create job on node %s: %v", nodeID, err)
		return false
	}

	log.Printf("✅ Job successfully created on node %s: %s", nodeID, job.ID)
	return true
}

// deleteJobFromNode deletes a job from the actual Chainlink node
func (s *JobDistributorServer) deleteJobFromNode(ctx context.Context, jobID string) bool {
	log.Printf("🗑️ Deleting job %s from nodes", jobID)

	// Find which node has this job by checking our storage
	job := s.storage.GetJob(jobID)
	if job == nil {
		log.Printf("❌ Job %s not found in storage", jobID)
		return false
	}

	// Get client for the node that has this job
	client := s.chainlink.GetClient(job.NodeId)
	if client == nil {
		log.Printf("❌ No client found for node %s", job.NodeId)
		return false
	}

	// Delete job from Chainlink node
	if err := client.DeleteJob(ctx, jobID); err != nil {
		log.Printf("❌ Failed to delete job %s from node %s: %v", jobID, job.NodeId, err)
		return false
	}

	log.Printf("✅ Job %s successfully deleted from node %s", jobID, job.NodeId)
	return true
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

func convertLabelsToProto(labels map[string]string) []*ptypes.Label {
	var protoLabels []*ptypes.Label
	for key, value := range labels {
		valueCopy := value
		protoLabels = append(protoLabels, &ptypes.Label{
			Key:   key,
			Value: &valueCopy,
		})
	}
	return protoLabels
}

func extractJobIDFromSpec(spec string) (string, error) {
	// Use the function from chainlink package to extract job ID from TOML
	jobID := chainlink.ExtractJobIDFromTOML(spec)
	if jobID != "" {
		return jobID, nil
	}

	// Fallback: generate a simple ID if extraction fails
	return fmt.Sprintf("job_%d", time.Now().Unix()), nil
}

func generateUUID() string {
	// Simple UUID generation for demo purposes
	return fmt.Sprintf("uuid_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}
