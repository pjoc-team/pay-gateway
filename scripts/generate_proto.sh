#!/usr/bin/env bash
#set -x
set -e
language="${1:-go}"
dir="${2}"
output_dir="${3}"

echo "search proto file in dir: ${2} and output to: ${3}"
mkdir -p "${output_dir}"

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/../"
PROTO_HOME="${WORK_HOME}/${dir}"
#IMPORT_HOME="$(go env GOPATH)/src"

echo "dirname $WORK_HOME"
#echo "IMPORT_HOME: $(cd "$IMPORT_HOME" && pwd)"
echo "WORK_HOME: $(cd "$WORK_HOME" && pwd)"
echo "PROTO_HOME: $(cd "$PROTO_HOME" && pwd)"

# swagger bindata generator
# see https://segmentfault.com/a/1190000013513469
# TODO move to docker
#go get -u github.com/go-bindata/go-bindata/...
#go get -u github.com/elazarl/go-bindata-assetfs/...
#mkdir -p swagger

find $PROTO_HOME -name "*.proto" | while read proto; do
  dir="$(dirname "$proto")"
  dir="$(cd "$dir" && pwd)"
  if [ -z "${output_dir}" ]; then
    out_dir="$dir/$base_dir"
  else
    out_dir="`pwd`/${output_dir}"
  fi
  # parse file name without directory and suffix
  # parse "./proto/adn.proto" to "adn"
  file_name="${proto##*/}"
  proto_name="${file_name%.*}"
  echo "proto file: $proto dir: ${dir} file_name: $file_name proto_name: $proto_name out_dir: ${out_dir}"
  echo "generating proto..."
  [[ "$language" == "go" ]] && addition=" --with-gateway --validate-out lang=go:/out "
#  docker run --rm -v "$dir":/defs -v "${dir}":/out blademainer/protoc-all:latest -i /defs -i /go/src -d /defs/ -l $language -o /out --lint $addition
  docker run --rm -v "$dir":/defs -v "${dir}":/out namely/protoc-all:latest -f $file_name -l $language -o /out --lint --with-validator $addition
#  docker run --rm -v $dir:/defs -v ${IMPORT_HOME}:/input blademainer/protoc-all:latest -i /defs -i /input -i /go/src/ -d /defs/ -l go -o /defs --validate-out "lang=go:/defs" --with-gateway --lint $addition
#  docker run --rm -v "$dir":/defs -v "${out_dir}":/out namely/protoc-all:latest -f ${file_name} -i ${dir} -l $language -o /out --lint --with-validator --validate-out --with-gateway
  addition=""
done

# generate js protos
find $WORK_HOME -name "generate.sh" | while read script; do
  sh $script
done