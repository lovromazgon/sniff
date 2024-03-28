# Sniff

A very simple reverse proxy that forwards TCP traffic and dumps requests and
responses to stdout for inspection.

Build the binary using `make build`.

Help:

```sh
$ ./sniff -h
Usage of ./sniff:
  -in string
    	The address to listen on for incoming requests. (default ":9093")
  -out string
    	The address to forward requests to. (default "localhost:9092")
```

Example output of a Kafka client requesting metadata:

```sh
./sniff -in ":9093" -out ":9092"
[::1]:52143 [2024-03-28 16:25:01] REQ:
00000000  00 00 00 1d 00 03 00 05  00 00 00 00 00 0e 6b 61  |..............ka|
00000010  66 6b 61 63 74 6c 2d 6c  6f 76 72 6f ff ff ff ff  |fkactl-lovro....|
00000020  00                                                |.|

[::1]:52143 [2024-03-28 16:25:01] DIAL: Dialing real server ...
[::1]:52143 [2024-03-28 16:25:01] WRITE: Writing to real server ...
[::1]:52143 [2024-03-28 16:25:01] RESPONSE:
00000000  00 00 00 4c 00 00 00 00  00 00 00 00 00 00 00 01  |...L............|
00000010  00 00 00 00 00 00 00 00  23 87 ff ff ff ff 00 00  |........#.......|
00000020  00 00 00 00 00 01 00 00  00 03 66 6f 6f 00 00 00  |..........foo...|
00000030  00 01 00 00 00 00 00 00  00 00 00 00 00 00 00 01  |................|
00000040  00 00 00 00 00 00 00 01  00 00 00 00 00 00 00 00  |................|
```
