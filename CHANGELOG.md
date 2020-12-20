# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.1] - 2020-12-13

- add option to use file-system session store instead of cookies(default) by setting the environmental variable GRAPHIK_PLAYGROUND_SESSION_STORE=file-system

## [0.6.0] - 2020-12-14

- refactor authorizers into either request or response authorizers
- authorizers are evaluated completely within gRPC middleware
- every byte that passes through the gRPC server may be targetted for authorization

## [0.7.0] - 2020-12-14
- add custom session store to resolve issues caused by gorilla/sessions
- add flags to require request/response authorizers

## [0.8.0] - 2020-12-17
- add high availability & horizontal scaleability via Raft consensus protocol
- automatically redirect mutations to raft leader

## [0.8.1] - 2020-12-17
- add raft cluster secret so only nodes with secret may join

## [0.8.2] - 2020-12-18
- add client side options for prometheus metrics / validation / logging

## [0.9.0] - 2020-12-19
- move raft pkg to https://github.com/graphikDB/raft
- move generic pkg to https://github.com/graphikDB/generic
- move raft related methods to RaftService
- remove raft related methods from graphQL API
- move main.go to cmd/graphik/main.go