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

from .interconnection.runtime import ecdh_psi_pb2

from .base import lctx_send_proto

import queue
from threading import Event
from tqdm import tqdm


class CipherStore():
    def __init__(self, use_cache: bool = True) -> None:
        self.peer_cipher = queue.Queue()

        if use_cache:
            import rocksdict
            self.peer_cipher_set = rocksdict.Rdict(f"./peer_cipher_set_{hash(self)}")
            setattr(self, "insert",
                    lambda cipher: self.peer_cipher_set.put(cipher, b''))
            setattr(self, "contains",
                    lambda cipher: self.peer_cipher_set.get(cipher) is not None)
        else:
            import crypto
            self.peer_cipher_set = crypto.BytesHashSet()
            setattr(self, "insert",
                    lambda cipher: self.peer_cipher_set.insert(cipher))
            setattr(self, "contains",
                    lambda cipher: self.peer_cipher_set.contains(cipher))

        self.peer_cipher_set_done = Event()

        self.local_index_record = 0
        self.local_take_index = []

    def calcu_add_peer_cipher(self, array, ctx, done):
        array = ctx.curve.diffie_hellman(array)

        if done:
            self.peer_cipher.put(None)
        self.peer_cipher.put(array)

        if ctx.need_recv_cipher:
            for cipher in array.to_pylist():
                self.insert(cipher)
            if done:
                self.peer_cipher_set_done.set()

    def send_dualenc(self, ctx, lctx):
        batch_index = 0
        is_last_batch = False
        bar = tqdm(desc='send_dualenc', total=100)
        while not is_last_batch:
            arr = self.peer_cipher.get()

            if arr is None:
                arr = self.peer_cipher.get()
                is_last_batch = True

            ciphertext = ctx.point_octet_marshal(arr)
            count = len(arr)

            protomsg = ecdh_psi_pb2.EcdhPsiCipherBatch(
                type="dual.enc",
                batch_index=batch_index,
                is_last_batch=is_last_batch,
                count=count,
                ciphertext=ciphertext
            )

            lctx_send_proto(lctx, protomsg)

            batch_index += 1
            bar.update(1)

    def recv_duaenc_local_cipher(self, array, ctx, done):
        self.peer_cipher_set_done.wait()
        for cipher in array.to_pylist():
            if self.contains(cipher):
                self.local_take_index.append(self.local_index_record)

            self.local_index_record += 1
