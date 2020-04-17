import time
import sys
from socket import socket, AF_UNIX, SHUT_RDWR
import struct
import os
import json

def encode_fragment_desc_size(size: int) -> bytes:
    return struct.pack("<I", size)

def decode_fragment_desc_size(size_bytes: bytes) -> int:
    return struct.unpack("<I", size_bytes)[0]

def encode_bytes(content: bytes) -> bytes:
    size = len(content)
    return encode_fragment_desc_size(size) + content

def decode_fragment_desc(target: socket) -> dict:
    enc_fragment_size = target.recv(ENC_FRAGMENT_SIZE_LENGTH)
    fragment_size = decode_fragment_desc_size(enc_fragment_size)
    enc_fragment = target.recv(fragment_size)
    return json.loads(enc_fragment)

def encode_fragment_desc(content: dict) -> bytes:
    return encode_bytes(json.dumps(content).encode("utf-8"))

def from_sidecar(socket_name:str) -> str:
    while True:
        try:
            sidecar_cat_input = socket(AF_UNIX) 
            sidecar_cat_input.connect(socket_name)
            while True:
                # Decode messages sennd over this socket one-by-one
                fragment = decode_fragment_desc(sidecar_cat_input)
                yield fragment
        except Exception as err:
            print(f"Exception: {err}, retrying from_sidecar connection in 5 seconds...", flush=True)
            time.sleep(5)
        finally: 
            # Clean up before trying again. Letting the other side know we quit this connection
            sidecar_cat_input.shutdown(SHUT_RDWR)
            sidecar_cat_input.close()


def to_sidecar(socket_name:str) -> None:
    msg_to_send:dict = None
    while True:
        try:
            to_sidecar_socket = socket(AF_UNIX) 
            to_sidecar_socket.connect(socket_name)
            while True:
                # If no received message, else we received one, but sending failed
                if msg_to_send == None: 
                    msg_to_send:dict = (yield)
                print(f"sending: {msg_to_send}", flush=True)
                encoding = encode_fragment_desc(msg_to_send)
                to_sidecar_socket.send(encoding)
                msg_to_send = None
                yield
        except Exception as err:
            print(f"Exception: {err}, retrying to_sidecar connection in 5 seconds...", flush=True)
            time.sleep(5)
            to_sidecar_socket.shutdown(SHUT_RDWR)
            to_sidecar_socket.close()


if __name__ == "__main__":
    ENC_FRAGMENT_SIZE_LENGTH = 4
    INPUT = "./build/tf.sock"
    OUTPUT = "./build/ff.sock"
    towards_sidecar = to_sidecar(OUTPUT)
    next(towards_sidecar)
    for files in from_sidecar(INPUT):
        for f in files:
            towards_sidecar.send([f])        
            next(towards_sidecar)
