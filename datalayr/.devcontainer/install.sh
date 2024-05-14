# Install foundry
curl -L https://foundry.paradigm.xyz | bash
~/.foundry/bin/foundryup 

# Install go dependencies
go install github.com/onsi/ginkgo/v2/ginkgo@v2.2.0
go install github.com/mikefarah/yq/v4@latest
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Setup environment variables
./setup.sh