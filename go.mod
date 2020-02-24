module github.com/sipsma/bincastle

go 1.12

replace github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace github.com/moby/buildkit => github.com/sipsma/buildkit v0.6.1-0.20200210232759-2a83edf4f03c

replace github.com/godbus/dbus => github.com/godbus/dbus v0.0.0-20181101234600-2ff6f7ffd60f

// TODO if I don't have these lines, "go build" downgrades runc to rc8 in this go.mod file everytime (why???!!!)
replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc10

replace github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v0.1.2-0.20190207185410-29686dbc5559

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/checkpoint-restore/go-criu v0.0.0-20191125063657-fcdcd07065c5 // indirect
	github.com/cilium/ebpf v0.0.0-20200224172853-0b019ed01187 // indirect
	github.com/containerd/console v0.0.0-20191219165238-8375c3424e4d
	github.com/containerd/containerd v1.4.0-0.20191014053712-acdcf13d5eaf
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gofrs/flock v0.7.1
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/moby/buildkit v0.0.0-00010101000000-000000000000
	github.com/mrunalp/fileutils v0.0.0-20171103030105-7d4729fb3618 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc8.0.20190621203724-f4982d86f7fd
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/opencontainers/selinux v1.3.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/seccomp/libseccomp-golang v0.9.1 // indirect
	github.com/sipsma/containerd v1.2.6
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.1.0 // indirect
	go.etcd.io/bbolt v1.3.3
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae
	google.golang.org/grpc v1.27.1
)
