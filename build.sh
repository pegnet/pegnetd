#!/usr/bin/env bash

go install -ldflags="-X github.com/pegnet/pegnetd/config.CompiledInBuild=`git rev-parse HEAD` -X github.com/pegnet/pegnetd/config.CompiledInVersion=`git describe --tags`"
