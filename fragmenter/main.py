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
    enc_fragment_size = target.recv(int(os.getenv("ENC_FRAGMENT_SIZE_LENGTH")))
    fragment_size = decode_fragment_desc_size(enc_fragment_size)
    enc_fragment = target.recv(fragment_size)
    return json.loads(enc_fragment)

def encode_fragment_desc(content: dict) -> bytes:
    return encode_bytes(json.dumps(content).encode("utf-8"))

def from_sidecar() -> str:
    received_one = False
    while not received_one:
        try:
            sidecar_input = socket(AF_UNIX) 
            sidecar_input.connect(os.getenv("FRAGMENTER_INPUT"))
            while not received_one:
                # Decode messages send over this socket one-by-one
                fragment = decode_fragment_desc(sidecar_input)
                received_one = True
                yield fragment
        except Exception as err:
            print(f"Exception: {err}, retrying from_sidecar connection in 5 seconds...", flush=True)
            time.sleep(5)
        finally: 
            # Clean up before trying again. Letting the other side know we quit this connection
            sidecar_input.shutdown(SHUT_RDWR)
            sidecar_input.close()


def to_sidecar() -> None:
    msg_to_send:dict = None
    had_connection = False
    while not had_connection:
        try:
            to_sidecar_socket = socket(AF_UNIX) 
            to_sidecar_socket.connect(os.getenv("FRAGMENTER_OUTPUT"))
            while True:
                # If no received message, else we received one, but sending failed
                if msg_to_send == None: 
                    msg_to_send:dict = (yield)
                encoding = encode_fragment_desc(msg_to_send)
                to_sidecar_socket.send(encoding)
                msg_to_send = None
                had_connection = True
                yield
        except Exception as err:
            if had_connection:
                print(f"Error: '{err}'", flush=True)
                exit(os.EX_SOFTWARE)
            else:
                print(f"Exception: '{err}', creating connection failed, retrying in 5 seconds...", flush=True)
                time.sleep(5)
        finally:
            to_sidecar_socket.shutdown(SHUT_RDWR)
            to_sidecar_socket.close()


if __name__ == "__main__":
    towards_sidecar = to_sidecar()
    next(towards_sidecar)
    files = next(from_sidecar())
    for f in files:
        print(f"sending: {[f]}", flush=True)
        towards_sidecar.send([f])        
        next(towards_sidecar)
    
    exit(0)
