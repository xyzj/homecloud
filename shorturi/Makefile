BINARY_NAME=shorturi
MAIN_VER=0.1.0

DIST_LINUX=_dist/${BINARY_NAME}

DATE_VER=`date '+%y%m%d.%H%M%S'`
GO_VER=`go version | cut -d \  -f 3`
BUILD_DATE=`date`
BUILD_OS=`uname -srv`
LDFLAGS="-s -w -X 'main.goVersion=${GO_VER}' -X 'main.buildDate=${BUILD_DATE}' -X 'main.platform=${BUILD_OS}' -X 'main.version=${MAIN_VER}.${DATE_VER}'"

# GOARCH for linux enable:
#	"amd64", "arm64", "mips64", "mips64le", "ppc64", "ppc64le", "riscv64", "s390x", "wasm"
#	"loong64" may need c source code
# Detail: https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63

# 编译linux 64位版本
linux:
	@echo "building linux amd64 version..."
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 CC=musl-gcc go build -o ${DIST_LINUX} -ldflags=${LDFLAGS} main.go
	@upx ${DIST_LINUX}
	@echo "done.\n"
