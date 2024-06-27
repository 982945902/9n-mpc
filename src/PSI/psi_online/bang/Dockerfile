# Copyright 2020 The 9nFL Authors. All Rights Reserved.
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

FROM buildpack-deps:24.04

RUN set -ex && \
    sed -i 's/archive.ubuntu.com/mirrors.jd.com/g' /etc/apt/sources.list && \
    sed -i 's/security.ubuntu.com/mirrors.jd.com/g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get upgrade -y --no-install-recommends && \
    apt-get install -y --no-install-recommends \
        golang && \
    rm -rf /var/lib/apt/lists/*


RUN set -ex && \
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs >rustup.sh && \
    chmod +x rustup.sh && \
    ./rustup.sh -y && \
    cp ~/.cargo/bin/* /usr/local/bin/

ENV LIBCLANG_PATH=/usr/lib/llvm-14/lib/

COPY . /App/bang
WORKDIR /App/bang
RUN go mod tidy && go build

WORKDIR /App/

ENTRYPOINT [ "sleep", "3600d" ]
