#!/usr/bin/env bash

protoc -Igrpctest grpctest/*.proto --go_out=plugins=grpc:grpctest
