module github.com/sipsma/bincastle

go 1.13

replace github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace github.com/godbus/dbus => github.com/godbus/dbus v0.0.0-20181101234600-2ff6f7ffd60f

// TODO if I don't have these lines, "go build" downgrades runc to rc8 in this go.mod file everytime (why???!!!)
replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc10

replace github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v0.1.2-0.20190207185410-29686dbc5559

replace github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200227195959-4d242818bf55

replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20200227233006-38f52c9fec82

require (
	github.com/checkpoint-restore/go-criu v0.0.0-20191125063657-fcdcd07065c5 // indirect
	github.com/cilium/ebpf v0.0.0-20200319110858-a7172c01168f // indirect
	github.com/containerd/cgroups v0.0.0-20200308110149-6c3dec43a103 // indirect
	github.com/containerd/console v0.0.0-20191219165238-8375c3424e4d
	github.com/containerd/containerd v1.4.0-0
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/cli v0.0.0-20200320120634-22acbbcc4b3f // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/gofrs/flock v0.7.1
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/moby/buildkit v0.7.0-rc1
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mrunalp/fileutils v0.0.0-20171103030105-7d4729fb3618 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc9.0.20200221051241-688cf6d43cc4
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/opencontainers/selinux v1.4.0 // indirect
	github.com/opentracing-contrib/go-stdlib v0.0.0-20190519235532-cf7a6c988dc9 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/seccomp/libseccomp-golang v0.9.1 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.22.3
	github.com/vishvananda/netlink v1.1.0 // indirect
	go.etcd.io/bbolt v1.3.4
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/crypto v0.0.0-20200320181102-891825fb96df // indirect
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys v0.0.0-20200321134203-328b4cd54aae
	google.golang.org/genproto v0.0.0-20200319113533-08878b785e9c // indirect
	google.golang.org/grpc v1.28.0
)
