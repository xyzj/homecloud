BINARY_NAME=hcloud
MAIN_VER=2.0.0

DIST_WINDOWS=_dist/${BINARY_NAME}.exe
DIST_LINUX=_dist/${BINARY_NAME}
DIST_ARM64=_dist/${BINARY_NAME}-arm64
DIST_MIPS64=_dist/${BINARY_NAME}-mips64

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
linux: modtidy
	@echo "building linux amd64 version..."
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 CC=musl-gcc go build -o ${DIST_LINUX} -ldflags=${LDFLAGS} main.go
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 CC=musl-gcc go build -o _dist/sender -ldflags=${LDFLAGS} sender/main.go
	@upx ${DIST_LINUX}
	@upx _dist/sender
	@echo "done.\n"

# 编译windows 64位版本
windows: modtidy
	@echo "building windows amd64 version..."
	@GOARCH=amd64 GOOS=windows CGO_ENABLED=0 CC=musl-gcc go build -o ${DIST_WINDOWS} -ldflags=${LDFLAGS} main.go
	@echo "done.\n"

# 编译arm64/aarch64架构版本
arm64: modtidy
	@echo "building linux arm64/aarch64 version..."
	@GOARCH=arm64 GOOS=linux CGO_ENABLED=0 CC=musl-gcc go build -o ${DIST_ARM64} -ldflags=${LDFLAGS} main.go
	@echo "done.\n"

# 编译arm的mips64架构版本
mips64: modtidy
	@echo "building linux mips64 version..."
	@GOARCH=mips64 GOOS=linux CGO_ENABLED=0 CC=musl-gcc go build -o ${DIST_MIPS64} -ldflags=${LDFLAGS} main.go
	@echo "done.\n"

test: linux
	@scp -p ${DIST_LINUX} cos83:/tmp/

# 更新go.mod文件
modtidy:
	@go mod tidy

# 更新所有依赖包
modupdate:
	@go get -u -v all
	@echo "done."

# 编译api文档
doc:
	@apidoc -o _docs -i apidoc -i lib -p --single
	@mv _docs/index.html _docs/apidoc-${BINARY_NAME}.html
	@scp -p _docs/*.html wlstl:/home/shares/archiving/_apidoc/luwak

# 清理编译文件
clean:
	@rm -fv _docs/*.html
	@rm -fv _dist/${BINARY_NAME}*

#计算代码行数
cloc:
	@~/go/bin/gocloc .

push:
	@git gc
	@git fsck
	@git push