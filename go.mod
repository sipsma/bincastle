module github.com/sipsma/bincastle

go 1.14

replace github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace github.com/godbus/dbus => github.com/godbus/dbus v0.0.0-20181101234600-2ff6f7ffd60f

// TODO if I don't have these lines, "go build" downgrades runc to rc8 in this go.mod file everytime (why???!!!)
replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc10

replace github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v0.1.2-0.20190207185410-29686dbc5559

replace github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200512144102-f13ba8f2f2fd

replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20200227233006-38f52c9fec82

replace github.com/checkpoint-restore/go-criu => github.com/checkpoint-restore/go-criu v0.0.0-20181120144056-17b0214f6c48

require (
	github.com/checkpoint-restore/go-criu v4.0.0+incompatible // indirect
	github.com/containerd/console v1.0.0
	github.com/containerd/containerd v1.4.0-0
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b
	github.com/creack/pty v1.1.10
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/gofrs/flock v0.7.1
	github.com/hashicorp/go-multierror v1.0.0
	github.com/moby/buildkit v0.7.1-0.20200717034530-97ca82f4c422
	github.com/mrunalp/fileutils v0.0.0-20200520151820-abd8a0e76976 // indirect
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc9.0.20200221051241-688cf6d43cc4
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/seccomp/libseccomp-golang v0.9.1 // indirect
	github.com/sipsma/bincastle-distro v0.0.0-20200720005411-7f80e786b36b
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.3
	go.etcd.io/bbolt v1.3.4
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys v0.0.0-20200321134203-328b4cd54aae
	google.golang.org/grpc v1.28.0
)
