SYS_OS=`uname -s`
if [ $SYS_OS != "Linux"  ];then
    echo "not support OS: $SYS_OS"
    exit
fi

[ -z "${GOPATH}"  ] && GOPATH=${HOME}/go

FIRST_GOPATH=$(echo ${GOPATH}| cut -d':' -f 1)

PKG_MONERO=${FIRST_GOPATH}/src/github.com/newbitx/mymonero-core-go
[ ! -d "${PKG_MONERO}"  ] && go get github.com/newbitx/mymonero-core-go

LIB_MONERO=/build/lib/libmymonero.a
if [ ! -f "${PKG_MONERO}${LIB_MONERO}"  ]
then
    cd ${PKG_MONERO}
    ./bin/update_submodules
    [ ! -d build ] && mkdir build
    cd build

    if [ -f /usr/bin/cmake3  ]
    then
        source /opt/rh/devtoolset-7/enable
        cmake3 ..
    else
        cmake ..
    fi

    make -j4
    cd ../test
    go test
fi

