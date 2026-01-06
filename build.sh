#!/bin/bash

mkdir -p build

set -e

TOOLCHAIN_VER="5.3.0-r07"

if [[ ! -d ~/.anki/vicos-sdk/dist/$TOOLCHAIN_VER ]]; then
  echo "Getting toolchain version $TOOLCHAIN_VER..."
  mkdir -p ~/.anki/vicos-sdk/dist/$TOOLCHAIN_VER
  cd ~/.anki/vicos-sdk/dist/$TOOLCHAIN_VER
  wget -q --show-progress https://github.com/os-vector/wire-os-externals/releases/download/$TOOLCHAIN_VER/vicos-sdk_"$TOOLCHAIN_VER"_amd64-linux.tar.gz -O - | tar -xz
  echo "Toolchain version $TOOLCHAIN_VER has been installed!"
else
  echo "Toolchain version $TOOLCHAIN_VER is already installed!"
fi

TC="/home/$USER/.anki/vicos-sdk/dist/$TOOLCHAIN_VER/prebuilt/bin/arm-oe-linux-gnueabi-"

cd vector-gobot
GCC="${TC}clang" GPP="${TC}clang++" make vector-gobot
cd ..
cp vector-gobot/build/libvector-gobot.so build/

CC=${TC}clang \
CXX=${TC}clang++ \
CGO_CFLAGS="-I$(pwd)/vector-gobot/include" \
CGO_LDFLAGS="-L$(pwd)/build -Wl,-rpath,/anki/lib" \
GOARCH=arm \
GOARM=7 \
CGO_ENABLED=1 \
go build \
-ldflags "-s -w -r /anki/lib" \
-o build/vic-verbose
