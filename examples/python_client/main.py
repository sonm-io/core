# SONM Node <-> Python connection example.
#
# 1. Install python requirements: grpcio grpcio-tools protobuf
#
# 2. generate proto for python using the following command:
# "python3 -m grpc_tools.protoc -I../../proto/ --python_out=./proto--grpc_python_out=./proto/ ../../proto/*.proto"
#
# 3. run the following code

import grpc
from grpc._cython.cygrpc import CompressionAlgorithm, CompressionLevel

from proto import node_pb2_grpc as node_rpc
from proto import node_pb2 as node_pb


def main():
    chan_ops = [
        ('grpc.default_compression_algorithm', CompressionAlgorithm.gzip),
        ('grpc.grpc.default_compression_level', CompressionLevel.high),
    ]

    chan = grpc.insecure_channel('unix:/tmp/sonm_node.sock', chan_ops)

    stub = node_rpc.DealManagementStub(chan)
    req = node_pb.DealListRequest(owner='0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD')

    reply = stub.List(req)
    print("reply: {}".format(reply))


if __name__ == '__main__':
    main()
