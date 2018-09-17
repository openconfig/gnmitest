#!/bin/sh

# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

proto_imports=".:${GOPATH}/src/github.com/google/protobuf/src:${GOPATH}/src"

# Go
protoc -I=$proto_imports --go_out=plugins=grpc:. proto/gnmitest/gnmitest.proto
protoc -I=$proto_imports --go_out=plugins=grpc:. proto/suite/suite.proto
protoc -I=$proto_imports --go_out=plugins=grpc:. proto/tests/tests.proto
protoc -I=$proto_imports --go_out=plugins=grpc:. proto/report/report.proto

# Python
python -m grpc_tools.protoc -I=$proto_imports --python_out=. --grpc_python_out=. proto/gnmitest/gnmitest.proto
python -m grpc_tools.protoc -I=$proto_imports --python_out=. --grpc_python_out=. proto/suite/suite.proto
python -m grpc_tools.protoc -I=$proto_imports --python_out=. --grpc_python_out=. proto/tests/tests.proto
python -m grpc_tools.protoc -I=$proto_imports --python_out=. --grpc_python_out=. proto/report/report.proto
