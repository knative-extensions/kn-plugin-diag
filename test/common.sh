# Copyright 2020 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

source $(dirname $0)/../vendor/knative.dev/hack/e2e-tests.sh

function cluster_setup() {
  header "Installing client"
  local kn_build=$(mktemp -d)
  local failed=""
  pushd "$kn_build"
  git clone https://github.com/knative/client . || failed="Cannot clone kn githup repo"
  hack/build.sh -f || failed="error while builing kn"
  cp kn /usr/local/bin/kn || failed="can't copy kn to /usr/local/bin"
  chmod a+x /usr/local/bin/kn || failed="can't chmod kn"
  if [ -n "$failed" ]; then
     echo "ERROR: $failed"
     exit 1
  fi
  popd
  rm -rf "$kn_build"

  header "Building plugin"
  ${REPO_ROOT_DIR}/hack/build.sh -f || return 1
}

# Copied from knative/serving setup script:
# https://github.com/knative/serving/blob/main/test/e2e-networking-library.sh#L17
function install_istio() {
  if [[ -z "${ISTIO_VERSION:-}" ]]; then
    readonly ISTIO_VERSION="stable"
  fi
  header "Installing Istio ${ISTIO_VERSION}"
# TODO: fix after 1.0 release
#  LATEST_NET_ISTIO_RELEASE_VERSION=$(
#  curl -L --silent "https://api.github.com/repos/knative/net-istio/releases" | grep '"tag_name"' \
#    | cut -f2 -d: | sed "s/[^v0-9.]//g" | sort | tail -n1)
  LATEST_NET_ISTIO_RELEASE_VERSION="knative-v1.0.0"

  # And checkout the setup script based on that release
  local NET_ISTIO_DIR=$(mktemp -d)
  (
    cd $NET_ISTIO_DIR \
      && git init \
      && git remote add origin https://github.com/knative-sandbox/net-istio.git \
      && git fetch --depth 1 origin $LATEST_NET_ISTIO_RELEASE_VERSION \
      && git checkout FETCH_HEAD
  )

  if [[ -z "${ISTIO_PROFILE:-}" ]]; then
    readonly ISTIO_PROFILE="istio-ci-no-mesh.yaml"
  fi

  if [[ -n "${CLUSTER_DOMAIN:-}" ]]; then
    sed -ie "s/cluster\.local/${CLUSTER_DOMAIN}/g" ${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/${ISTIO_PROFILE}
  fi

  echo ">> Installing Istio"
  echo "Istio version: ${ISTIO_VERSION}"
  echo "Istio profile: ${ISTIO_PROFILE}"
  ${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/install-istio.sh ${ISTIO_PROFILE}

}

function knative_setup() {
  install_istio

  header "Installing Knative Serving"
  # Defined by knative/hack/library.sh
  kubectl apply --filename ${KNATIVE_SERVING_RELEASE_CRDS}
  kubectl apply --filename ${KNATIVE_SERVING_RELEASE_CORE}
  kubectl apply --filename ${KNATIVE_NET_ISTIO_RELEASE}
  wait_until_pods_running knative-serving || return 1
}
