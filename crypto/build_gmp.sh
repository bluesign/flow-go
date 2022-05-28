#!/bin/bash -x


# Set defaults
GMP_VERSION="6.2.1"
DESTDIR=$(pwd)/relic/build
PREFIX=${DESTDIR}/"gmp"/${GMP_VERSION}
BUILDDIR=${PREFIX}/build

if [ -z "$GMP_VERSION" ]; then
  echo "[FATAL] \$GMP_VERSION not set"
  echo "[INFO] Usage: GMP_VERSION=<version> $0"
  exit 1
fi


# Use a build directory
mkdir -p "${BUILDDIR}" || exit 1
pushd    "${BUILDDIR}" || exit 1


# Download and unpack software
TARFILE="gmp-${GMP_VERSION}.tar.lz"
if ! test -e "${TARFILE}"; then
    curl "https://gmplib.org/download/gmp/${TARFILE}" -o ./${TARFILE} || ./exit 1
fi

if ! test -e "gmp-${GMP_VERSION}"; then
    tar jxvf "${TARFILE}" || exit 1
fi

# Install
SRCDIR="$(pwd)/gmp-${GMP_VERSION}"
"${SRCDIR}/configure" --prefix="${PREFIX}" --disable-shared --disable-fft || exit 1
make install -j || exit 1


# Setup permissions
chmod -R g+r "${PREFIX}" || exit 1
find "${PREFIX}" -type d -exec chmod g+x {} \; || exit 1
