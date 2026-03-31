#!/usr/bin/env bash
set -euo pipefail

# 简单开发脚本：
# 1) 远程/容器内构建 manylinux wheel 并缓存在 .whl 目录下
# 2) 使用缓存的 wheel 做一次全量构建后，再进行快速的 pip install -e python/
#
# 用法示例：
#   ./dev_build.sh wheel 3.11            # 构建 Python 3.11 的 manylinux wheel 到 .whl/
#   ./dev_build.sh editable 3.11         # 先构建/复用 wheel，然后做一次带缓存的 pip install -e python/

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_MINOR="${2:-3.11}"
MODE="${1:-editable}"  # 默认 editable

WHL_DIR="${ROOT_DIR}/.whl"

function usage() {
  echo "用法："
  echo "  $0 wheel    <3.10|3.11|3.12>   # 仅构建 manylinux wheel"
  echo "  $0 editable <3.10|3.11|3.12>   # 构建/复用 wheel 后，再 pip install -e python/"
  exit 1
}

if [[ "${MODE}" != "wheel" && "${MODE}" != "editable" ]]; then
  usage
fi

if [[ -z "${PYTHON_MINOR}" ]]; then
  usage
fi

PY_TAG="python${PYTHON_MINOR}"

function build_wheel_manylinux() {
  echo "===> 使用 build-manylinux-wheel.sh 构建 manylinux wheel (PY=${PY_TAG})"
  mkdir -p "${WHL_DIR}"

  # 这里假设你是在支持 manylinux 构建的环境/容器里运行，
  # 并且 ./ci/build/build-manylinux-wheel.sh 可直接调用。
  (
    cd "${ROOT_DIR}"
    BUILDKITE_COMMIT="${BUILDKITE_COMMIT:-dev}" \
      ./ci/build/build-manylinux-wheel.sh "${PY_TAG}"
  )

  echo "===> 当前缓存的 wheel 文件："
  ls -1 "${WHL_DIR}" || true
}

function fast_editable_install() {
  echo "===> 确保已有缓存 wheel（如有需要会先构建一次）"

  # 如果没有任何 ray*.whl，就先构建一次
  if ! ls "${WHL_DIR}"/ray-*.whl >/dev/null 2>&1; then
    build_wheel_manylinux
  fi

  echo "===> 使用缓存 wheel 安装一次普通 ray（非 editable）以准备二进制和第三方依赖"
  # 这里默认使用当前虚拟环境里的 python/pip
  python -m pip install --upgrade pip
  python -m pip install "${WHL_DIR}"/ray-*.whl

  echo "===> 使用缓存的构建结果做快速 editable 安装"
  (
    cd "${ROOT_DIR}"
    export SKIP_BAZEL_BUILD=1
    export SKIP_THIRDPARTY_INSTALL_CONDA_FORGE=1
    python -m pip install -e python/
  )

  echo "===> 完成：已基于缓存构建结果做 pip install -e python/"
}

case "${MODE}" in
  wheel)
    build_wheel_manylinux
    ;;
  editable)
    fast_editable_install
    ;;
esac

