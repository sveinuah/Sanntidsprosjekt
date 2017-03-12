#!/bin/python
import socket
host="localhost"
port=20014
udp_sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
udp_sock.sendto('TestPacket: Please Ignore', (host, port))