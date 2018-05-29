#!/bin/bash

# Helper script for building v8 from source

# die on errors
set -Eeuo pipefail

# return to current directory on exit
V8_START_DIR="$(pwd)"
function v8_exit {
  cd ${V8_START_DIR}
}
trap v8_exit EXIT

# print out a status header
function _v8_header {
  echo 1>&2
  echo "##############################################" 1>&2
  echo "# ${@}" 1>&2
  echo "##############################################" 1>&2
  echo 1>&2
}

# make sure the tools we need exist
function _validate_env {
  if ! (git --version | grep 'git version' > /dev/null); then
    echo "git not found, aborting" 1>&2
    exit 1
  fi

  if ! (ninja --version | grep '^1.' > /dev/null); then
    echo "ninja-build not found, aborting" 1>&2
    exit 1
  fi
}

# entrypoint function
function _v8_run {
  _validate_env

  # some configuration variables
  local _VER="${V8_VERSION:-6.7.288.42}"
  local _B_DIR="$(pwd)/${V8_BUILD_DIR:-build-v8}"

  # working directory
  mkdir -p "${_B_DIR}"
  cd "${_B_DIR}"

  # chromium depot tools
  _v8_header "Check out Chromium Depot Tools"
  if [ ! -d "depot_tools" ]; then
    git clone https://chromium.googlesource.com/chromium/tools/depot_tools.git
  else
    cd "depot_tools"
    git reset --hard
    git checkout master
    git pull
  fi
  cd "${_B_DIR}"
  local _D_DIR="$(pwd)/depot_tools"
  export PATH="${_D_DIR}:${PATH}"

  # initial checkout
  if [ ! -d "v8" ]; then
    _v8_header "Initial Checkout"
    fetch v8
  fi

  # finalize checkout
  _v8_header "Finalize Checkout"
  cd v8
  git checkout "${_VER}"
  gclient sync

  # build v8
  _v8_header "Build V8"
  gn gen out.gn/golib --args=$'
    target_cpu = "x64"
    is_debug = false

    symbol_level = 0
    strip_debug_info = true
    v8_experimental_extra_library_files = []
    v8_extra_library_files = []

    v8_static_library = true
    is_component_build = false
    use_custom_libcxx = false
    use_custom_libcxx_for_host = false

    is_desktop_linux = false
    v8_enable_i18n_support = false
    v8_use_external_startup_data = false
    v8_enable_gdbjit = false'
  ninja -C out.gn/golib \
    v8_libbase v8_libplatform v8_base v8_nosnapshot \
    v8_libsampler v8_init v8_initializers
}

# entrypoint
_v8_run "$@"
