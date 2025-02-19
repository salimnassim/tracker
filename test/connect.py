import socket
import struct
import random
import logging

def send_connect_request(tracker):
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(5)

    tracker_ip, tracker_port = tracker.split(":")
    tracker_addr = (tracker_ip, int(tracker_port))

    protocol_id = 0x41727101980
    action = 0
    transaction_id = random.randint(0, 0xFFFFFFFF)

    packet = struct.pack(">QII", protocol_id, action, transaction_id)

    sock.sendto(packet, tracker_addr)

    response, _ = sock.recvfrom(16)

    resp_action, resp_transaction_id, connection_id = struct.unpack(">IIQ", response)

    logging.debug(f"action={resp_action}, transaction_id={resp_transaction_id}, connection_id={connection_id}")

send_connect_request("127.0.0.1:8888")