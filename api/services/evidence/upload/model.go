package evidence

// IPFSClient is a mock interface for the IPFS client
// It provides methods to upload files to IPFS and retrieve their CIDs (Content Identifiers).

type Service struct {
	ipfs *IPFSClient
}
