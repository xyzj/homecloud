#!/bin/bash

# CGO_ENABLED=0 CC=x86_64-w64-mingw32-gcc fyne package --appID=serlreader.xy --os=windows --icon=favicon.png --release --name=settingtools.exe 

CGO_ENABLED=0 fyne package --os=linux --icon=favicon.png --release --name=coder && \
tar xvf coder.tar.xz usr/local/bin/coder --strip-components 3 

#CGO_ENABLED=0 CGO_CFLAGS="-g -O2 -Wno-return-local-addr" fyne package --os=android --appID=com.xyzjdays.coder --name=coder --icon=favicon.png
# fyne-cross android -app-id=xyz.xyzjdays.coder