# Use >=1.13.1 for the ed25519 update
FROM golang:1.13.1

# Get git
RUN apt-get update \
    && apt-get -y install curl git \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Where pegnet sources will live
WORKDIR $GOPATH/src/github.com/pegnet/pegnetd

# Populate the rest of the source
COPY . .

ARG GOOS=linux
ENV GO111MODULE=on

# We take the config file from ~/.pegnetd first, then the active directory.
# So we do not need to copy the config file to anywhere

RUN go get
# place pegnetd in the path
RUN go install

ENTRYPOINT ["pegnetd"]