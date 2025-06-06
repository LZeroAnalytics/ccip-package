package storage

import (
	"sync"

	jobv1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/job"
	nodev1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/node"
)

// Storage provides in-memory storage for Job Distributor data
type Storage struct {
	mu        sync.RWMutex
	jobs      map[string]*jobv1.Job
	proposals map[string]*jobv1.Proposal
	nodes     map[string]*nodev1.Node
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{
		jobs:      make(map[string]*jobv1.Job),
		proposals: make(map[string]*jobv1.Proposal),
		nodes:     make(map[string]*nodev1.Node),
	}
}

// =============================================================================
// JOB STORAGE
// =============================================================================

// StoreJob stores a job
func (s *Storage) StoreJob(job *jobv1.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.Id] = job
}

// GetJob retrieves a job by ID
func (s *Storage) GetJob(id string) *jobv1.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jobs[id]
}

// ListJobs returns all jobs matching the filter
func (s *Storage) ListJobs(filter *jobv1.ListJobsRequest_Filter) []*jobv1.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*jobv1.Job
	for _, job := range s.jobs {
		if s.jobMatchesFilter(job, filter) {
			result = append(result, job)
		}
	}
	return result
}

// DeleteJob removes a job by ID
func (s *Storage) DeleteJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
}

// jobMatchesFilter checks if a job matches the given filter
func (s *Storage) jobMatchesFilter(job *jobv1.Job, filter *jobv1.ListJobsRequest_Filter) bool {
	if filter == nil {
		return true
	}

	// Filter by node IDs
	if len(filter.NodeIds) > 0 {
		found := false
		for _, nodeID := range filter.NodeIds {
			if job.NodeId == nodeID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by job IDs
	if len(filter.Ids) > 0 {
		found := false
		for _, id := range filter.Ids {
			if job.Id == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// =============================================================================
// PROPOSAL STORAGE
// =============================================================================

// StoreProposal stores a proposal
func (s *Storage) StoreProposal(proposal *jobv1.Proposal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proposals[proposal.Id] = proposal
}

// GetProposal retrieves a proposal by ID
func (s *Storage) GetProposal(id string) *jobv1.Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.proposals[id]
}

// ListProposals returns all proposals matching the filter
func (s *Storage) ListProposals(filter *jobv1.ListProposalsRequest_Filter) []*jobv1.Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*jobv1.Proposal
	for _, proposal := range s.proposals {
		if s.proposalMatchesFilter(proposal, filter) {
			result = append(result, proposal)
		}
	}
	return result
}

// proposalMatchesFilter checks if a proposal matches the given filter
func (s *Storage) proposalMatchesFilter(proposal *jobv1.Proposal, filter *jobv1.ListProposalsRequest_Filter) bool {
	if filter == nil {
		return true
	}

	// Filter by job IDs
	if len(filter.JobIds) > 0 {
		found := false
		for _, jobID := range filter.JobIds {
			if proposal.JobId == jobID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// =============================================================================
// NODE STORAGE
// =============================================================================

// StoreNode stores a node
func (s *Storage) StoreNode(node *nodev1.Node) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[node.Id] = node
}

// GetNode retrieves a node by ID
func (s *Storage) GetNode(id string) *nodev1.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[id]
}

// GetNodeByCSAKey retrieves a node by CSA public key
func (s *Storage) GetNodeByCSAKey(publicKey string) *nodev1.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, node := range s.nodes {
		if node.PublicKey == publicKey {
			return node
		}
	}
	return nil
}

// NodeExists checks if a node exists
func (s *Storage) NodeExists(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.nodes[id]
	return exists
}

// ListNodes returns all nodes matching the filter
func (s *Storage) ListNodes(filter *nodev1.ListNodesRequest_Filter) []*nodev1.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*nodev1.Node
	for _, node := range s.nodes {
		if s.nodeMatchesFilter(node, filter) {
			result = append(result, node)
		}
	}
	return result
}

// nodeMatchesFilter checks if a node matches the given filter
func (s *Storage) nodeMatchesFilter(node *nodev1.Node, filter *nodev1.ListNodesRequest_Filter) bool {
	if filter == nil {
		return true
	}

	// Filter by node IDs
	if len(filter.Ids) > 0 {
		found := false
		for _, id := range filter.Ids {
			if node.Id == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
