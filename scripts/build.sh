#!/bin/sh

set -e

DIR=$(pwd)
GIT2GO_PATH=$GOPATH/src/github.com/libgit2/git2go
GIT2GO_VENDOR_PATH=$GIT2GO_PATH/vendor/libgit2
OS=$(uname -s | awk '{print tolower($0)}')
ARCH=$(uname -m)

TestPackage() {
  package=$1
  go test -v -covermode=count -coverprofile=coverage.out $package

  if [ ! -z $COVERALLS_TOKEN ] && [ -f ./coverage.out ]; then 
    $HOME/gopath/bin/goveralls -coverprofile=coverage.out -service=travis-ci -repotoken $COVERALLS_TOKEN
  fi
}

go get github.com/libgit2/git2go || true &&
go get golang.org/x/tools/cmd/cover
go get github.com/mattn/goveralls

cd $GIT2GO_PATH &&
git checkout v26 &&
git submodule update --init || true &&

cd $GIT2GO_VENDOR_PATH &&
mkdir -p install/lib &&
mkdir -p build &&
cd build &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      -DUSE_SSH=OFF \
      -DCURL=OFF \
      .. &&

cmake --build . &&
make -j2 install &&

export PKG_CONFIG_PATH=$GIT2GO_VENDOR_PATH/build
export CGO_LDFLAGS="$(pkg-config --libs --static $GOPATH/src/github.com/libgit2/git2go/vendor/libgit2/build/libgit2.pc)"
go install -x github.com/libgit2/git2go &&

cd $DIR

VERSION=$TRAVIS_TAG
if [ -z $VERSION ]; then
  VERSION="0.0.0"
fi

OUT="mbt_${OS}_${ARCH}"

make restore

TestPackage .
TestPackage ./lib
TestPackage ./cmd

go build -o "build/${OUT}"
shasum -a 1 -p "build/${OUT}" | cut -d ' ' -f 1 > "build/${OUT}.sha1"
echo "testing the bin"
"./build/${OUT}" version

cat >build/bintray.json << EOL
{
    "package": {
        "name": "${OUT}",
        "repo": "bin",
        "subject": "buddyspike",
        "desc": "Monorepo Build Tool",
        "website_url": "https://github.com/mbtproject/mbt", "issue_tracker_url": "https://github.com/mbtproject/mbt/issues", "vcs_url": "https://github.com/buddyspike/mbt.git", "public_download_numbers": true, "public_stats": true }, "version": {
        "name": "${VERSION}",
        "gpgSign": false
    },
    "files": [ {"includePattern": "build/${OUT}", "uploadPattern": "/${OUT}"} ],
    "publish": true
}
EOL