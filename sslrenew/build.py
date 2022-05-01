#!/usr/bin/env python
# -*- coding:utf-8 -*-

import time
import os
import platform


def build(outname, goos, goarch, enableups):
    x = time.localtime()
    y = x[3] * 60 * 60 + x[4] * 60 + x[5]
    mainver = "1.0.0.{0}.{1}".format(
        time.strftime("%y%m%d", time.localtime()), y)
    r = os.popen('go version')
    gover = r.read().strip().replace("go version ", "")
    pf = "{0}({1})".format(platform.platform(), platform.node())
    r.close()
    if goos == "windows":
        if goarch == "386":
            outpath = "dist_x86"
        else:
            outpath = "dist_win"
    else:
        outpath = "dist_linux"
    outname = outpath + "/" + outname
    buildcmd = 'CGO_ENABLED=0 GOOS={5} GOARCH={6} go build -ldflags="-s -w -X main.version={1} -X \'main.buildDate={2}\' -X \'main.goVersion={3}\' -X \'main.platform={4}\'" -o {0} main.go'.format(
        outname, mainver, time.ctime(time.time()), gover, pf, goos, goarch)
    # print(buildcmd)
    os.system(buildcmd)
    if enableups:
        os.system("upx {0}".format(outname))


def build_service(outname, goos, goarch, enableups):
    x = time.localtime()
    y = x[3] * 60 * 60 + x[4] * 60 + x[5]
    mainver = "1.0.0.{0}.{1}".format(
        time.strftime("%y%m%d", time.localtime()), y)
    r = os.popen('go version')
    gover = r.read().strip().replace("go version ", "")
    pf = "{0}({1})".format(platform.platform(), platform.node())
    r.close()
    if goos == "windows":
        if goarch == "386":
            outpath = "dist_x86"
        else:
            outpath = "dist_win"
    else:
        outpath = "dist_linux"
    outname = outpath + "/" + outname
    buildcmd = 'GOOS={5} GOARCH={6} go build -ldflags="-s -w -H windowsgui -X main.version={1} -X \'main.buildDate={2}\' -X \'main.goVersion={3}\' -X \'main.platform={4}\'" -o {0} main.go'.format(
        outname, mainver, time.ctime(time.time()), gover, pf, goos, goarch)
    # print(buildcmd)
    os.system(buildcmd)
    if enableups:
        os.system("upx dist_win/{0}".format(outname))


if __name__ == "__main__":
    x = time.localtime()
    y = x[3] * 60 * 60 + x[4] * 60 + x[5]
    nv = "1.0.0.{0}.{1}".format(time.strftime("%y%m%d", time.localtime()), y)
    ov = "\"0.0.0\""
    r = os.popen('go version')
    gv = r.read().strip().replace("go version ", "")
    pl = "{0}({1})".format(platform.platform(), platform.node())
    r.close()
    # print("=== start build windows x86 ...")
    # build("backend_x86.exe",  "windows", "386", False)
    # print("\n=== start build windows x64 ...")
    # build("backend.exe", "windows", "amd64", False)
    print("\n=== start build windows x64 service ...")
    build("sslrenew.exe", "windows", "amd64", False)
    print("\n=== start build linux x64 ...")
    build("sslrenew", "linux", "amd64", True)

    # os.system("cp -f distwin/ecms*.exe ../../python/mwsc/dist/bin/")
