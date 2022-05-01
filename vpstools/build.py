#!/usr/bin/env python
# -*- coding:utf-8 -*-

import time
import os
import platform


def build(outname, goos, goarch):
    x = time.localtime()
    y = x[3] * 60 * 60 + x[4] * 60 + x[5]
    mainver = "2.0.0.{0}.{1}".format(
        time.strftime("%y%m%d", time.localtime()), y)
    r = os.popen('go version')
    gover = r.read().strip().replace("go version ", "")
    pf = "{0}({1})".format(platform.platform(), platform.node())
    r.close()
    # if goos == "windows":
    #     if goarch == "386":
    #         outpath = "dist_x86"
    #     else:
    #         outpath = "dist_win"
    # else:
    #     outpath = "dist_linux"
    # outname = outpath + "/"+outname
    buildcmd = 'CGO_ENABLED=0 go build -ldflags="-s -w -X \'main.version={1}\' -X \'main.buildDate={2}\' -X \'main.goVersion={3}\' -X \'main.platform={4}\'" -o {0}'.format(
        outname, mainver, time.ctime(time.time()), gover, pf)
    # print(buildcmd)
    os.system(buildcmd)
    os.system("upx {0}".format(outname))


def build_service(outname, goos, goarch):
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
    os.system("upx --best dist_win/{0}".format(outname))


if __name__ == "__main__":
    x = time.localtime()
    y = x[3] * 60 * 60 + x[4] * 60 + x[5]
    nv = "2.0.0.{0}.{1}".format(time.strftime("%y%m%d", time.localtime()), y)
    ov = "\"0.0.0\""
    r = os.popen('go version')
    gv = r.read().strip().replace("go version ", "")
    pl = "{0}({1})".format(platform.platform(), platform.node())
    r.close()
    # print("=== start build windows x86 ...")
    # build("sample.exe",  "windows", "386")
    # print("\n=== start build windows x64 ...")
    # build("sample.exe", "windows", "amd64")
    # 编译成windows后台程序，不会有黑框，在进程列表中可查看到
    # print("\n=== start build windows x64 service ...")
    # build("sampled.exe", "windows", "amd64")
    print("\n=== start build linux x64 ...")
    build("vpstools", "linux", "amd64")

    # os.system("cp -f distwin/ecms*.exe ../../python/mwsc/dist/bin/")
