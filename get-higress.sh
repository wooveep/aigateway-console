#!/usr/bin/env bash

#  Copyright (c) 2023 Alibaba Group Holding Ltd.

#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at

#       http:www.apache.org/licenses/LICENSE-2.0

#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

export VERSION

MODE="install"
LOCAL_CONSOLE_IMAGE_DEFAULT="aigateway-console:0.0.1"
CUSTOM_CONSOLE_IMAGE=""
CUSTOM_CONSOLE_TAG=""
CUSTOM_GATEWAY_IMAGE=""
CUSTOM_GATEWAY_TAG=""
CUSTOM_CONTROLLER_IMAGE=""
CUSTOM_CONTROLLER_TAG=""
CUSTOM_PILOT_IMAGE=""
CUSTOM_PILOT_TAG=""
CUSTOM_API_SERVER_IMAGE=""
CUSTOM_API_SERVER_TAG=""
CUSTOM_NACOS_IMAGE=""
CUSTOM_NACOS_TAG=""
CUSTOM_GRAFANA_IMAGE=""
CUSTOM_GRAFANA_TAG=""
CUSTOM_PROMETHEUS_IMAGE=""
CUSTOM_PROMETHEUS_TAG=""
CUSTOM_PROMTAIL_IMAGE=""
CUSTOM_PROMTAIL_TAG=""
CUSTOM_LOKI_IMAGE=""
CUSTOM_LOKI_TAG=""

HAS_CURL="$(type "curl" &> /dev/null && echo true || echo false)"
HAS_WGET="$(type "wget" &> /dev/null && echo true || echo false)"
HAS_DOCKER="$(type "docker" &> /dev/null && echo true || echo false)"

parseArgs() {
  CONFIG_ARGS=()

  DESTINATION=""
  MODE="install"
  CUSTOM_CONSOLE_IMAGE=""
  CUSTOM_CONSOLE_TAG=""
  CUSTOM_GATEWAY_IMAGE=""
  CUSTOM_GATEWAY_TAG=""
  CUSTOM_CONTROLLER_IMAGE=""
  CUSTOM_CONTROLLER_TAG=""
  CUSTOM_PILOT_IMAGE=""
  CUSTOM_PILOT_TAG=""
  CUSTOM_API_SERVER_IMAGE=""
  CUSTOM_API_SERVER_TAG=""
  CUSTOM_NACOS_IMAGE=""
  CUSTOM_NACOS_TAG=""
  CUSTOM_GRAFANA_IMAGE=""
  CUSTOM_GRAFANA_TAG=""
  CUSTOM_PROMETHEUS_IMAGE=""
  CUSTOM_PROMETHEUS_TAG=""
  CUSTOM_PROMTAIL_IMAGE=""
  CUSTOM_PROMTAIL_TAG=""
  CUSTOM_LOKI_IMAGE=""
  CUSTOM_LOKI_TAG=""

  if [[ $1 != "-"* ]]; then
    DESTINATION="$1"
    shift
  fi

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        outputUsage
        exit 0
        ;;
      -u|--update)
        MODE="update"
        shift
        ;;
      --local-console-image)
        CUSTOM_CONSOLE_IMAGE="$LOCAL_CONSOLE_IMAGE_DEFAULT"
        shift
        ;;
      --console-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --console-image." && exit 1
        fi
        CUSTOM_CONSOLE_IMAGE="$2"
        shift
        shift
        ;;
      --console-image=*)
        CUSTOM_CONSOLE_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_CONSOLE_IMAGE" ]; then
          echo "Missing value for --console-image." && exit 1
        fi
        shift
        ;;
      --console-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --console-tag." && exit 1
        fi
        CUSTOM_CONSOLE_TAG="$2"
        shift
        shift
        ;;
      --console-tag=*)
        CUSTOM_CONSOLE_TAG="${1#*=}"
        if [ -z "$CUSTOM_CONSOLE_TAG" ]; then
          echo "Missing value for --console-tag." && exit 1
        fi
        shift
        ;;
      --gateway-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --gateway-image." && exit 1
        fi
        CUSTOM_GATEWAY_IMAGE="$2"
        shift
        shift
        ;;
      --gateway-image=*)
        CUSTOM_GATEWAY_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_GATEWAY_IMAGE" ]; then
          echo "Missing value for --gateway-image." && exit 1
        fi
        shift
        ;;
      --gateway-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --gateway-tag." && exit 1
        fi
        CUSTOM_GATEWAY_TAG="$2"
        shift
        shift
        ;;
      --gateway-tag=*)
        CUSTOM_GATEWAY_TAG="${1#*=}"
        if [ -z "$CUSTOM_GATEWAY_TAG" ]; then
          echo "Missing value for --gateway-tag." && exit 1
        fi
        shift
        ;;
      --controller-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --controller-image." && exit 1
        fi
        CUSTOM_CONTROLLER_IMAGE="$2"
        shift
        shift
        ;;
      --controller-image=*)
        CUSTOM_CONTROLLER_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_CONTROLLER_IMAGE" ]; then
          echo "Missing value for --controller-image." && exit 1
        fi
        shift
        ;;
      --controller-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --controller-tag." && exit 1
        fi
        CUSTOM_CONTROLLER_TAG="$2"
        shift
        shift
        ;;
      --controller-tag=*)
        CUSTOM_CONTROLLER_TAG="${1#*=}"
        if [ -z "$CUSTOM_CONTROLLER_TAG" ]; then
          echo "Missing value for --controller-tag." && exit 1
        fi
        shift
        ;;
      --pilot-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --pilot-image." && exit 1
        fi
        CUSTOM_PILOT_IMAGE="$2"
        shift
        shift
        ;;
      --pilot-image=*)
        CUSTOM_PILOT_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_PILOT_IMAGE" ]; then
          echo "Missing value for --pilot-image." && exit 1
        fi
        shift
        ;;
      --pilot-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --pilot-tag." && exit 1
        fi
        CUSTOM_PILOT_TAG="$2"
        shift
        shift
        ;;
      --pilot-tag=*)
        CUSTOM_PILOT_TAG="${1#*=}"
        if [ -z "$CUSTOM_PILOT_TAG" ]; then
          echo "Missing value for --pilot-tag." && exit 1
        fi
        shift
        ;;
      --api-server-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --api-server-image." && exit 1
        fi
        CUSTOM_API_SERVER_IMAGE="$2"
        shift
        shift
        ;;
      --api-server-image=*)
        CUSTOM_API_SERVER_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_API_SERVER_IMAGE" ]; then
          echo "Missing value for --api-server-image." && exit 1
        fi
        shift
        ;;
      --api-server-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --api-server-tag." && exit 1
        fi
        CUSTOM_API_SERVER_TAG="$2"
        shift
        shift
        ;;
      --api-server-tag=*)
        CUSTOM_API_SERVER_TAG="${1#*=}"
        if [ -z "$CUSTOM_API_SERVER_TAG" ]; then
          echo "Missing value for --api-server-tag." && exit 1
        fi
        shift
        ;;
      --nacos-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --nacos-image." && exit 1
        fi
        CUSTOM_NACOS_IMAGE="$2"
        shift
        shift
        ;;
      --nacos-image=*)
        CUSTOM_NACOS_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_NACOS_IMAGE" ]; then
          echo "Missing value for --nacos-image." && exit 1
        fi
        shift
        ;;
      --nacos-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --nacos-tag." && exit 1
        fi
        CUSTOM_NACOS_TAG="$2"
        shift
        shift
        ;;
      --nacos-tag=*)
        CUSTOM_NACOS_TAG="${1#*=}"
        if [ -z "$CUSTOM_NACOS_TAG" ]; then
          echo "Missing value for --nacos-tag." && exit 1
        fi
        shift
        ;;
      --grafana-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --grafana-image." && exit 1
        fi
        CUSTOM_GRAFANA_IMAGE="$2"
        shift
        shift
        ;;
      --grafana-image=*)
        CUSTOM_GRAFANA_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_GRAFANA_IMAGE" ]; then
          echo "Missing value for --grafana-image." && exit 1
        fi
        shift
        ;;
      --grafana-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --grafana-tag." && exit 1
        fi
        CUSTOM_GRAFANA_TAG="$2"
        shift
        shift
        ;;
      --grafana-tag=*)
        CUSTOM_GRAFANA_TAG="${1#*=}"
        if [ -z "$CUSTOM_GRAFANA_TAG" ]; then
          echo "Missing value for --grafana-tag." && exit 1
        fi
        shift
        ;;
      --prometheus-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --prometheus-image." && exit 1
        fi
        CUSTOM_PROMETHEUS_IMAGE="$2"
        shift
        shift
        ;;
      --prometheus-image=*)
        CUSTOM_PROMETHEUS_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_PROMETHEUS_IMAGE" ]; then
          echo "Missing value for --prometheus-image." && exit 1
        fi
        shift
        ;;
      --prometheus-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --prometheus-tag." && exit 1
        fi
        CUSTOM_PROMETHEUS_TAG="$2"
        shift
        shift
        ;;
      --prometheus-tag=*)
        CUSTOM_PROMETHEUS_TAG="${1#*=}"
        if [ -z "$CUSTOM_PROMETHEUS_TAG" ]; then
          echo "Missing value for --prometheus-tag." && exit 1
        fi
        shift
        ;;
      --promtail-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --promtail-image." && exit 1
        fi
        CUSTOM_PROMTAIL_IMAGE="$2"
        shift
        shift
        ;;
      --promtail-image=*)
        CUSTOM_PROMTAIL_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_PROMTAIL_IMAGE" ]; then
          echo "Missing value for --promtail-image." && exit 1
        fi
        shift
        ;;
      --promtail-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --promtail-tag." && exit 1
        fi
        CUSTOM_PROMTAIL_TAG="$2"
        shift
        shift
        ;;
      --promtail-tag=*)
        CUSTOM_PROMTAIL_TAG="${1#*=}"
        if [ -z "$CUSTOM_PROMTAIL_TAG" ]; then
          echo "Missing value for --promtail-tag." && exit 1
        fi
        shift
        ;;
      --loki-image)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --loki-image." && exit 1
        fi
        CUSTOM_LOKI_IMAGE="$2"
        shift
        shift
        ;;
      --loki-image=*)
        CUSTOM_LOKI_IMAGE="${1#*=}"
        if [ -z "$CUSTOM_LOKI_IMAGE" ]; then
          echo "Missing value for --loki-image." && exit 1
        fi
        shift
        ;;
      --loki-tag)
        if [ -z "$2" ] || [[ $2 == "-"* ]]; then
          echo "Missing value for --loki-tag." && exit 1
        fi
        CUSTOM_LOKI_TAG="$2"
        shift
        shift
        ;;
      --loki-tag=*)
        CUSTOM_LOKI_TAG="${1#*=}"
        if [ -z "$CUSTOM_LOKI_TAG" ]; then
          echo "Missing value for --loki-tag." && exit 1
        fi
        shift
        ;;
      *)
        CONFIG_ARGS+=("$1")
        shift
        ;;
    esac
  done

  if [ -n "$CUSTOM_CONSOLE_IMAGE" ] && [ -n "$CUSTOM_CONSOLE_TAG" ]; then
    echo "Only one of --console-image/--local-console-image and --console-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_GATEWAY_IMAGE" ] && [ -n "$CUSTOM_GATEWAY_TAG" ]; then
    echo "Only one of --gateway-image and --gateway-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_CONTROLLER_IMAGE" ] && [ -n "$CUSTOM_CONTROLLER_TAG" ]; then
    echo "Only one of --controller-image and --controller-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_PILOT_IMAGE" ] && [ -n "$CUSTOM_PILOT_TAG" ]; then
    echo "Only one of --pilot-image and --pilot-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_API_SERVER_IMAGE" ] && [ -n "$CUSTOM_API_SERVER_TAG" ]; then
    echo "Only one of --api-server-image and --api-server-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_NACOS_IMAGE" ] && [ -n "$CUSTOM_NACOS_TAG" ]; then
    echo "Only one of --nacos-image and --nacos-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_GRAFANA_IMAGE" ] && [ -n "$CUSTOM_GRAFANA_TAG" ]; then
    echo "Only one of --grafana-image and --grafana-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_PROMETHEUS_IMAGE" ] && [ -n "$CUSTOM_PROMETHEUS_TAG" ]; then
    echo "Only one of --prometheus-image and --prometheus-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_PROMTAIL_IMAGE" ] && [ -n "$CUSTOM_PROMTAIL_TAG" ]; then
    echo "Only one of --promtail-image and --promtail-tag can be provided." && exit 1
  fi
  if [ -n "$CUSTOM_LOKI_IMAGE" ] && [ -n "$CUSTOM_LOKI_TAG" ]; then
    echo "Only one of --loki-image and --loki-tag can be provided." && exit 1
  fi

  DESTINATION=${DESTINATION:-$PWD/higress}
}

validateArgs() {
  if [ -d "$DESTINATION" ]; then
    if [ -e "${DESTINATION}/compose/.configured" -a "$MODE" != "update" ]; then
      echo "Higress is already installed in the target folder \"$DESTINATION\". Add \"-u\" to update an existed Higress instance." && exit 1
    fi
    if [ ! -w "$DESTINATION" ]; then
      echo "The target folder \"$DESTINATION\" is not writeable." && exit 1
    fi
  else
    if [ "$MODE" == "update" ]; then
      echo "The target folder \"$DESTINATION\" for update doesn't exist." && exit 1
    fi
    mkdir -p "$DESTINATION"
    if [ $? -ne 0 ]; then
      exit 1
    fi
  fi

  cd "$DESTINATION"
  DESTINATION=$(pwd -P)
  cd - > /dev/null
}

outputUsage() {
  echo "Usage: $(basename -- "$0") [DIR] [OPTIONS...]"
  echo 'Install Higress (standalone version) into the DIR ("./higress" by default).'
  echo '
 -c, --config-url=URL       URL of the config storage
                            Use Nacos with format: nacos://192.168.0.1:8848
                            Use local files with format: file:///opt/higress/conf
     --use-builtin-nacos    use the built-in Nacos service instead of
                            an external one
     --nacos-ns=NACOS-NAMESPACE
                            the ID of Nacos namespace to store configurations
                            default to "aigateway-system" if unspecified
     --nacos-username=NACOS-USERNAME
                            the username used to access Nacos
                            only needed if auth is enabled in Nacos
     --nacos-password=NACOS-PASSWORD
                            the password used to access Nacos
                            only needed if auth is enabled in Nacos
 -k, --data-enc-key=KEY     the key used to encrypt sensitive configurations
                            MUST contain 32 characters
                            A random key will be generated if unspecified
     --nacos-service-port=NACOS-SERVICE-PORT
                            the HTTP port used to access the built-in Nacos service
                            default to 8848 if unspecified
     --nacos-console-port=NACOS-CONSOLE-PORT
                            the HTTP port used to access the built-in Nacos console
                            default to 8888 if unspecified
     --gateway-http-port=GATEWAY-HTTP-PORT
                            the HTTP port to be listened by the gateway
                            default to 80 if unspecified
     --gateway-https-port=GATEWAY-HTTPS-PORT
                            the HTTPS port to be listened by the gateway
                            default to 443 if unspecified
     --gateway-metrics-port=GATEWAY-METRICS-PORT
                            the metrics port to be listened by the gateway
                            default to 15020 if unspecified
     --console-port=CONSOLE-PORT
                            the port used to visit Higress Console
                            default to 8080 if unspecified
     --local-console-image  use the local Higress Console image
                            "aigateway-console:0.0.1"
     --console-image=IMAGE  use a custom Higress Console image
                            e.g. aigateway-console:dev
     --console-tag=TAG      override the Higress Console image tag while
                            keeping the default repository
     --gateway-image=IMAGE  use a custom Higress Gateway image
                            e.g. local/gateway:dev
     --gateway-tag=TAG      override the Higress Gateway image tag while
                            keeping the default repository
     --controller-image=IMAGE
                            use a custom Higress Controller image
                            e.g. local/controller:dev
     --controller-tag=TAG   override the Higress Controller image tag while
                            keeping the default repository
     --pilot-image=IMAGE    use a custom Higress Pilot image
                            e.g. local/pilot:dev
     --pilot-tag=TAG        override the Higress Pilot image tag while
                            keeping the default repository
     --api-server-image=IMAGE
                            use a custom Higress API Server image
                            e.g. local/api-server:0.0.29
     --api-server-tag=TAG   override the Higress API Server image tag while
                            keeping the default repository
     --nacos-image=IMAGE    use a custom Nacos image
                            e.g. local/nacos-server:v3.0.1
     --nacos-tag=TAG        override the Nacos image tag while
                            keeping the default repository
     --grafana-image=IMAGE  use a custom Grafana image
                            e.g. local/grafana:9.3.6
     --grafana-tag=TAG      override the Grafana image tag while
                            keeping the default repository
     --prometheus-image=IMAGE
                            use a custom Prometheus image
                            e.g. local/prometheus:v2.40.7
     --prometheus-tag=TAG   override the Prometheus image tag while
                            keeping the default repository
     --promtail-image=IMAGE use a custom Promtail image
                            e.g. local/promtail:2.9.4
     --promtail-tag=TAG     override the Promtail image tag while
                            keeping the default repository
     --loki-image=IMAGE     use a custom Loki image
                            e.g. local/loki:2.9.4
     --loki-tag=TAG         override the Loki image tag while
                            keeping the default repository
 -u, --update               update an existed Higress instance.
                            no user configuration will be changed during update.
 -h, --help                 give this help list'
}

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS="$(uname|tr '[:upper:]' '[:lower:]')"
  case "$OS" in
    # Minimalist GNU for Windows
    mingw*|cygwin*) OS='windows';;
  esac
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  if [ $EUID -ne 0 ]; then
    sudo "${@}"
  else
    "${@}"
  fi
}

# verifySupported checks that the os/arch combination is supported for
# binary builds, as well whether or not necessary tools are present.
verifySupported() {
  local supported="darwin-amd64\nlinux-amd64\nwindows-amd64\ndarwin-arm64\nlinux-arm64\nwindows-arm64\n"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "${OS}-${ARCH} platform isn't supported at the moment."
    echo "Stay tuned for updates on https://github.com/alibaba/higress."
    exit 1
  fi

  if [ "${HAS_CURL}" != "true" ] && [ "${HAS_WGET}" != "true" ]; then
    echo "Either curl or wget is required"
    exit 1
  fi

  if [ "${HAS_DOCKER}" != "true" ]; then
    echo "Docker is required"
    exit 1
  fi
}

# checkDesiredVersion checks if the desired version is available.
checkDesiredVersion() {
  if [ -z "$VERSION" ]; then
    # Get tag from release URL
    local latest_release_url="https://github.com/higress-group/higress-standalone/releases"
    if [ "${HAS_CURL}" == "true" ]; then
      VERSION=$(curl -Ls $latest_release_url | grep 'href="/higress-group/higress-standalone/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | sed -E 's/.*\/higress-group\/higress-standalone\/releases\/tag\/(v[0-9\.]+)".*/\1/g' | head -1)
    elif [ "${HAS_WGET}" == "true" ]; then
      VERSION=$(wget $latest_release_url -O - 2>&1 | grep 'href="/higress-group/higress-standalone/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | sed -E 's/.*\/higress-group\/higress-standalone\/releases\/tag\/(v[0-9\.]+)".*/\1/g' | head -1)
    fi
  fi
}

# download downloads the latest package
download() {
  HIGRESS_DIST="higress_${VERSION}.tar.gz"
  DOWNLOAD_URL="https://github.com/higress-group/higress-standalone/archive/refs/tags/${VERSION}.tar.gz"
  HIGRESS_TMP_ROOT="$(mktemp -dt higress-installer-XXXXXX)"
  HIGRESS_TMP_FILE="$HIGRESS_TMP_ROOT/$HIGRESS_DIST"
  echo "Downloading $DOWNLOAD_URL..."
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "$DOWNLOAD_URL" > "$HIGRESS_TMP_FILE"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O - "$DOWNLOAD_URL" > "$HIGRESS_TMP_FILE"
  fi
}

setEnvVar() {
  local env_file="$1"
  local key="$2"
  local value="$3"
  local escaped_value="${value//&/\\&}"

  if [ ! -f "$env_file" ]; then
    return
  fi

  if grep -q "^${key}=" "$env_file"; then
    sed -i -E "s|^${key}=.*$|${key}='${escaped_value}'|" "$env_file"
  else
    printf "%s='%s'\n" "$key" "$value" >> "$env_file"
  fi
}

getComposeFile() {
  local compose_dir="$DESTINATION/compose"

  for compose_name in docker-compose.yml docker-compose.yaml compose.yml compose.yaml; do
    if [ -f "$compose_dir/$compose_name" ]; then
      echo "$compose_dir/$compose_name"
      return 0
    fi
  done

  return 1
}

setComposeServiceImage() {
  local service="$1"
  local image="$2"
  local compose_file=""
  local tmp_file=""

  compose_file="$(getComposeFile 2>/dev/null || true)"
  if [ -z "$compose_file" ] || [ ! -f "$compose_file" ]; then
    return
  fi

  tmp_file="$(mktemp)"
  awk -v service="$service" -v image="$image" '
    BEGIN {
      service_header = "  " service ":"
      in_service = 0
    }
    {
      line = $0
      sub(/\r$/, "", line)
    }
    line == service_header {
      in_service = 1
      print
      next
    }
    in_service && line ~ /^  [^[:space:]].*:$/ {
      in_service = 0
    }
    in_service && line ~ /^    image:/ {
      $0 = "    image: " image
      in_service = 0
    }
    { print }
  ' "$compose_file" > "$tmp_file" && mv "$tmp_file" "$compose_file"
}

copyComposeArtifacts() {
  local compose_dir="$DESTINATION/compose"
  local source_compose=""

  source_compose="$(getComposeFile 2>/dev/null || true)"

  if [ -n "$source_compose" ]; then
    cp "$source_compose" "$DESTINATION/docker-compose.yml"
    echo "Generated docker-compose file: $DESTINATION/docker-compose.yml"
  fi

  if [ -f "$compose_dir/.env" ]; then
    cp "$compose_dir/.env" "$DESTINATION/.env"
  fi
}

applyCustomImageSettings() {
  local env_file="$DESTINATION/compose/.env"

  if [ -n "$CUSTOM_CONSOLE_IMAGE" ]; then
    setComposeServiceImage "console" "$CUSTOM_CONSOLE_IMAGE"
    echo "Using custom Higress Console image: $CUSTOM_CONSOLE_IMAGE"
  elif [ -n "$CUSTOM_CONSOLE_TAG" ]; then
    setEnvVar "$env_file" "HIGRESS_CONSOLE_TAG" "$CUSTOM_CONSOLE_TAG"
    echo "Using custom Higress Console image tag: $CUSTOM_CONSOLE_TAG"
  fi

  if [ -n "$CUSTOM_GATEWAY_IMAGE" ]; then
    setComposeServiceImage "gateway" "$CUSTOM_GATEWAY_IMAGE"
    echo "Using custom Higress Gateway image: $CUSTOM_GATEWAY_IMAGE"
  elif [ -n "$CUSTOM_GATEWAY_TAG" ]; then
    setEnvVar "$env_file" "HIGRESS_GATEWAY_TAG" "$CUSTOM_GATEWAY_TAG"
    echo "Using custom Higress Gateway image tag: $CUSTOM_GATEWAY_TAG"
  fi

  if [ -n "$CUSTOM_CONTROLLER_IMAGE" ]; then
    setComposeServiceImage "controller" "$CUSTOM_CONTROLLER_IMAGE"
    echo "Using custom Higress Controller image: $CUSTOM_CONTROLLER_IMAGE"
  elif [ -n "$CUSTOM_CONTROLLER_TAG" ]; then
    setEnvVar "$env_file" "HIGRESS_CONTROLLER_TAG" "$CUSTOM_CONTROLLER_TAG"
    echo "Using custom Higress Controller image tag: $CUSTOM_CONTROLLER_TAG"
  fi

  if [ -n "$CUSTOM_PILOT_IMAGE" ]; then
    setComposeServiceImage "pilot" "$CUSTOM_PILOT_IMAGE"
    echo "Using custom Higress Pilot image: $CUSTOM_PILOT_IMAGE"
  elif [ -n "$CUSTOM_PILOT_TAG" ]; then
    setEnvVar "$env_file" "HIGRESS_PILOT_TAG" "$CUSTOM_PILOT_TAG"
    echo "Using custom Higress Pilot image tag: $CUSTOM_PILOT_TAG"
  fi

  if [ -n "$CUSTOM_API_SERVER_IMAGE" ]; then
    setComposeServiceImage "apiserver" "$CUSTOM_API_SERVER_IMAGE"
    echo "Using custom Higress API Server image: $CUSTOM_API_SERVER_IMAGE"
  elif [ -n "$CUSTOM_API_SERVER_TAG" ]; then
    setEnvVar "$env_file" "HIGRESS_API_SERVER_TAG" "$CUSTOM_API_SERVER_TAG"
    echo "Using custom Higress API Server image tag: $CUSTOM_API_SERVER_TAG"
  fi

  if [ -n "$CUSTOM_NACOS_IMAGE" ]; then
    setComposeServiceImage "nacos" "$CUSTOM_NACOS_IMAGE"
    echo "Using custom Nacos image: $CUSTOM_NACOS_IMAGE"
  elif [ -n "$CUSTOM_NACOS_TAG" ]; then
    setEnvVar "$env_file" "NACOS_SERVER_TAG" "$CUSTOM_NACOS_TAG"
    echo "Using custom Nacos image tag: $CUSTOM_NACOS_TAG"
  fi

  if [ -n "$CUSTOM_GRAFANA_IMAGE" ]; then
    setComposeServiceImage "grafana" "$CUSTOM_GRAFANA_IMAGE"
    echo "Using custom Grafana image: $CUSTOM_GRAFANA_IMAGE"
  elif [ -n "$CUSTOM_GRAFANA_TAG" ]; then
    setEnvVar "$env_file" "GRAFANA_TAG" "$CUSTOM_GRAFANA_TAG"
    echo "Using custom Grafana image tag: $CUSTOM_GRAFANA_TAG"
  fi

  if [ -n "$CUSTOM_PROMETHEUS_IMAGE" ]; then
    setComposeServiceImage "prometheus" "$CUSTOM_PROMETHEUS_IMAGE"
    echo "Using custom Prometheus image: $CUSTOM_PROMETHEUS_IMAGE"
  elif [ -n "$CUSTOM_PROMETHEUS_TAG" ]; then
    setEnvVar "$env_file" "PROMETHEUS_TAG" "$CUSTOM_PROMETHEUS_TAG"
    echo "Using custom Prometheus image tag: $CUSTOM_PROMETHEUS_TAG"
  fi

  if [ -n "$CUSTOM_PROMTAIL_IMAGE" ]; then
    setComposeServiceImage "promtail" "$CUSTOM_PROMTAIL_IMAGE"
    echo "Using custom Promtail image: $CUSTOM_PROMTAIL_IMAGE"
  elif [ -n "$CUSTOM_PROMTAIL_TAG" ]; then
    setEnvVar "$env_file" "PROMTAIL_TAG" "$CUSTOM_PROMTAIL_TAG"
    echo "Using custom Promtail image tag: $CUSTOM_PROMTAIL_TAG"
  fi

  if [ -n "$CUSTOM_LOKI_IMAGE" ]; then
    setComposeServiceImage "loki" "$CUSTOM_LOKI_IMAGE"
    echo "Using custom Loki image: $CUSTOM_LOKI_IMAGE"
  elif [ -n "$CUSTOM_LOKI_TAG" ]; then
    setEnvVar "$env_file" "LOKI_TAG" "$CUSTOM_LOKI_TAG"
    echo "Using custom Loki image tag: $CUSTOM_LOKI_TAG"
  fi
}

postConfigure() {
  applyCustomImageSettings
  copyComposeArtifacts
}

# install installs the product.
install() {
  tar -zx --exclude=".github" --exclude="all-in-one" --exclude="docs" --exclude="src" --exclude="test" --exclude="CODEOWNERS" -f "$HIGRESS_TMP_FILE" -C "$DESTINATION" --strip-components=1
  echo -n "$VERSION" > "$DESTINATION/VERSION"
  bash "$DESTINATION/bin/configure.sh" ${CONFIG_ARGS[@]}
  postConfigure
}

# update updates the product.
update() {
  CURRENT_VERSION="0.0.0"
  if [ -f "$DESTINATION/VERSION" ]; then
    CURRENT_VERSION="$(cat "$DESTINATION/VERSION")"
  fi
  if [ "$CURRENT_VERSION" == "$VERSION" ]; then
    if [ -n "$CUSTOM_CONSOLE_IMAGE" ] || [ -n "$CUSTOM_CONSOLE_TAG" ] || \
       [ -n "$CUSTOM_GATEWAY_IMAGE" ] || [ -n "$CUSTOM_GATEWAY_TAG" ] || \
       [ -n "$CUSTOM_CONTROLLER_IMAGE" ] || [ -n "$CUSTOM_CONTROLLER_TAG" ] || \
       [ -n "$CUSTOM_PILOT_IMAGE" ] || [ -n "$CUSTOM_PILOT_TAG" ] || \
       [ -n "$CUSTOM_API_SERVER_IMAGE" ] || [ -n "$CUSTOM_API_SERVER_TAG" ] || \
       [ -n "$CUSTOM_NACOS_IMAGE" ] || [ -n "$CUSTOM_NACOS_TAG" ] || \
       [ -n "$CUSTOM_GRAFANA_IMAGE" ] || [ -n "$CUSTOM_GRAFANA_TAG" ] || \
       [ -n "$CUSTOM_PROMETHEUS_IMAGE" ] || [ -n "$CUSTOM_PROMETHEUS_TAG" ] || \
       [ -n "$CUSTOM_PROMTAIL_IMAGE" ] || [ -n "$CUSTOM_PROMTAIL_TAG" ] || \
       [ -n "$CUSTOM_LOKI_IMAGE" ] || [ -n "$CUSTOM_LOKI_TAG" ] || \
       [ ! -f "$DESTINATION/docker-compose.yml" ]; then
      postConfigure
      echo "Higress image settings are updated."
      exit 0
    fi
    echo "Higress is already up-to-date."
    exit 0
  fi

  BACKUP_FOLDER="$(cd ${DESTINATION}/.. ; pwd)"
  BACKUP_FILE="${BACKUP_FOLDER}/higress_backup_$(date '+%Y%m%d%H%M%S').tar.gz" 
  tar -zc -f "$BACKUP_FILE" -C "$DESTINATION" .
  echo "The current version is packed here: $BACKUP_FILE"
  echo ""

  download
  echo ""

  tar -zx --exclude=".github" --exclude="all-in-one" --exclude="docs" --exclude="src" --exclude="test" --exclude="CODEOWNERS" --exclude="compose/.env" -f "$HIGRESS_TMP_FILE" -C "$DESTINATION" --strip-components=1
  tar -zx -f "$HIGRESS_TMP_FILE" -C "$DESTINATION" --transform='s/env/env_new/g' --strip-components=1 "higress-standalone-${VERSION#v}/compose/.env"
  bash "$DESTINATION/bin/update.sh"
  echo -n "$VERSION" > "$DESTINATION/VERSION"
  postConfigure
  return
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [ -n "$INPUT_ARGUMENTS" ]; then
      echo "Failed to ${MODE} Higress with the arguments provided: $INPUT_ARGUMENTS"
    else
      echo "Failed to ${MODE} Higress"
    fi
    echo -e "\tFor support, go to https://github.com/alibaba/higress."
  fi
  exit $result
}

# cleanup temporary files.
cleanup() {
  if [[ -d "${HIGRESS_TMP_ROOT:-}" ]]; then
    rm -rf "$HIGRESS_TMP_ROOT"
  fi
}

parseArgs "$@"
validateArgs

# Stop execution on any error
trap "fail_trap" EXIT
set -e

initArch
initOS
verifySupported

checkDesiredVersion
case "$MODE" in
  update)
    update
    ;;
  *)
    download
    install
    ;;
esac
cleanup
